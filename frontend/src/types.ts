import type { FileDiffOptions } from '@pierre/diffs/react';

export type FileSnapshot = {
  name: string;
  previousName?: string;
  kind: string;
  patch: string;
  fullPatch?: string;
  fullContent?: string;
};
export type Snapshot = {
  repository: string;
  arguments: string[];
  createdAt: string;
  files: FileSnapshot[];
};
export type Layout = 'split' | 'unified';
export type ViewOptions = Pick<
  FileDiffOptions<undefined, undefined>,
  'diffStyle' | 'theme' | 'disableFileHeader' | 'overflow'
>;
