import { File as CodeFile, FileDiff, processFile, type FileDiffMetadata } from '@pierre/diffs';
import './style.css';

type FileSnapshot = { name: string; previousName?: string; kind: string; patch: string; fullPatch?: string; fullContent?: string };
type Snapshot = { repository: string; arguments: string[]; createdAt: string; files: FileSnapshot[] };
type View = { instance: FileDiff; host: HTMLElement; patch: FileDiffMetadata; full?: FileDiffMetadata; expanded: boolean };
const views: View[] = [];
let layout: 'split' | 'unified' = matchMedia('(max-width: 800px)').matches ? 'unified' : 'split';
const root = document.querySelector<HTMLElement>('#diffs')!;
const nav = document.querySelector<HTMLElement>('#files')!;

function updateLayout() {
  for (const mode of ['split', 'unified'] as const) {
    document.getElementById(mode)!.setAttribute('aria-pressed', String(mode === layout));
  }
  for (const view of views) {
    view.instance.setOptions(options());
    view.instance.rerender();
  }
}
function options() {
  return { diffStyle: layout, theme: { light: 'pierre-light', dark: 'pierre-dark' }, disableFileHeader: true, overflow: 'scroll' as const };
}
for (const mode of ['split', 'unified'] as const) {
  document.getElementById(mode)!.addEventListener('click', () => { layout = mode; updateLayout(); });
}
updateLayout();

function textElement<K extends keyof HTMLElementTagNameMap>(tag: K, text: string, className?: string): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  el.textContent = text;
  if (className) el.className = className;
  return el;
}

function addFile(file: FileSnapshot, index: number) {
  const section = document.createElement('section');
  section.className = 'file';
  section.id = `file-${index}`;
  section.dataset.path = file.name;
  const header = document.createElement('div');
  header.className = 'file-header';
  header.append(textElement('h2', file.previousName ? `${file.previousName} → ${file.name}` : file.name, 'file-name'));
  header.append(textElement('span', file.kind, 'file-kind'));
  section.append(header);
  root.append(section);

  const link = document.createElement('a');
  link.href = `#${section.id}`;
  link.append(textElement('span', ({ added: 'A', deleted: 'D', renamed: 'R', copied: 'C' } as Record<string, string>)[file.kind] ?? 'M', 'status'));
  link.append(textElement('span', file.name, 'name'));
  link.addEventListener('click', () => {
    nav.querySelector('.active')?.classList.remove('active');
    link.classList.add('active');
  });
  nav.append(link);

  let patch: FileDiffMetadata | undefined;
  try { patch = processFile(file.patch, { throwOnError: true, cacheKey: `patch-${index}` }); }
  catch (err) { console.error(err); }
  if (!patch || patch.hunks.length === 0) {
    const summary = textElement('div', file.patch.includes('Binary files ') || file.patch.includes('GIT binary patch') ? 'Binary file changed' : 'File metadata changed', 'file-summary');
    summary.append(textElement('pre', file.patch));
    section.append(summary);
    if (file.fullContent !== undefined) {
      const expand = textElement('button', 'Show full file');
      expand.setAttribute('aria-expanded', 'false');
      const host = document.createElement('diffs-container');
      host.className = 'diff-host';
      host.hidden = true;
      section.append(host);
      let instance: CodeFile | undefined;
      expand.addEventListener('click', () => {
        const expanded = host.hidden;
        host.hidden = !expanded;
        summary.hidden = expanded;
        if (expanded && !instance) {
          instance = new CodeFile(options());
          instance.render({ file: { name: file.name, contents: file.fullContent!, cacheKey: `metadata-${index}` }, fileContainer: host });
        }
        expand.textContent = expanded ? 'Show changes only' : 'Show full file';
        expand.setAttribute('aria-expanded', String(expanded));
      });
      header.append(expand);
    }
    return;
  }
  const host = document.createElement('diffs-container');
  host.className = 'diff-host';
  section.append(host);
  const instance = new FileDiff(options());
  const view: View = { instance, host, patch, expanded: false };
  views.push(view);
  instance.render({ fileDiff: patch, fileContainer: host });
  if (file.fullPatch && file.fullPatch !== file.patch) {
    const expand = textElement('button', 'Show full file');
    expand.setAttribute('aria-expanded', 'false');
    expand.addEventListener('click', () => {
      try {
        view.full ??= processFile(file.fullPatch!, { throwOnError: true, cacheKey: `full-${index}` });
        if (!view.full) throw new Error('Full-file context is unavailable');
        view.expanded = !view.expanded;
        instance.render({ fileDiff: view.expanded ? view.full : view.patch, fileContainer: host });
        expand.textContent = view.expanded ? 'Show changes only' : 'Show full file';
        expand.setAttribute('aria-expanded', String(view.expanded));
      } catch (err) { expand.textContent = 'Expansion unavailable'; expand.disabled = true; console.error(err); }
    });
    header.append(expand);
  }
}

try {
  const response = await fetch('./snapshot.json');
  if (!response.ok) throw new Error(`Could not load snapshot (${response.status})`);
  const snapshot: Snapshot = await response.json();
  root.replaceChildren();
  document.getElementById('repository')!.textContent = snapshot.repository.split('/').filter(Boolean).at(-1) ?? snapshot.repository;
  document.getElementById('repository')!.title = snapshot.repository;
  document.getElementById('file-count')!.textContent = String(snapshot.files.length);
  document.getElementById('snapshot-time')!.textContent = `Snapshot · ${new Date(snapshot.createdAt).toLocaleTimeString()}\n git diff ${snapshot.arguments.join(' ')}`;
  document.title = `${snapshot.repository.split('/').at(-1)} · diffthing`;
  snapshot.files.forEach(addFile);
  if (snapshot.files.length === 0) root.append(textElement('div', 'No changes.', 'empty'));
} catch (err) {
  root.replaceChildren(textElement('div', err instanceof Error ? err.message : 'Could not load diff.', 'empty'));
}
