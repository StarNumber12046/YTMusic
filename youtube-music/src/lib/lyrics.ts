import type { ModelsLyricsLine } from "../generated/types.gen";

export interface ParsedLyricsLine {
  timeMs: number;
  text: string;
}

export function parseLRC(lrc: string): ParsedLyricsLine[] {
  const lines = lrc.split("\n");
  const parsed: ParsedLyricsLine[] = [];
  const lrcRegex = /\[(\d{2}):(\d{2})\.(\d{2,3})\](.*)/;

  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) continue;

    const match = lrcRegex.exec(trimmed);
    if (match) {
      const minutes = parseInt(match[1], 10);
      const seconds = parseInt(match[2], 10);
      const msStr = match[3];
      const ms = msStr.length === 2 ? parseInt(msStr, 10) * 10 : parseInt(msStr, 10);

      const timeMs = minutes * 60 * 1000 + seconds * 1000 + ms;
      const text = match[4].trim();

      if (text) {
        parsed.push({ timeMs, text });
      }
    }
  }

  return parsed;
}

export function findCurrentLineIndex(
  lines: ParsedLyricsLine[] | ModelsLyricsLine[],
  currentTimeMs: number
): number {
  if (!lines || lines.length === 0) return -1;

  for (let i = lines.length - 1; i >= 0; i--) {
    const line = lines[i];
    const lineTime = "time_ms" in line ? (line as any).time_ms : (line as any).timeMs;
    if (currentTimeMs >= lineTime) {
      return i;
    }
  }

  return 0;
}

export function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}
