import { describe, it, expect } from "vitest";
import { extractCodeRefs } from "../codeRefs";

describe("extractCodeRefs", () => {
  it("returns empty for messages with no refs", () => {
    expect(extractCodeRefs("idle (last: 100/200)")).toEqual([]);
    expect(extractCodeRefs("")).toEqual([]);
  });

  it("captures a single setting reference", () => {
    expect(extractCodeRefs("preflight failed at Setting 414")).toEqual([
      { kind: "setting", number: 414 },
    ]);
  });

  it("is case-insensitive on the kind word", () => {
    expect(extractCodeRefs("ALARM 152: bad number format")).toEqual([
      { kind: "alarm", number: 152 },
    ]);
    expect(extractCodeRefs("setting 11 misconfigured")).toEqual([
      { kind: "setting", number: 11 },
    ]);
  });

  it("captures multiple distinct refs in order", () => {
    const out = extractCodeRefs(
      "alarm 150 raised; check Setting 11 and Setting 12"
    );
    expect(out).toEqual([
      { kind: "alarm", number: 150 },
      { kind: "setting", number: 11 },
      { kind: "setting", number: 12 },
    ]);
  });

  it("dedupes repeated refs", () => {
    expect(
      extractCodeRefs("Setting 414 — Setting 414 — alarm 414")
    ).toEqual([
      { kind: "setting", number: 414 },
      { kind: "alarm", number: 414 },
    ]);
  });

  it("ignores numbers that aren't preceded by a kind word", () => {
    expect(extractCodeRefs("wrote line 414 / 2030")).toEqual([]);
    expect(extractCodeRefs("RPM 1815 spindle actual")).toEqual([]);
  });

  it("requires word boundaries on the kind", () => {
    // "presetting" should NOT match "setting" — \b prevents it.
    expect(extractCodeRefs("presetting 414")).toEqual([]);
  });

  it("captures parameter references", () => {
    expect(extractCodeRefs("parameter 2201 out of range")).toEqual([
      { kind: "parameter", number: 2201 },
    ]);
  });

  it("clamps to 5-digit numbers", () => {
    // 999999 has 6 digits — the regex caps at 5.
    expect(extractCodeRefs("alarm 999999")).toEqual([]);
  });
});
