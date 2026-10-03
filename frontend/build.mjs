// Builds the frontend into backend/web/dist, which the Go server embeds.
//
//   node build.mjs          production build (minified)
//   node build.mjs --watch  rebuild on change; run the server with WEB_DIR=web/dist
//
// OUT_DIR overrides the output directory.
import { spawn } from 'node:child_process';
import { cp, mkdir, readdir, rm, watch } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import * as esbuild from 'esbuild';

const root = path.dirname(fileURLToPath(import.meta.url));
const out = path.resolve(root, process.env.OUT_DIR || '../backend/web/dist');
const watching = process.argv.includes('--watch');

// Static files copied as-is: [source, destination inside out/].
const statics = [
  ['src/index.html', 'index.html'],
  ['src/favicon.svg', 'favicon.svg'],
  ['src/og.png', 'og.png'],
  // Icons: Google Search needs a raster favicon (multiple of 48px), browsers and
  // crawlers request /favicon.ico and /apple-touch-icon.png directly.
  ['src/favicon.ico', 'favicon.ico'],
  ['src/apple-touch-icon.png', 'apple-touch-icon.png'],
  ['src/icon-192.png', 'icon-192.png'],
  ['src/icon-512.png', 'icon-512.png'],
  ['src/site.webmanifest', 'site.webmanifest'],
  // Persian-only subset of the variable font (all weights in one ~48 KB file);
  // Latin text falls back to the system UI font.
  ['node_modules/vazirmatn/misc/Non-Latin/fonts/webfonts/Vazirmatn-NL[wght].woff2', 'assets/vazirmatn.woff2'],
];

async function clean() {
  await mkdir(out, { recursive: true });
  for (const name of await readdir(out)) {
    if (name !== '.gitkeep') await rm(path.join(out, name), { recursive: true, force: true });
  }
  await mkdir(path.join(out, 'assets'), { recursive: true });
}

async function copyStatics() {
  for (const [src, dest] of statics) {
    await cp(path.join(root, src), path.join(out, dest));
  }
}

function tailwind() {
  const args = ['-c', 'tailwind.config.js', '-i', 'src/input.css', '-o', path.join(out, 'assets/app.css'), '--minify'];
  if (watching) args.push('--watch');
  const bin = path.join(root, 'node_modules/.bin/tailwindcss');
  return new Promise((resolve, reject) => {
    const p = spawn(bin, args, { cwd: root, stdio: 'inherit' });
    if (watching) return resolve();
    p.on('exit', (code) => (code === 0 ? resolve() : reject(new Error(`tailwindcss exited with ${code}`))));
  });
}

const jsOptions = {
  entryPoints: [path.join(root, 'src/js/main.js')],
  outfile: path.join(out, 'assets/app.js'),
  bundle: true,
  format: 'iife',
  target: ['es2019'],
  minify: !watching,
  sourcemap: watching ? 'inline' : false,
  legalComments: 'none',
  charset: 'utf8',
  logLevel: 'info',
};

await clean();
await copyStatics();

if (watching) {
  const ctx = await esbuild.context(jsOptions);
  await ctx.watch();
  await tailwind();
  // Re-copy static files when they change.
  (async () => {
    for await (const _ of watch(path.join(root, 'src'))) await copyStatics().catch(console.error);
  })();
  console.log(`watching; output in ${out}`);
} else {
  await Promise.all([esbuild.build(jsOptions), tailwind()]);
  console.log(`built into ${out}`);
}
