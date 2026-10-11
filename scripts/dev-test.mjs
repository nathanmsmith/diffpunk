import assert from 'node:assert/strict';
import { spawn, execFileSync } from 'node:child_process';
import {
  cp,
  mkdir,
  mkdtemp,
  readFile,
  realpath,
  rm,
  symlink,
  writeFile,
} from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { once } from 'node:events';
import { chromium, firefox } from 'playwright';

const workspace = await mkdtemp(join(tmpdir(), 'diff-punk-dev-'));
const app = join(workspace, 'app');
const repository = join(workspace, 'repo');
const binary = join(app, 'bin/diff-punk');
const devURL = 'http://127.0.0.1:5173/__diff_punk_dev/';
let vite, cli, browser;
let viteOutput = '';

async function stop(child) {
  if (!child || child.exitCode !== null || child.signalCode !== null) return;
  const exited = once(child, 'exit');
  child.kill('SIGTERM');
  const timeout = setTimeout(() => child.kill('SIGKILL'), 5000);
  try {
    await exited;
  } finally {
    clearTimeout(timeout);
  }
}

async function startCLI(executable = binary, args = []) {
  cli = spawn(executable, args, {
    cwd: repository,
    env: { ...process.env, DIFF_PUNK_NO_OPEN: '1' },
    stdio: ['ignore', 'ignore', 'pipe'],
  });
  return new Promise((resolveURL, reject) => {
    let output = '';
    const timer = setTimeout(
      () => reject(new Error(`CLI startup timed out: ${output}`)),
      15000,
    );
    cli.once('error', reject);
    cli.once('exit', (code) => {
      clearTimeout(timer);
      reject(new Error(`CLI exited ${code}: ${output}`));
    });
    cli.stderr.on('data', (chunk) => {
      output += chunk;
      const match = output.match(/http:\/\/127\.0\.0\.1:5555\/[a-f0-9]+\//);
      if (match) {
        clearTimeout(timer);
        resolveURL({ url: match[0], output });
      }
    });
  });
}

try {
  // Edit an isolated frontend copy so this test never touches the working tree.
  await mkdir(app);
  await mkdir(join(app, 'bin'));
  await cp('bin/diff-punk', binary);
  await mkdir(repository);
  await cp('frontend', join(app, 'frontend'), { recursive: true });
  await cp('package.json', join(app, 'package.json'));
  await cp('vite.config.mjs', join(app, 'vite.config.mjs'));
  await symlink(resolve('node_modules'), join(app, 'node_modules'), 'dir');
  const git = (...args) =>
    execFileSync('git', args, { cwd: repository, encoding: 'utf8' });
  git('init', '-q');
  git('config', 'user.name', 'Test');
  git('config', 'user.email', 'test@example.invalid');
  git('config', 'commit.gpgsign', 'false');
  const original =
    Array.from({ length: 30 }, (_, i) => `const value${i} = ${i};`).join('\n') +
    '\n';
  await writeFile(join(repository, 'foo.ts'), original);
  git('add', '.');
  git('commit', '-qm', 'initial');
  await writeFile(
    join(repository, 'foo.ts'),
    original.replace('value14 = 14', 'value14 = 140'),
  );
  await writeFile(join(repository, 'new.ts'), 'const fresh = true;\n');

  vite = spawn(process.execPath, [resolve('node_modules/vite/bin/vite.js')], {
    cwd: app,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  vite.stdout.on('data', (chunk) => {
    viteOutput += chunk;
  });
  vite.stderr.on('data', (chunk) => {
    viteOutput += chunk;
  });
  const deadline = Date.now() + 15000;
  for (;;) {
    if (vite.exitCode !== null || Date.now() > deadline) {
      throw new Error(`Vite did not start: ${viteOutput}`);
    }
    try {
      const response = await fetch(`${devURL}status`, {
        signal: AbortSignal.timeout(500),
      });
      if (response.ok && (await response.text()) === 'diff-punk-vite-v1') break;
    } catch {
      /* Wait until Vite is listening. */
    }
    await new Promise((resolveWait) => setTimeout(resolveWait, 50));
  }

  const { url, output } = await startCLI(binary, []);
  assert.match(output, /React dev server detected/);
  const snapshot = await (await fetch(`${url}snapshot.json`)).json();
  assert.equal(snapshot.repository, await realpath(repository));
  browser = await (
    process.env.DIFF_PUNK_TEST_BROWSER === 'firefox' ? firefox : chromium
  ).launch({ headless: true });
  const page = await browser.newPage({
    viewport: { width: 1440, height: 1000 },
  });
  const errors = [];
  const sockets = [];
  page.on('pageerror', (error) => errors.push(error.message));
  page.on('websocket', (socket) => sockets.push(socket.url()));
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text());
  });
  await page.goto(url);
  const section = page.locator('section[data-path="foo.ts"]');
  await section.locator('pre').filter({ hasText: 'value14 = 140' }).waitFor();
  await section.getByRole('button', { name: 'Show full file' }).click();
  await page.locator('#unified').click();
  await page.evaluate(() => {
    window.devTestSentinel = 'preserved';
  });

  // An App edit reruns effects; an individual file edit must retain expansion.
  const appFile = join(app, 'frontend/src/App.tsx');
  await writeFile(
    appFile,
    (await readFile(appFile, 'utf8')).replace(
      'className="workspace"',
      'className="workspace refreshed-app"',
    ),
  );
  await page.locator('.refreshed-app').waitFor();
  const sectionFile = join(app, 'frontend/src/FileSection.tsx');
  await writeFile(
    sectionFile,
    (await readFile(sectionFile, 'utf8')).replace(
      'className="file-header"',
      'className="file-header refreshed-file"',
    ),
  );
  await section.locator('.refreshed-file').waitFor();
  assert.equal(
    await page.locator('#unified').getAttribute('aria-pressed'),
    'true',
  );
  assert.equal(
    await section
      .getByRole('button', { name: 'Show changes only' })
      .getAttribute('aria-expanded'),
    'true',
  );
  assert.match(await section.locator('pre').innerText(), /value0 = 0/);
  assert.equal(
    await section.locator('pre').getAttribute('data-diff-type'),
    'single',
  );

  const cssFile = join(app, 'frontend/src/style.css');
  await writeFile(
    cssFile,
    (await readFile(cssFile, 'utf8')) +
      '\nbody { --dev-refresh-test: applied; }\n',
  );
  await page.waitForFunction(
    () =>
      getComputedStyle(document.body)
        .getPropertyValue('--dev-refresh-test')
        .trim() === 'applied',
  );
  assert.equal(await page.evaluate(() => window.devTestSentinel), 'preserved');
  assert.ok(sockets.length > 0, 'No hot-reload WebSocket connected');
  assert.ok(
    sockets.every((socket) => socket.startsWith('ws://127.0.0.1:5555/')),
    `WebSocket bypassed Go: ${sockets}`,
  );
  assert.deepEqual(errors, []);
  await writeFile(join(repository, 'foo.ts'), 'changed after startup\n');
  assert.deepEqual(await (await fetch(`${url}snapshot.json`)).json(), snapshot);
  await browser.close();
  browser = undefined;
  await stop(cli);

  // Even a release called "dev" must ignore a running Vite server.
  const releaseBinary = join(workspace, 'release-diff-punk');
  execFileSync('go', [
    'build',
    '-ldflags=-X main.development=false',
    '-o',
    releaseBinary,
    './cli',
  ]);
  const release = await startCLI(releaseBinary);
  assert.doesNotMatch(release.output, /React dev server detected/);
  assert.doesNotMatch(
    await (await fetch(release.url)).text(),
    /@vite|@react-refresh|__diff_punk_dev/,
  );
  assert.equal(
    (await fetch(new URL('/__diff_punk_dev/status', release.url))).status,
    404,
  );
  await stop(cli);

  const attached = await startCLI();
  assert.match(attached.output, /React dev server detected/);
  await stop(vite);
  const unavailable = await fetch(attached.url);
  assert.equal(unavailable.status, 502);
  assert.match(await unavailable.text(), /npm run dev/);
  assert.equal((await fetch(`${attached.url}snapshot.json`)).status, 200);
  await stop(cli);
  const automatic = await startCLI();
  assert.match(automatic.output, /starting React dev server/);
  assert.match(automatic.output, /React dev server detected/);
  assert.match(await (await fetch(automatic.url)).text(), /@vite\/client/);
  await stop(cli);
  await assert.rejects(fetch(`${devURL}status`));
  const standalone = join(workspace, 'standalone-diff-punk');
  await cp(binary, standalone);
  const fallback = await startCLI(standalone);
  assert.doesNotMatch(fallback.output, /React dev server detected/);
  assert.match(await (await fetch(fallback.url)).text(), /assets\/.*\.js/);
  console.log(
    'Dev checks passed: automatic startup and shutdown, reuse, real frozen diffs, React/CSS refresh, preserved state, WebSocket proxy, release isolation, and embedded fallback.',
  );
} catch (error) {
  console.error(viteOutput);
  throw error;
} finally {
  await browser?.close();
  await stop(cli);
  await stop(vite);
  await rm(workspace, { recursive: true, force: true });
}
