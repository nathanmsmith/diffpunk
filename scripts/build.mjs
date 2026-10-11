import { build } from 'vite';
import { readFile, readdir, writeFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { join } from 'node:path';
const out = 'cli/assets/dist';
await build();
await writeFile(`${out}/.gitkeep`, '');

// Ship the licenses of the embedded production dependencies with the binary.
const packages = execFileSync(
  'npm',
  ['ls', '--omit=dev', '--all', '--parseable'],
  { encoding: 'utf8' },
)
  .trim()
  .split('\n')
  .slice(1);
const notices = [];
for (const dir of [...new Set(packages)].sort()) {
  const pkg = JSON.parse(await readFile(join(dir, 'package.json'), 'utf8'));
  const licenses = (await readdir(dir)).filter((name) =>
    /^(licen[cs]e|notice)(\..*)?$/i.test(name),
  );
  let notice = `${pkg.name} ${pkg.version}\nLicense: ${pkg.license ?? 'See package documentation'}\n`;
  for (const file of licenses)
    notice += `\n${await readFile(join(dir, file), 'utf8')}\n`;
  notices.push(notice);
}
await writeFile(
  `${out}/THIRD_PARTY_LICENSES.txt`,
  notices.join('\n' + '='.repeat(72) + '\n\n'),
);
