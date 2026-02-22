function readU32LE(bytes, offset) {
  return (
    bytes[offset] |
    (bytes[offset + 1] << 8) |
    (bytes[offset + 2] << 16) |
    (bytes[offset + 3] << 24)
  ) >>> 0;
}

export function hasCompleteNode(buffer) {
  if (buffer.byteLength < 9) {
    return false;
  }

  const nameLength = buffer[1];
  if (buffer.byteLength < 6 + nameLength) {
    return false;
  }

  const dataLengthOffset = 2 + nameLength;
  const dataLength = readU32LE(buffer, dataLengthOffset);
  const total = 6 + nameLength + dataLength;
  return buffer.byteLength >= total;
}

export function decodeNode(buffer) {
  const typeByte = buffer[0];
  const nameLength = buffer[1];
  const nameStart = 2;
  const nameEnd = nameStart + nameLength;
  const dataLength = readU32LE(buffer, nameEnd);
  const dataStart = nameEnd + 4;
  const dataEnd = dataStart + dataLength;

  const decoder = new TextDecoder();
  const name = decoder.decode(buffer.slice(nameStart, nameEnd));

  return {
    type: typeByte === 1 ? "tensor" : "metadata",
    name,
    data: buffer.slice(dataStart, dataEnd),
    size: dataEnd,
  };
}
