import { describe, it, expect } from "vitest";
import { extractMfgRefs, splitTextAroundRefs } from "../mfgRefs";

describe("extractMfgRefs", () => {
  it("returns empty for unmatched text", () => {
    expect(extractMfgRefs("1/2 endmill")).toEqual([]);
    expect(extractMfgRefs("")).toEqual([]);
  });

  it("matches a single Harvey part number", () => {
    const out = extractMfgRefs("HARVEY 50050 bullnose");
    expect(out).toHaveLength(1);
    expect(out[0].vendor).toBe("harvey");
    expect(out[0].partNumber).toBe("50050");
    expect(out[0].url).toContain("harveytool.com");
    expect(out[0].url).toContain("50050");
  });

  it("matches an OSG hyphenated part", () => {
    const out = extractMfgRefs("OSG-A-7-1234 tap");
    expect(out).toHaveLength(1);
    expect(out[0].vendor).toBe("osg");
    expect(out[0].partNumber).toBe("A-7-1234");
  });

  it("is case-insensitive on the vendor", () => {
    const out = extractMfgRefs("harvey 50050");
    expect(out[0].vendor).toBe("harvey");
  });

  it("rejects part tokens without a digit", () => {
    expect(extractMfgRefs("HARVEY tool")).toEqual([]);
    expect(extractMfgRefs("OSG endmill")).toEqual([]);
  });

  it("dedupes repeated refs", () => {
    const out = extractMfgRefs("HARVEY 50050 - harvey 50050");
    expect(out).toHaveLength(1);
  });

  it("captures multi-word vendor names", () => {
    const out = extractMfgRefs("LAKESHORE CARBIDE LC-12345");
    expect(out).toHaveLength(1);
    expect(out[0].vendor).toBe("lakeshore carbide");
    expect(out[0].url).toContain("lakeshorecarbide.com");
  });

  it("prefers the longer multi-word match over the single-word fallback", () => {
    // "niagara cutter NC123" should match "niagara cutter NC123",
    // NOT split into "niagara" + "cutter".
    const out = extractMfgRefs("Niagara Cutter NC123");
    expect(out).toHaveLength(1);
    expect(out[0].vendor).toBe("niagara cutter");
    expect(out[0].partNumber).toBe("NC123");
  });

  it("handles multiple refs in one comment", () => {
    const out = extractMfgRefs("primary HARVEY 50050; backup HELICAL EUDP3-30000");
    expect(out.map((r) => r.vendor)).toEqual(["harvey", "helical"]);
  });

  it("falls back to unscoped search when vendor has no domain", () => {
    const out = extractMfgRefs("AccuPro AP-9988");
    expect(out).toHaveLength(1);
    expect(out[0].url).not.toContain("site:");
    expect(out[0].url).toContain("AP-9988");
  });

  it("ignores numbers without a vendor prefix", () => {
    expect(extractMfgRefs("0.5 dia 4.0 length")).toEqual([]);
    expect(extractMfgRefs("RPM 12000 feed 50")).toEqual([]);
  });
});

describe("splitTextAroundRefs", () => {
  it("returns single text segment when no refs", () => {
    expect(splitTextAroundRefs("hello world", [])).toEqual([
      { type: "text", text: "hello world" },
    ]);
  });

  it("interleaves plain and ref segments in source order", () => {
    const text = "bullnose HARVEY 50050 long-reach";
    const refs = extractMfgRefs(text);
    const segs = splitTextAroundRefs(text, refs);
    expect(segs).toEqual([
      { type: "text", text: "bullnose " },
      { type: "ref", ref: refs[0] },
      { type: "text", text: " long-reach" },
    ]);
  });

  it("returns empty array for empty text", () => {
    expect(splitTextAroundRefs("", [])).toEqual([]);
  });

  it("handles ref at start of string", () => {
    const text = "HARVEY 50050 description";
    const refs = extractMfgRefs(text);
    const segs = splitTextAroundRefs(text, refs);
    expect(segs[0].type).toBe("ref");
    expect(segs[segs.length - 1]).toEqual({ type: "text", text: " description" });
  });
});
