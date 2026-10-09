import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { Window } from 'happy-dom';

// Exercise the actual library and frontend in a DOM when a native browser
// cannot be launched. The Playwright suite separately checks browser assets.
const window = new Window({ url: 'http://127.0.0.1:1234/session/' });
for (const key of ['window', 'document', 'HTMLElement', 'HTMLDivElement', 'HTMLPreElement', 'HTMLSpanElement', 'Element', 'Node', 'DocumentFragment', 'ShadowRoot', 'CSSStyleSheet', 'ResizeObserver', 'MutationObserver', 'DOMException', 'customElements', 'getComputedStyle', 'requestAnimationFrame', 'cancelAnimationFrame', 'matchMedia']) {
  const value = key === 'window' ? window : window[key];
  Object.defineProperty(globalThis, key, { value: typeof value === 'function' && /^[a-z]/.test(key) ? value.bind(window) : value, configurable: true });
}
for (const key of Object.getOwnPropertyNames(window)) {
  if (/^(HTML|SVG|CSS|DOM|Intersection)/.test(key)) {
    Object.defineProperty(globalThis, key, { value: window[key], configurable: true });
  }
}
window.document.write(await readFile('web/index.html', 'utf8'));
const patch = `diff --git a/foo.ts b/foo.ts\nindex 1111111..2222222 100644\n--- a/foo.ts\n+++ b/foo.ts\n@@ -10,1 +10,1 @@\n-const value = 1;\n+const value = 2;\n`;
const full = `diff --git a/foo.ts b/foo.ts\nindex 1111111..2222222 100644\n--- a/foo.ts\n+++ b/foo.ts\n@@ -1,11 +1,11 @@\n // TOP_CONTEXT\n${Array.from({length: 8}, () => ' // context\n').join('')}-const value = 1;\n+const value = 2;\n // END_CONTEXT\n`;
const snapshot = { repository: '/test/repo', arguments: [], createdAt: '2026-10-08T12:00:00Z', files: [
  { name: 'foo.ts', kind: 'modified', patch, fullPatch: full },
  { name: 'binary.dat', kind: 'modified', patch: 'diff --git a/binary.dat b/binary.dat\nindex 1111111..2222222 100644\nBinary files a/binary.dat and b/binary.dat differ\n' },
  { name: 'old → new.txt', previousName: '<img src=x onerror=alert(1)>', kind: 'renamed', fullContent: 'unchanged rename\n', patch: 'diff --git a/old.txt b/new.txt\nsimilarity index 100%\nrename from old.txt\nrename to new.txt\n' },
] };
globalThis.fetch = async (url) => { assert.equal(url, './snapshot.json'); return new Response(JSON.stringify(snapshot)); };
async function waitFor(predicate) {
  const deadline = Date.now() + 5000;
  while (!predicate()) {
    if (Date.now() > deadline) throw new Error('DOM render timed out');
    await new Promise(resolve => setTimeout(resolve, 10));
  }
}
try {
  await import('../internal/viewer/dist/assets/main.js');
  const section = window.document.querySelector('section[data-path="foo.ts"]');
  assert.ok(section, `File not rendered: ${window.document.body.innerHTML}`);
  const pre = () => section.querySelector('.diff-host').shadowRoot.querySelector('pre');
  await waitFor(() => pre()?.textContent.includes('const value = 2'));
  assert.ok(pre(), 'Diff library did not render');
  assert.match(pre().textContent, /const value = 2/);
  assert.doesNotMatch(pre().textContent, /TOP_CONTEXT/);
  const button = section.querySelector('button');
  button.click();
  await waitFor(() => pre()?.textContent.includes('TOP_CONTEXT'));
  assert.match(pre().textContent, /TOP_CONTEXT/);
  assert.match(pre().textContent, /END_CONTEXT/);
  assert.equal(button.getAttribute('aria-expanded'), 'true');
  assert.equal(pre().getAttribute('data-diff-type'), 'split');
  window.document.getElementById('unified').click();
  await waitFor(() => pre()?.textContent.includes('TOP_CONTEXT'));
  assert.equal(window.document.getElementById('unified').getAttribute('aria-pressed'), 'true');
  assert.equal(pre().getAttribute('data-diff-type'), 'single');
  assert.match(pre().textContent, /TOP_CONTEXT/);
  button.click();
  await waitFor(() => pre()?.textContent.includes('const value = 2') && !pre().textContent.includes('TOP_CONTEXT'));
  assert.doesNotMatch(pre().textContent, /TOP_CONTEXT/);
  window.document.getElementById('split').click();
  assert.equal(window.document.getElementById('split').getAttribute('aria-pressed'), 'true');
  await waitFor(() => pre()?.getAttribute('data-diff-type') === 'split');
  const rename = window.document.querySelector('section[data-path="old → new.txt"]');
  rename.querySelector('button').click();
  await waitFor(() => rename.querySelector('.diff-host').shadowRoot?.textContent.includes('unchanged rename'));
  assert.equal(rename.querySelector('.file-summary').hidden, true);
  rename.querySelector('button').click();
  assert.equal(rename.querySelector('.file-summary').hidden, false);
  assert.equal(window.document.querySelectorAll('section').length, 3);
  assert.match(window.document.querySelector('section[data-path="binary.dat"]').textContent, /Binary file changed/);
  assert.equal(window.document.querySelectorAll('.file-header img').length, 0, 'Filename interpreted as HTML');
  console.log('DOM checks passed: actual diff rendering, full-file expansion/collapse, layout toggle, binary summary, escaped filenames.');
} finally {
  await window.happyDOM.abort();
}
