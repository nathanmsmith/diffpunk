import assert from 'node:assert/strict';
import { spawn, execFileSync } from 'node:child_process';
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';
import { once } from 'node:events';
import { chromium, firefox } from 'playwright';

const binary = resolve('bin/diff-punk');
const fixture = await mkdtemp(join(tmpdir(), 'diff-punk-browser-'));
const git = (...args) =>
  execFileSync('git', args, { cwd: fixture, encoding: 'utf8' });
const write = (name, body) => writeFile(join(fixture, name), body);
let server, browser;
try {
  git('init', '-q');
  git('config', 'user.name', 'Test');
  git('config', 'user.email', 'test@example.invalid');
  git('config', 'commit.gpgsign', 'false');
  await mkdir(join(fixture, 'src/deep'), { recursive: true });
  const original =
    [
      '// TOP_CONTEXT',
      ...Array.from({ length: 28 }, (_, i) => `const value${i} = ${i};`),
      '// END_CONTEXT',
    ].join('\n') + '\n';
  await write('foo.ts', original);
  await write('foo.test.ts', 'test("original", () => {});\n');
  await write('src/deep/bar.ts', 'const bar = 1;\n');
  await write('old.txt', 'unchanged rename\n');
  await write('deleted.txt', 'deleted\n');
  await write('binary.dat', Buffer.from([0, 1, 2]));
  git('add', '.');
  git('commit', '-qm', 'initial');
  await write(
    'foo.ts',
    original.replace('const value14 = 14;', 'const value14 = 140;'),
  );
  await write('foo.test.ts', 'test("changed", () => {});\n');
  await write('src/deep/bar.ts', 'const bar = 2;\n');
  await write('binary.dat', Buffer.from([0, 3, 4]));
  await rm(join(fixture, 'deleted.txt'));
  git('mv', 'old.txt', 'renamed.txt');
  await write('added.ts', 'const added = "<script>alert(1)</script>";\n');
  git('add', 'added.ts');
  server = spawn(binary, ['HEAD'], {
    cwd: fixture,
    env: { ...process.env, DIFF_PUNK_NO_OPEN: '1' },
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const url = await new Promise((resolveURL, reject) => {
    let output = '';
    const timeout = setTimeout(
      () => reject(new Error(`Server did not start: ${output}`)),
      15000,
    );
    server.stderr.on('data', (chunk) => {
      output += chunk;
      const match = output.match(/http:\/\/127\.0\.0\.1:\d+\/[a-f0-9]+\//);
      if (match) {
        clearTimeout(timeout);
        resolveURL(match[0]);
      }
    });
    server.once('exit', (code) => {
      clearTimeout(timeout);
      reject(new Error(`Server exited ${code}: ${output}`));
    });
  });
  const snapshot = await (await fetch(`${url}snapshot.json`)).json();
  const order = snapshot.files.map((f) => f.name);
  assert.equal(order.indexOf('foo.test.ts'), order.indexOf('foo.ts') + 1);
  assert.ok(order.indexOf('foo.test.ts') < order.indexOf('src/deep/bar.ts'));
  // Later edits must not leak into the browser's full-file view.
  await write('foo.ts', '// EDITED_AFTER_SNAPSHOT\n');
  browser = await (
    process.env.DIFF_PUNK_TEST_BROWSER === 'firefox' ? firefox : chromium
  ).launch({ headless: true });
  const page = await browser.newPage({
    viewport: { width: 1440, height: 1000 },
  });
  const errors = [];
  const external = [];
  page.on('pageerror', (err) => errors.push(err.message));
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(msg.text());
  });
  page.on('request', (req) => {
    if (!req.url().startsWith(url) && !req.url().startsWith('data:'))
      external.push(req.url());
  });
  await page.goto(url);
  await page.locator('.diff-host pre').first().waitFor();
  const foo = page.locator('section[data-path="foo.ts"]');
  await foo.locator('pre').filter({ hasText: 'value14 = 140' }).waitFor();
  assert.match(await foo.locator('pre').innerText(), /value14 = 140/);
  assert.doesNotMatch(await foo.locator('pre').innerText(), /TOP_CONTEXT/);
  async function assertHighlighting() {
    await foo.locator('span[style*="--diffs-token-light"]').first().waitFor();
    await page.waitForFunction(() => {
      const shadow = document.querySelector(
        'section[data-path="foo.ts"] .diff-host',
      )?.shadowRoot;
      const spans = shadow?.querySelectorAll(
        'span[style*="--diffs-token-light"]',
      );
      return (
        spans &&
        new Set([...spans].map((span) => getComputedStyle(span).color)).size > 1
      );
    });
    const colors = await foo.locator('.diff-host').evaluate((host) => {
      const shadow = host.shadowRoot;
      const addition = shadow.querySelector(
        '[data-line][data-line-type="change-addition"]',
      );
      const deletion = shadow.querySelector(
        '[data-line][data-line-type="change-deletion"]',
      );
      return {
        addition: getComputedStyle(addition).backgroundColor,
        deletion: getComputedStyle(deletion).backgroundColor,
        stylesheets: shadow.adoptedStyleSheets.length,
      };
    });
    assert.equal(colors.stylesheets, 1, 'Diff stylesheet is missing');
    assert.notEqual(
      colors.addition,
      colors.deletion,
      'Addition and deletion colors are identical',
    );
    for (const color of [colors.addition, colors.deletion]) {
      assert.ok(
        color !== 'transparent' && color !== 'rgba(0, 0, 0, 0)',
        'Diff background is transparent',
      );
    }
  }
  await assertHighlighting();
  await foo
    .getByRole('button', { name: 'Show full file', exact: true })
    .click();
  await foo.locator('pre').filter({ hasText: 'TOP_CONTEXT' }).waitFor();
  assert.match(await foo.locator('pre').innerText(), /TOP_CONTEXT/);
  assert.match(await foo.locator('pre').innerText(), /END_CONTEXT/);
  assert.doesNotMatch(
    await foo.locator('pre').innerText(),
    /EDITED_AFTER_SNAPSHOT/,
  );
  await page.getByRole('button', { name: 'Unified', exact: true }).click();
  assert.equal(
    await page.locator('#unified').getAttribute('aria-pressed'),
    'true',
  );
  await foo
    .getByRole('button', { name: 'Show changes only', exact: true })
    .click();
  await foo
    .locator('pre')
    .filter({ hasText: 'value14 = 140', hasNotText: 'TOP_CONTEXT' })
    .waitFor();
  assert.doesNotMatch(await foo.locator('pre').innerText(), /TOP_CONTEXT/);
  await page.getByRole('button', { name: 'Side-by-side', exact: true }).click();
  assert.equal(
    await page.locator('#split').getAttribute('aria-pressed'),
    'true',
  );
  assert.match(
    await page.locator('section[data-path="binary.dat"]').innerText(),
    /Binary file changed/,
  );
  assert.match(
    await page.locator('section[data-path="renamed.txt"]').innerText(),
    /old.txt → renamed.txt/,
  );
  const renamed = page.locator('section[data-path="renamed.txt"]');
  await renamed
    .getByRole('button', { name: 'Show full file', exact: true })
    .click();
  await renamed
    .locator('pre')
    .filter({ hasText: 'unchanged rename' })
    .waitFor();
  await renamed
    .getByRole('button', { name: 'Show changes only', exact: true })
    .click();
  assert.match(
    await page.locator('section[data-path="deleted.txt"]').innerText(),
    /deleted/,
  );
  await page.screenshot({ path: 'bin/viewer.png', fullPage: true });
  await page.emulateMedia({ colorScheme: 'dark' });
  await assertHighlighting();
  await page.screenshot({ path: 'bin/viewer-dark.png', fullPage: true });
  await page.setViewportSize({ width: 600, height: 800 });
  await page.reload();
  await page.locator('.diff-host pre').first().waitFor();
  assert.equal(
    await page.locator('#unified').getAttribute('aria-pressed'),
    'true',
  );
  assert.ok(await page.locator('#files').isHidden());
  assert.deepEqual(errors, [], `Browser errors: ${errors.join('\n')}`);
  assert.deepEqual(external, [], `Unexpected external requests: ${external}`);
  await browser.close();
  browser = undefined;
  const exited = once(server, 'exit');
  server.kill('SIGINT');
  const [code] = await exited;
  assert.equal(code, 0, 'Ctrl-C should shut down cleanly');
  await assert.rejects(fetch(url));
  server = undefined;
  const expected = git('diff', '--stat', '--no-color');
  const actual = execFileSync(binary, ['--stat', '--no-color'], {
    cwd: fixture,
    encoding: 'utf8',
  });
  assert.equal(actual, expected, 'Reporting output differs from Git');
  const status = spawn(binary, ['--quiet'], { cwd: fixture, stdio: 'ignore' });
  assert.equal((await once(status, 'exit'))[0], 1, 'Git exit code was lost');
  console.log(
    'Browser and executable checks passed: rendering, ordering, full-file snapshot, layout toggle, mobile, offline assets, reporting, exit status, Ctrl-C.',
  );
} finally {
  if (browser) await browser.close();
  if (server) server.kill('SIGTERM');
  await rm(fixture, { recursive: true, force: true });
}
