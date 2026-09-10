import { readFile, copyFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const root = resolve(process.argv[2] || '/tmp/opuspurego/dist/libopus/wasm');
// Emscripten's Node loader needs a file URL and an explicit ES module suffix.
await copyFile(`${root}/libopus.js`, `${root}/libopus.mjs`);
const { default: createLibopus } = await import(pathToFileURL(`${root}/libopus.mjs`));
const m = await createLibopus({ wasmBinary: await readFile(`${root}/libopus.wasm`) });
const status = m._malloc(4), pcm = m._malloc(1920), packet = m._malloc(4000);
let enc = 0, dec = 0;
try {
  enc = m._opus_encoder_create(48000, 1, 2049, status);
  if (!enc || m.getValue(status, 'i32')) throw new Error('encoder_create');
  dec = m._opus_decoder_create(48000, 1, status);
  if (!dec || m.getValue(status, 'i32')) throw new Error('decoder_create');
  for (let i = 0; i < 960; i++) m.setValue(pcm + i * 2, Math.round(12000 * Math.sin(2 * Math.PI * 440 * i / 48000)), 'i16');
  const n = m._opus_encode(enc, pcm, 960, packet, 4000);
  if (n <= 0) throw new Error(`encode: ${n}`);
  if (m._opus_decode(dec, packet, n, pcm, 960, 0) !== 960) throw new Error('decode');
  if (m._opus_decode(dec, 0, 0, pcm, 960, 0) !== 960) throw new Error('PLC');
  console.log(`${m.UTF8ToString(m._opus_get_version_string())}: WASM encode/decode/PLC passed`);
} finally {
  if (enc) m._opus_encoder_destroy(enc);
  if (dec) m._opus_decoder_destroy(dec);
  m._free(status); m._free(pcm); m._free(packet);
}
