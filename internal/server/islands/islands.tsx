import { Game } from "./game.tsx";
import { createRoot } from "react-dom/client";
import { createElement } from "react";

export const ISLANDS = {
  Game,
};

const islands = document.querySelectorAll("[data-island]");

async function main() {
  for (let i = 0; i < islands.length; i++) {
    const island = islands.item(i);

    const name = island.getAttribute("data-island");
    if (!name || !ISLANDS[name as keyof typeof ISLANDS]) {
      console.error("island not found:", name);
      return;
    }

    const propsAttr = island.getAttribute("data-props");
    const props = propsAttr ? await decodeBase64GzipJson(propsAttr) : null;
    try {
      island.removeAttribute("data-island");
      island.removeAttribute("data-props");
    } catch {
      // ignore
    }

    createRoot(island).render(
      createElement(ISLANDS[name as keyof typeof ISLANDS], props as any),
    );
  }
}

/**
 * Converts a Base64-encoded string to a Uint8Array.
 * @param base64 - The Base64 encoded string.
 * @returns The decoded bytes.
 */
function base64ToUint8Array(base64: string): Uint8Array {
  const binaryString = atob(base64);
  const len = binaryString.length;
  const bytes = new Uint8Array(len);
  for (let i = 0; i < len; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes;
}

/**
 * Decodes a Base64-encoded, gzipped JSON string into an object using DecompressionStream.
 * @param encoded - The Base64-encoded string.
 * @returns A promise that resolves with the parsed JSON object.
 */
async function decodeBase64GzipJson<T>(encoded: string): Promise<T> {
  // Step 1: Base64-decode, which gives us the gzipped bytes.
  const compressedBytes = base64ToUint8Array(encoded);

  // Step 2: Create a ReadableStream from the compressed bytes.
  // The Response constructor accepts a Uint8Array and its body is a stream.
  const compressedStream = new Response(compressedBytes).body;
  if (!compressedStream) {
    throw new Error("Failed to create stream from compressed bytes.");
  }

  // Step 3: Decompress using the DecompressionStream API.
  const ds = new DecompressionStream("gzip");
  const decompressedStream = compressedStream.pipeThrough(ds);

  // Step 4: Read the decompressed stream as text.
  const decompressedText = await new Response(decompressedStream).text();

  // Step 5: Parse the JSON text.
  return JSON.parse(decompressedText);
}

if (islands && typeof islands.length === "number" && islands.length > 0) {
  main();
}
