import { cp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

const targets = ['chrome', 'firefox'];

await rm('dist', { recursive: true, force: true });
for (const target of targets) {
  const destination = join('dist', target);
  await mkdir(destination, { recursive: true });
  await cp('shared/src', destination, { recursive: true });
  const manifest = await readFile(join(target, 'manifest.json'), 'utf8');
  await writeFile(join(destination, 'manifest.json'), manifest);
}

console.log(`Built ${targets.join(' and ')} extensions in extensions/dist.`);

