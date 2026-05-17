// Parse an NC program once and build two parallel indices:
//   1. {N → line} so a block number reported by Haas macro #3030 can
//      be resolved to a 1-based source line for highlight.
//   2. [{line, x, y, z}] running modal X/Y/Z so a current Mach position
//      can be matched to the nearest cut block when the macro isn't
//      populated.
//
// Coarse heuristic by design — won't be frame-accurate during fast
// rapids, but good enough to anchor the operator's read on a paused
// program (FEED HOLD / single-block) which is the common follow-along
// case for an externally-loaded run.

export interface NcLineMap {
  /** Number of source lines parsed. */
  totalLines: number;
  /** N-block → first source line containing that block (1-based). */
  byN: Map<number, number>;
  /** Per source line: the modal (X, Y, Z) after that line ran. */
  positions: NcLinePos[];
  /** Lines that carry at least one motion word — for nearest-pos match. */
  motionLines: number[];
}

export interface NcLinePos {
  line: number;
  x: number;
  y: number;
  z: number;
}

const N_RE = /\bN(\d+)\b/i;
const AXIS_RE = /(?<![A-Za-z])([XYZ])(-?\d*\.?\d+)/gi;

// stripComment removes (...) and ;... so axis words inside a comment
// don't get picked up as motion. Block-delete (/) at the start is
// retained — it doesn't affect parsing.
function stripComment(line: string): string {
  // ( ... ) — Haas-style inline comment. Drop everything between the
  // first ( and the matching ).
  let out = "";
  let depth = 0;
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (ch === "(") {
      depth++;
      continue;
    }
    if (ch === ")") {
      if (depth > 0) depth--;
      continue;
    }
    if (ch === ";") {
      break;
    }
    if (depth === 0) out += ch;
  }
  return out;
}

/**
 * buildLineMap parses content and returns the {N→line} + position map.
 * Modal carryover: any axis word not present on a line inherits the
 * previous value. Initial state is (0, 0, 0). Lines with no axis word
 * and no N number are still counted in totalLines but don't appear in
 * positions / motionLines.
 */
export function buildLineMap(content: string): NcLineMap {
  const lines = content.split(/\r?\n/);
  const byN = new Map<number, number>();
  const positions: NcLinePos[] = [];
  const motionLines: number[] = [];

  let x = 0, y = 0, z = 0;

  for (let i = 0; i < lines.length; i++) {
    const lineNo = i + 1;
    const stripped = stripComment(lines[i]);
    if (!stripped.trim()) continue;

    const nMatch = stripped.match(N_RE);
    if (nMatch) {
      const n = parseInt(nMatch[1], 10);
      if (!byN.has(n)) byN.set(n, lineNo);
    }

    let touched = false;
    AXIS_RE.lastIndex = 0;
    let m: RegExpExecArray | null;
    while ((m = AXIS_RE.exec(stripped)) !== null) {
      const axis = m[1].toUpperCase();
      const val = parseFloat(m[2]);
      if (Number.isNaN(val)) continue;
      touched = true;
      if (axis === "X") x = val;
      else if (axis === "Y") y = val;
      else if (axis === "Z") z = val;
    }
    if (touched) {
      positions.push({ line: lineNo, x, y, z });
      motionLines.push(lineNo);
    }
  }

  return {
    totalLines: lines.length,
    byN,
    positions,
    motionLines,
  };
}

/** resolveByN returns the source line for the given N-block, or null. */
export function resolveByN(map: NcLineMap, n: number): number | null {
  if (!n || n <= 0) return null;
  return map.byN.get(n) ?? null;
}

/**
 * resolveByPosition returns the source line whose modal (X, Y, Z) end
 * point is closest to the current machine position. Z is down-weighted
 * (CNC programs frequently retract to clearance Z; matching XY end
 * points is much more informative). Returns null when the map has no
 * motion lines.
 *
 * `tolerance` (inches) is the max distance considered a "match" — beyond
 * this we return null rather than highlighting a meaningless nearest.
 */
export function resolveByPosition(
  map: NcLineMap,
  pos: { x: number; y: number; z: number },
  tolerance = 0.25
): number | null {
  if (map.positions.length === 0) return null;
  const Z_WEIGHT = 0.25;
  let bestLine: number | null = null;
  let bestDist = Number.POSITIVE_INFINITY;
  for (const p of map.positions) {
    const dx = p.x - pos.x;
    const dy = p.y - pos.y;
    const dz = (p.z - pos.z) * Z_WEIGHT;
    const d2 = dx * dx + dy * dy + dz * dz;
    if (d2 < bestDist) {
      bestDist = d2;
      bestLine = p.line;
    }
  }
  if (bestLine == null) return null;
  if (Math.sqrt(bestDist) > tolerance) return null;
  return bestLine;
}
