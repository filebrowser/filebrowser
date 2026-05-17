import { describe, it, expect } from "vitest";
import {
  buildLineMap,
  resolveByN,
  resolveByPosition,
} from "../ncLineMap";

const SAMPLE = `%
O03002
(spot drill 4)
N30 T14 M6
N35 S5000 M3
N40 G17 G90
N55 G0 X0.6693 Y-2.9921
N60 G43 Z1.2746 H14
N70 G0 Z0.8746
N75 G98 G81 X0.6693 Y-2.9921 Z0.1075 R0.3575 F20.
N80 Y2.9921
N85 G80
N90 G0 Z1.2746
N95 M9
N100 M5
N105 G53 G0 Z0.
M30
%
`;

describe("ncLineMap", () => {
  const map = buildLineMap(SAMPLE);

  it("indexes N-blocks to source lines", () => {
    // Sample's first N30 is on line 4 (1-based: %=1, O03002=2, comment=3).
    expect(resolveByN(map, 30)).toBe(4);
    expect(resolveByN(map, 75)).toBe(10);
    expect(resolveByN(map, 99999)).toBeNull();
    expect(resolveByN(map, 0)).toBeNull();
  });

  it("tracks modal XYZ across lines", () => {
    // After N70 (line 9: G0 Z0.8746), modal state is X=0.6693, Y=-2.9921, Z=0.8746
    const at70 = map.positions.find((p) => p.line === 9);
    expect(at70).toBeDefined();
    expect(at70!.x).toBeCloseTo(0.6693);
    expect(at70!.y).toBeCloseTo(-2.9921);
    expect(at70!.z).toBeCloseTo(0.8746);
  });

  it("resolves a position to the nearest motion line", () => {
    // Position at (0.67, -2.99, 0.87) → should land on line 9 (N70)
    const got = resolveByPosition(
      map,
      { x: 0.67, y: -2.99, z: 0.87 },
      0.05
    );
    expect(got).toBe(9);
  });

  it("returns null when position is far from any line", () => {
    const got = resolveByPosition(map, { x: 999, y: 999, z: 999 }, 0.25);
    expect(got).toBeNull();
  });

  it("ignores axis words inside ( ) comments", () => {
    const m = buildLineMap("N10 G0 (X9 Y9 Z9 reference only)\nN20 G1 X1 Y1");
    // Line 1 should have no positions; line 2 should be at (1, 1, 0).
    expect(m.positions[0].line).toBe(2);
    expect(m.positions[0].x).toBe(1);
    expect(m.positions[0].y).toBe(1);
    expect(m.positions[0].z).toBe(0);
  });

  it("handles semicolon comments", () => {
    const m = buildLineMap("N10 G1 X5 Y5 ; X9 Y9 ignored");
    expect(m.positions[0].x).toBe(5);
    expect(m.positions[0].y).toBe(5);
  });

  it("does not match G41 / G54 as axis words", () => {
    const m = buildLineMap("N10 G54 G41 X1 Y2");
    expect(m.positions[0].x).toBe(1);
    expect(m.positions[0].y).toBe(2);
  });
});
