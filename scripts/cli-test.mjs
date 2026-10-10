import assert from 'node:assert/strict';
import { spawn, spawnSync, execFileSync } from 'node:child_process';
import { mkdtemp, writeFile, rm, mkdir, readFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { resolve, join } from 'node:path';
import { once } from 'node:events';
import { request } from 'node:http';

const binary = resolve('bin/diff-punk');
const fixture = await mkdtemp(join(tmpdir(), 'diff-punk-cli-'));
const outside = await mkdtemp(join(tmpdir(), 'diff-punk-files-'));
const git = (...args) =>
  execFileSync('git', args, { cwd: fixture, encoding: 'utf8' });
const run = (program, args) =>
  spawnSync(program, args, { cwd: fixture, encoding: 'utf8' });
let server;
try {
  git('init', '-q');
  git('config', 'user.name', 'Test');
  git('config', 'user.email', 'test@example.invalid');
  git('config', 'commit.gpgsign', 'false');
  const empty = run(binary, []);
  assert.equal(empty.status, 0);
  assert.equal(empty.stdout, '');
  assert.equal(empty.stderr, '');
  await writeFile(join(fixture, 'foo.txt'), 'one\ntwo\nthree\n');
  git('add', '.');
  git('commit', '-qm', 'initial');
  await writeFile(join(fixture, 'foo.txt'), 'one\nchanged\nthree\n');
  for (const args of [
    ['--stat'],
    ['--untracked', '--stat'],
    ['--name-only'],
    ['--name-status'],
    ['--raw'],
    ['--word-diff=plain'],
    ['--quiet'],
    ['--exit-code'],
    ['--check'],
    ['--no-patch'],
    ['--', 'missing-file'],
  ]) {
    const expected = run('git', [
      '--no-pager',
      'diff',
      '--no-color',
      ...args.filter((arg) => arg !== '--untracked'),
    ]);
    const actual = run(binary, ['--no-color', ...args]);
    assert.equal(
      actual.status,
      expected.status,
      `${args}: exit status differs`,
    );
    assert.equal(actual.stdout, expected.stdout, `${args}: output differs`);
  }
  const bad = run(binary, ['--unknown-diff-punk-option']);
  assert.equal(
    bad.status,
    run('git', ['diff', '--unknown-diff-punk-option']).status,
  );
  await writeFile(join(fixture, 'left.txt'), 'left\n');
  await writeFile(join(fixture, 'right.txt'), 'right\n');
  const noIndexArgs = ['--no-color', '--no-index', 'left.txt', 'right.txt'];
  const native = run('git', ['diff', ...noIndexArgs]);
  const noIndex = run(binary, noIndexArgs);
  assert.equal(noIndex.status, native.status);
  assert.equal(noIndex.stdout, native.stdout);
  await writeFile(join(outside, 'left.txt'), 'left\n');
  await writeFile(join(outside, 'right.txt'), 'right\n');
  const implicitNative = spawnSync(
    'git',
    ['diff', '--no-color', 'left.txt', 'right.txt'],
    { cwd: outside, encoding: 'utf8' },
  );
  const implicit = spawnSync(binary, ['--no-color', 'left.txt', 'right.txt'], {
    cwd: outside,
    encoding: 'utf8',
  });
  assert.equal(implicit.status, implicitNative.status);
  assert.equal(implicit.stdout, implicitNative.stdout);
  // Intercept the OS launcher to verify the default browser gets the exact URL
  // without actually opening a window during tests.
  const launchers = join(fixture, 'launchers');
  await mkdir(launchers);
  for (const name of ['open', 'xdg-open']) {
    await writeFile(
      join(launchers, name),
      '#!/bin/sh\nprintf "%s" "$1" > "$DIFF_PUNK_BROWSER_LOG"\n',
      { mode: 0o755 },
    );
  }
  const browserLog = join(fixture, 'browser-url');
  const env = {
    ...process.env,
    PATH: `${launchers}:${process.env.PATH}`,
    DIFF_PUNK_BROWSER_LOG: browserLog,
  };
  delete env.DIFF_PUNK_NO_OPEN;
  server = spawn(binary, [], {
    cwd: fixture,
    env,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const url = await new Promise((resolveURL, reject) => {
    let output = '';
    const timeout = setTimeout(
      () => reject(new Error(`Server did not start: ${output}`)),
      10000,
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
      reject(new Error(`Unexpected exit ${code}: ${output}`));
    });
  });
  const snapshot = await (await fetch(`${url}snapshot.json`)).json();
  const tracked = snapshot.files.filter((file) => file.kind !== 'untracked');
  assert.equal(tracked.length, 1);
  assert.equal(tracked[0].name, 'foo.txt');
  assert.match(tracked[0].fullPatch, /three/);
  assert.equal(snapshot.untrackedAvailable, true);
  assert.equal(snapshot.showUntracked, false);
  assert.ok(
    snapshot.files.some(
      (file) => file.name === 'left.txt' && file.kind === 'untracked',
    ),
  );
  await writeFile(join(fixture, 'foo.txt'), 'edited after snapshot\n');
  assert.deepEqual(await (await fetch(`${url}snapshot.json`)).json(), snapshot);
  const html = await (await fetch(url)).text();
  assert.match(html, /assets\/main.js/);
  for (const asset of ['assets/main.js', 'assets/main.css'])
    assert.equal((await fetch(`${url}${asset}`)).status, 200);
  assert.equal((await fetch(new URL('/snapshot.json', url))).status, 404);
  const wrongHostStatus = await new Promise((resolveStatus, reject) => {
    const req = request(
      `${url}snapshot.json`,
      { headers: { Host: 'evil.example' } },
      (response) => {
        response.resume();
        resolveStatus(response.statusCode);
      },
    );
    req.once('error', reject);
    req.end();
  });
  assert.equal(wrongHostStatus, 404);
  assert.equal(
    (await fetch(`${url}snapshot.json`, { method: 'POST' })).status,
    405,
  );
  let openedURL;
  const deadline = Date.now() + 5000;
  while (Date.now() < deadline) {
    try {
      openedURL = await readFile(browserLog, 'utf8');
      if (openedURL) break;
    } catch {
      // The launcher may not have written its log yet; retry until the deadline.
    }
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  assert.equal(
    openedURL,
    url,
    'Default browser launcher did not receive the URL',
  );
  const exited = once(server, 'exit');
  server.kill('SIGINT');
  assert.equal((await exited)[0], 0);
  await assert.rejects(fetch(url));
  server = undefined;
  console.log(
    'Executable checks passed: Git output/exit codes, empty diff, default browser launcher, frozen snapshot, embedded assets, request isolation, Ctrl-C.',
  );
} finally {
  if (server) {
    const exited = once(server, 'exit');
    server.kill('SIGTERM');
    await exited;
  }
  await rm(fixture, { recursive: true, force: true });
  await rm(outside, { recursive: true, force: true });
}
