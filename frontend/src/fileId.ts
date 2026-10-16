import type { FileSnapshot } from './types';

// A staged deletion and an untracked replacement can share the same path.
export function fileId(file: Pick<FileSnapshot, 'name' | 'kind'>) {
  const prefix = file.kind === 'untracked' ? 'untracked-file' : 'file';
  return `${prefix}-${file.name}`;
}
