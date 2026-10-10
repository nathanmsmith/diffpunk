import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { Window } from 'happy-dom';

// Exercise the actual library and frontend in a DOM when a native browser
// cannot be launched. The Playwright suite separately checks browser assets.
const scenario = process.argv[2] ?? 'viewer';
const window = new Window({ url: 'http://127.0.0.1:1234/session/' });
for (const key of [
  'window',
  'document',
  'HTMLElement',
  'HTMLDivElement',
  'HTMLPreElement',
  'HTMLSpanElement',
  'Element',
  'Node',
  'DocumentFragment',
  'ShadowRoot',
  'CSSStyleSheet',
  'ResizeObserver',
  'MutationObserver',
  'DOMException',
  'customElements',
  'getComputedStyle',
  'requestAnimationFrame',
  'cancelAnimationFrame',
  'matchMedia',
]) {
  const value = key === 'window' ? window : window[key];
  Object.defineProperty(globalThis, key, {
    value:
      typeof value === 'function' && /^[a-z]/.test(key)
        ? value.bind(window)
        : value,
    configurable: true,
  });
}
for (const key of Object.getOwnPropertyNames(window)) {
  if (/^(HTML|SVG|CSS|DOM|Intersection)/.test(key)) {
    Object.defineProperty(globalThis, key, {
      value: window[key],
      configurable: true,
    });
  }
}
const html = await readFile('assets/dist/index.html', 'utf8');
window.document.write(html);
const patch = `diff --git a/foo.ts b/foo.ts\nindex 1111111..2222222 100644\n--- a/foo.ts\n+++ b/foo.ts\n@@ -10,1 +10,1 @@\n-const value = 1;\n+const value = 2;\n`;
const full = `diff --git a/foo.ts b/foo.ts\nindex 1111111..2222222 100644\n--- a/foo.ts\n+++ b/foo.ts\n@@ -1,11 +1,11 @@\n // TOP_CONTEXT\n${Array.from({ length: 8 }, () => ' // context\n').join('')}-const value = 1;\n+const value = 2;\n // END_CONTEXT\n`;
const snapshot = {
  repository: '/test/repo',
  arguments: [],
  createdAt: '2026-10-08T12:00:00Z',
  untrackedAvailable: true,
  showUntracked: false,
  files: [
    {
      name: 'new.ts',
      kind: 'untracked',
      patch:
        'diff --git a/new.ts b/new.ts\nnew file mode 100644\n--- /dev/null\n+++ b/new.ts\n@@ -0,0 +1 @@\n+const fresh = true;\n',
    },
    { name: 'foo.ts', kind: 'modified', patch, fullPatch: full },
    {
      name: 'binary.dat',
      kind: 'modified',
      patch:
        'diff --git a/binary.dat b/binary.dat\nindex 1111111..2222222 100644\nBinary files a/binary.dat and b/binary.dat differ\n',
    },
    {
      name: 'old → new.txt',
      previousName: '<img src=x onerror=alert(1)>',
      kind: 'renamed',
      fullContent: 'unchanged rename\n',
      patch:
        'diff --git a/old.txt b/new.txt\nsimilarity index 100%\nrename from old.txt\nrename to new.txt\n',
    },
  ],
};
let releaseSnapshot;
const pendingSnapshot = new Promise((resolve) => {
  releaseSnapshot = resolve;
});
globalThis.fetch = async (url) => {
  assert.equal(url, './snapshot.json');
  await pendingSnapshot;
  if (scenario === 'http-error') return new Response('', { status: 503 });
  if (scenario === 'network-error') throw new Error('Network unavailable');
  if (scenario === 'invalid-json') return new Response('invalid snapshot');
  return new Response(
    JSON.stringify(
      scenario === 'empty'
        ? { ...snapshot, files: [] }
        : scenario === 'untracked-only'
          ? { ...snapshot, files: snapshot.files.slice(0, 1) }
          : scenario === 'untracked-enabled'
            ? { ...snapshot, showUntracked: true }
            : scenario === 'historical'
              ? {
                  ...snapshot,
                  untrackedAvailable: false,
                  files: snapshot.files.slice(1),
                }
              : snapshot,
    ),
  );
};
async function waitFor(predicate) {
  const deadline = Date.now() + 5000;
  while (!predicate()) {
    if (Date.now() > deadline) throw new Error('DOM render timed out');
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
}
try {
  const entry = html.match(/<script[^>]+src="([^"]+)"/)[1];
  await import(new URL(`../assets/dist/${entry}`, import.meta.url));
  await waitFor(() => window.document.getElementById('loading'));
  assert.equal(
    window.document.getElementById('loading').textContent,
    'Loading diff…',
  );
  releaseSnapshot();
  if (scenario !== 'viewer') {
    await waitFor(() => !window.document.getElementById('loading'));
    const message = window.document.querySelector('#diffs .empty')?.textContent;
    if (scenario === 'empty') {
      assert.equal(message, 'No changes.');
      assert.equal(
        window.document.getElementById('file-count').textContent,
        '0',
      );
      assert.equal(window.document.querySelectorAll('#files a').length, 0);
    } else if (
      scenario === 'untracked-only' ||
      scenario === 'untracked-enabled'
    ) {
      const toggle = window.document.getElementById('show-untracked');
      assert.equal(
        toggle.getAttribute('aria-pressed'),
        scenario === 'untracked-enabled' ? 'true' : 'false',
      );
      if (scenario === 'untracked-only') {
        assert.match(message, /Enable.*Show untracked files/);
        toggle.click();
      }
      await waitFor(() =>
        window.document
          .querySelector('section[data-path="new.ts"] .diff-host')
          ?.shadowRoot?.textContent.includes('const fresh = true'),
      );
      toggle.click();
      await waitFor(
        () => !window.document.querySelector('section[data-path="new.ts"]'),
      );
    } else if (scenario === 'historical') {
      assert.equal(window.document.getElementById('show-untracked'), null);
    } else if (scenario === 'http-error') {
      assert.equal(message, 'Could not load snapshot (503)');
    } else if (scenario === 'network-error') {
      assert.equal(message, 'Network unavailable');
    } else {
      assert.match(message, /JSON/);
    }
    console.log(`DOM ${scenario} checks passed.`);
  } else {
    await waitFor(() =>
      window.document.querySelector('section[data-path="foo.ts"]'),
    );
    const section = window.document.querySelector(
      'section[data-path="foo.ts"]',
    );
    assert.ok(section, `File not rendered: ${window.document.body.innerHTML}`);
    const pre = () =>
      section.querySelector('.diff-host').shadowRoot.querySelector('pre');
    await waitFor(() => pre()?.textContent.includes('const value = 2'));
    assert.ok(pre(), 'Diff library did not render');
    assert.match(pre().textContent, /const value = 2/);
    assert.doesNotMatch(pre().textContent, /TOP_CONTEXT/);
    const shadow = section.querySelector('.diff-host').shadowRoot;
    assert.equal(
      shadow.adoptedStyleSheets.length,
      1,
      'Diff stylesheet is missing',
    );
    await waitFor(() =>
      pre().querySelector('span[style*="--diffs-token-light"]'),
    );
    const tokenColors = new Set(
      [...pre().querySelectorAll('span[style*="--diffs-token-light"]')].map(
        (span) => span.style.getPropertyValue('--diffs-token-light'),
      ),
    );
    assert.ok(tokenColors.size > 1, 'Syntax tokens have no distinct colors');
    const button = section.querySelector('button');
    button.click();
    await waitFor(() => pre()?.textContent.includes('TOP_CONTEXT'));
    assert.match(pre().textContent, /TOP_CONTEXT/);
    assert.match(pre().textContent, /END_CONTEXT/);
    assert.equal(button.getAttribute('aria-expanded'), 'true');
    assert.equal(pre().getAttribute('data-diff-type'), 'split');
    window.document.getElementById('unified').click();
    await waitFor(() => pre()?.getAttribute('data-diff-type') === 'single');
    assert.equal(
      window.document.getElementById('unified').getAttribute('aria-pressed'),
      'true',
    );
    assert.equal(pre().getAttribute('data-diff-type'), 'single');
    assert.match(pre().textContent, /TOP_CONTEXT/);
    button.click();
    await waitFor(
      () =>
        pre()?.textContent.includes('const value = 2') &&
        !pre().textContent.includes('TOP_CONTEXT'),
    );
    assert.doesNotMatch(pre().textContent, /TOP_CONTEXT/);
    window.document.getElementById('split').click();
    await waitFor(() => pre()?.getAttribute('data-diff-type') === 'split');
    assert.equal(
      window.document.getElementById('split').getAttribute('aria-pressed'),
      'true',
    );
    await waitFor(() => pre()?.getAttribute('data-diff-type') === 'split');
    const rename = window.document.querySelector(
      'section[data-path="old → new.txt"]',
    );
    rename.querySelector('button').click();
    await waitFor(() =>
      rename
        .querySelector('.diff-host')
        ?.shadowRoot?.textContent.includes('unchanged rename'),
    );
    assert.equal(rename.querySelector('.file-summary').hidden, true);
    assert.equal(
      rename.querySelector('.diff-host').shadowRoot.adoptedStyleSheets.length,
      1,
      'Full-file stylesheet is missing',
    );
    rename.querySelector('button').click();
    await waitFor(() => !rename.querySelector('.file-summary').hidden);
    assert.equal(rename.querySelector('.file-summary').hidden, false);
    const links = window.document.querySelectorAll('#files a');
    links[0].click();
    await waitFor(() => links[0].classList.contains('active'));
    links[2].click();
    await waitFor(() => links[2].classList.contains('active'));
    assert.equal(links[0].classList.contains('active'), false);
    assert.equal(
      window.document.getElementById('repository').textContent,
      'repo',
    );
    assert.equal(window.document.getElementById('file-count').textContent, '3');
    assert.equal(window.document.title, 'repo · diff punk');
    assert.equal(window.document.querySelectorAll('section').length, 3);
    assert.match(
      window.document.querySelector('section[data-path="binary.dat"]')
        .textContent,
      /Binary file changed/,
    );
    assert.equal(
      window.document.querySelectorAll('.file-header img').length,
      0,
      'Filename interpreted as HTML',
    );
    const toggle = window.document.getElementById('show-untracked');
    assert.equal(toggle.getAttribute('aria-pressed'), 'false');
    toggle.click();
    await waitFor(() =>
      window.document
        .querySelector('section[data-path="new.ts"] .diff-host')
        ?.shadowRoot?.textContent.includes('const fresh = true'),
    );
    assert.equal(window.document.getElementById('file-count').textContent, '4');
    assert.equal(
      window.document.querySelector('section[data-path="new.ts"] .file-kind')
        .textContent,
      'Untracked',
    );
    for (const link of window.document.querySelectorAll('#files a')) {
      assert.ok(window.document.querySelector(link.getAttribute('href')));
    }
    toggle.click();
    await waitFor(
      () => window.document.getElementById('file-count').textContent === '3',
    );
    assert.equal(
      window.document.querySelector('section[data-path="new.ts"]'),
      null,
    );
    console.log(
      'DOM checks passed: diff stylesheet, syntax colors, actual diff rendering, full-file expansion/collapse, layout toggle, binary summary, escaped filenames.',
    );
  }
} finally {
  await window.happyDOM.abort();
}
if (scenario === 'viewer') {
  for (const testCase of [
    'empty',
    'untracked-only',
    'untracked-enabled',
    'historical',
    'http-error',
    'network-error',
    'invalid-json',
  ]) {
    execFileSync(process.execPath, [fileURLToPath(import.meta.url), testCase], {
      stdio: 'inherit',
    });
  }
}
