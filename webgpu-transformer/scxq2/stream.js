import { decodeNode, hasCompleteNode } from "./decoder.js";

function concat(a, b) {
  const out = new Uint8Array(a.length + b.length);
  out.set(a, 0);
  out.set(b, a.length);
  return out;
}

export async function* decodeSCXQ2Stream(url) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`Unable to fetch SCXQ2 file: ${res.status}`);
  }

  const reader = res.body.getReader();
  let buffer = new Uint8Array();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer = concat(buffer, value);

    while (hasCompleteNode(buffer)) {
      const node = decodeNode(buffer);
      buffer = buffer.slice(node.size);
      yield node;
    }
  }
}
