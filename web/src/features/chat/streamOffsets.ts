const streamOffsets = new Map<string, number>();

export function readStreamOffset(sessionID: string) {
  return streamOffsets.get(sessionID);
}

export function writeStreamOffset(sessionID: string, offset: number) {
  streamOffsets.set(sessionID, offset);
}

export function clearStreamOffset(sessionID: string) {
  streamOffsets.delete(sessionID);
}
