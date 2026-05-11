// Detect manufacturer tool-part references in NC comment text and
// build search URLs for them. Used to pillify CAM tool-list comments
// like "(T5 D=0.5 - HARVEY 50050 - bullnose)" so the operator can
// click through to the vendor page without manually retyping the
// part number.
//
// Detection is curated by a vendor list (below) — operators add to it
// over time. A bare alphanumeric token with no vendor prefix is too
// noisy to auto-link (would catch line numbers, RPMs, dimensions).
//
// Self-contained module by design: kept independent of @/api so the
// extractor is unit-testable without a DOM environment.

// Vendor → domain. Domain is used to scope the Google site-search;
// when null, the search runs unscoped against the vendor name. Names
// match case-insensitively; multi-word names are listed verbatim and
// space-tolerant (a single regex spans whitespace).
//
// Ordering: longer multi-word names FIRST so they're tried before
// single-word collisions ("niagara cutter" before "niagara").
const VENDOR_DOMAINS: Array<[string, string | null]> = [
  ["lakeshore carbide", "lakeshorecarbide.com"],
  ["niagara cutter", "niagaracutter.com"],
  ["ma ford", "maford.com"],
  ["m.a. ford", "maford.com"],
  ["m.a ford", "maford.com"],
  ["harvey", "harveytool.com"],
  ["harvey tool", "harveytool.com"],
  ["osg", "osgtool.com"],
  ["helical", "helicaltool.com"],
  ["lakeshore", "lakeshorecarbide.com"],
  ["sandvik", "sandvik.coromant.com"],
  ["kennametal", "kennametal.com"],
  ["iscar", "iscar.com"],
  ["niagara", "niagaracutter.com"],
  ["garr", "garrtool.com"],
  ["mitsubishi", "mitsubishicarbide.com"],
  ["walter", "walter-tools.com"],
  ["seco", "secotools.com"],
  ["emuge", "emuge.com"],
  ["onsrud", "onsrud.com"],
  ["amana", "amanatool.com"],
  ["tungaloy", "tungaloyamerica.com"],
  ["korloy", "korloy.com"],
  ["hertel", "hertelusa.com"],
  ["kyocera", "kyoceraprecisiontools.com"],
  ["yg-1", "yg1.kr"],
  ["yg1", "yg1.kr"],
  ["whitney", "whitneytool.com"],
  ["fullerton", "fullertontool.com"],
  ["accupro", null],
  ["destiny", "destinytool.com"],
];

export interface MfgRef {
  vendor: string;       // canonical name, lower-cased
  partNumber: string;   // as captured from the comment
  matchedText: string;  // full "VENDOR PART" substring (for replacement)
  url: string;          // resolved search URL
}

const VENDOR_RE = (() => {
  const alt = VENDOR_DOMAINS
    .map(([name]) => name.replace(/[.+*?^$(){}|[\]\\]/g, "\\$&").replace(/ /g, "\\s+"))
    .join("|");
  // Vendor name (case-insensitive), optional separator (#:-) and any
  // whitespace, then a part-token. The part-token must start
  // alphanumeric and may contain hyphens / dots / slashes; total
  // length up to 32 chars (real Haas / CAM part numbers don't exceed
  // ~24). The `digit-required` guard happens at JS level since regex
  // lookaheads vary across legacy targets.
  return new RegExp(
    `\\b(${alt})\\b\\s*[#:\\-]?\\s*([A-Za-z0-9][A-Za-z0-9\\-./]{1,30})`,
    "gi"
  );
})();

// Words that look like part tokens but aren't — appears AFTER a vendor
// name but is clearly descriptive. Trims the bulk of false positives.
const PART_DENYLIST = new Set([
  "tool", "tools", "cutter", "cutters", "endmill", "endmills",
  "drill", "drills", "tap", "taps", "reamer", "reamers", "mill",
  "mills", "carbide", "hss", "coated", "uncoated", "coolant",
  "thru", "ballnose", "bullnose", "chamfer", "fluted", "flute",
  "flutes",
]);

// canonicalVendor returns the canonical (lower-case, single-space)
// form of a captured vendor match so the lookup hits the map.
function canonicalVendor(raw: string): string {
  return raw.toLowerCase().replace(/\s+/g, " ").trim();
}

function vendorDomain(vendor: string): string | null {
  for (const [name, domain] of VENDOR_DOMAINS) {
    if (name === vendor) return domain;
  }
  return null;
}

function buildSearchURL(vendor: string, part: string): string {
  const dom = vendorDomain(vendor);
  if (dom) {
    return `https://www.google.com/search?q=${encodeURIComponent(`site:${dom} ${part}`)}`;
  }
  return `https://www.google.com/search?q=${encodeURIComponent(`${vendor} ${part}`)}`;
}

// extractMfgRefs scans `text` for vendor+part-number patterns and
// returns one MfgRef per unique (vendor, partNumber) match. Order
// reflects appearance in the text.
export function extractMfgRefs(text: string): MfgRef[] {
  if (!text) return [];
  VENDOR_RE.lastIndex = 0;
  const seen = new Set<string>();
  const out: MfgRef[] = [];
  let m: RegExpExecArray | null;
  while ((m = VENDOR_RE.exec(text)) !== null) {
    const vendor = canonicalVendor(m[1]);
    const part = m[2];
    if (!/\d/.test(part)) continue; // must contain a digit somewhere
    if (PART_DENYLIST.has(part.toLowerCase())) continue;
    const key = `${vendor}::${part.toLowerCase()}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({
      vendor,
      partNumber: part,
      matchedText: m[0],
      url: buildSearchURL(vendor, part),
    });
  }
  return out;
}

// splitTextAroundRefs walks the input text and produces an alternating
// array of plain-text and ref segments, so the renderer can interleave
// pillified links inline without resorting to v-html.
//
// Returns segments in source order. Plain segments may be empty
// strings if a ref runs flush against another or the boundary.
export interface PlainSegment { type: "text"; text: string }
export interface RefSegment { type: "ref"; ref: MfgRef }
export type Segment = PlainSegment | RefSegment;

export function splitTextAroundRefs(text: string, refs: MfgRef[]): Segment[] {
  if (!text) return [];
  if (refs.length === 0) return [{ type: "text", text }];

  // Find each ref's offset by re-running the regex; cheaper than
  // tracking offsets through extractMfgRefs's dedup loop.
  type Hit = { start: number; end: number; ref: MfgRef };
  const hits: Hit[] = [];
  for (const ref of refs) {
    const idx = text.indexOf(ref.matchedText);
    if (idx < 0) continue;
    hits.push({ start: idx, end: idx + ref.matchedText.length, ref });
  }
  hits.sort((a, b) => a.start - b.start);

  const out: Segment[] = [];
  let cursor = 0;
  for (const h of hits) {
    if (h.start < cursor) continue; // overlap protection
    if (h.start > cursor) {
      out.push({ type: "text", text: text.slice(cursor, h.start) });
    }
    out.push({ type: "ref", ref: h.ref });
    cursor = h.end;
  }
  if (cursor < text.length) {
    out.push({ type: "text", text: text.slice(cursor) });
  }
  return out;
}
