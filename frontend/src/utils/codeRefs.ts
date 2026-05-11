// Detect Haas alarm / setting / parameter references in a log message
// and resolve them against /api/cnc/codes/lookup. Used to enrich the
// activity log so an operator seeing "alarm 152" gets the human title
// without having to look it up manually.
//
// Detection is intentionally narrow:
//   - "Setting 414" / "setting 414"
//   - "Alarm 152" / "alarm 152"
//   - "Parameter 2201" / "parameter 2201"
// Bare numbers don't trigger — the activity log carries plenty of
// numbers (line counts, RPMs, latencies) and we don't want every "47"
// to fan out a lookup.
//
// NOTE: the extract path is kept independent of the @/api import so a
// unit test can pull this file in without dragging the api-utils
// transitive (auth / i18n / navigator) into the suite.

export type CodeKind = "setting" | "alarm" | "parameter";

export interface CodeEntry {
  kind: CodeKind;
  number: number;
  title: string;
  category?: string;
  summary: string;
  hint?: string;
}

export interface CodeRef {
  kind: CodeKind;
  number: number;
}

// Reference detector. Captures kind word + number. /gi to find every
// match in one pass; case-insensitive so "Alarm" and "alarm" both hit.
const codeRefRe = /\b(setting|alarm|parameter)\s+(\d{1,5})\b/gi;

// Extract refs in order of appearance, deduped by (kind, number) so a
// log line that mentions the same code twice only resolves once.
export function extractCodeRefs(msg: string): CodeRef[] {
  if (!msg) return [];
  const seen = new Set<string>();
  const out: CodeRef[] = [];
  let m: RegExpExecArray | null;
  codeRefRe.lastIndex = 0;
  while ((m = codeRefRe.exec(msg)) !== null) {
    const kind = m[1].toLowerCase() as CodeKind;
    const number = parseInt(m[2], 10);
    if (!Number.isFinite(number)) continue;
    const key = `${kind}:${number}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ kind, number });
  }
  return out;
}

// LRU-ish cache. Cap is small — operators reference a handful of
// distinct codes per session. Hits dominate after the first lookup;
// missing entries also cache (negative result) so we don't re-poll for
// codes the server doesn't know.
const cache = new Map<string, CodeEntry | null>();
const inFlight = new Map<string, Promise<CodeEntry | null>>();
const CACHE_CAP = 128;

function cacheKey(kind: CodeKind, number: number): string {
  return `${kind}:${number}`;
}

// resolveCodeRef returns the catalog entry for a (kind, number) ref,
// or null if the server has no entry. De-dupes concurrent lookups for
// the same ref so a log refresh doesn't fan out duplicate requests.
//
// Imports @/api/cnc lazily — keeping the top-level import-graph free
// of the auth/i18n chain lets `extractCodeRefs` be unit-tested
// without a DOM environment.
export async function resolveCodeRef(ref: CodeRef): Promise<CodeEntry | null> {
  const key = cacheKey(ref.kind, ref.number);
  if (cache.has(key)) return cache.get(key) ?? null;
  const existing = inFlight.get(key);
  if (existing) return existing;
  const p = (async () => {
    try {
      const { lookupCode } = await import("@/api/cnc");
      const r = await lookupCode(ref.kind, ref.number);
      const entry = r.ok ? (r.entry as CodeEntry) : null;
      // Trim cache when over cap. Map preserves insertion order so the
      // oldest entries go first. Simple and good enough for our scale.
      if (cache.size >= CACHE_CAP) {
        const firstKey = cache.keys().next().value;
        if (firstKey !== undefined) cache.delete(firstKey);
      }
      cache.set(key, entry);
      return entry;
    } catch {
      // Network / 4xx failure — return null but DON'T cache the
      // failure (the user might fix a transient issue and we want a
      // retry to succeed).
      return null;
    } finally {
      inFlight.delete(key);
    }
  })();
  inFlight.set(key, p);
  return p;
}

// resolveAll resolves an array of refs in parallel. Returns the same
// length array with null for unknown codes.
export function resolveAllCodeRefs(
  refs: CodeRef[]
): Promise<(CodeEntry | null)[]> {
  return Promise.all(refs.map(resolveCodeRef));
}

// Hard reset for tests.
export function _resetCodeRefCache(): void {
  cache.clear();
  inFlight.clear();
}
