# CLAUDE.md

Guidance for Claude when working in this repository (File Browser, `filebrowser/filebrowser`).

## Repo orientation

- Go backend (`github.com/filebrowser/filebrowser/v2`, ecosystem `go`) + Vue frontend under `frontend/`.
- The project is in **maintenance-only mode** (see `SECURITY.md`). Prefer small, surgical, well-tested changes.
- Version scheme: `v2.63.x`. Conventional-commit messages (`fix(scope): …`, `feat: …`, `chore: …`).
- Verify with `go build ./...`, `go vet ./...`, `go test ./...`. Reuse existing test harnesses (e.g. `signToken`, `scopedUserStorage`, `handle`, `customFSUser`, `mockUserStore`).

---

# Handling security advisories

Use this playbook when asked to triage, verify, fix, or manage GitHub security advisories.
All advisory state lives on GitHub and is driven with the `gh` CLI. States are
`triage → draft → published`, plus `closed`.

## 1. Fetch

```bash
# List by state (also: published, draft, closed)
gh api '/repos/filebrowser/filebrowser/security-advisories?state=triage&per_page=100' \
  --jq '.[] | {ghsa_id, severity, summary, state}'

# Full report for one advisory (read .summary and .description)
gh api /repos/filebrowser/filebrowser/security-advisories/GHSA-xxxx-xxxx-xxxx \
  --jq '.summary, "---", .description'
```

Always pull the **published** set too — you need it to dedup against.

## 2. Verify each report — do NOT trust the report text

Read the actual source **at HEAD** and reach one verdict per advisory:

- **CONFIRMED** — defect exists at HEAD. Quote the exact `file:line`.
- **FIXED** — already patched (find the fix commit).
- **FALSE / NOT APPLICABLE** — claim is wrong, or targets a different project.
- **NOT EXPLOITABLE** — pattern exists but no code path can reach the precondition.
- **DUPLICATE** — of a published advisory, or of another triage advisory.

Common traps — check each before accepting a report:

- **"Incomplete fix of a prior advisory."** Read the original fix commit and confirm the *specific*
  sibling code path is actually still unguarded — a prior fix may already cover a related path.
- **Wrong project.** Confirm the referenced files, symbols, and endpoints exist in this repo; reports
  sometimes describe a fork. If the cited code isn't here, close as not applicable.
- **"Legacy / upgraded / imported records are affected."** Trace the field's git history and confirm that
  some released version could actually produce such a record before believing it
  (`git log -S'Field' -- path`, `git show <commit>`, `git tag --contains <commit>`).
- **Overlapping reports.** Two triage advisories may share one root cause — consolidate and fix once.
- **Known, intentionally-unaddressed classes.** Some areas are known and tracked but not fixed (see
  `SECURITY.md`'s Known Issues); matching reports are duplicates.

Record, per advisory: verdict, the `file:line` evidence, exploitability preconditions
(default config? platform-specific? auth required?), and the disposition.

## 3. Fix (one commit per advisory)

- Branch off `master` first; never commit fixes directly to `master`.
- **One commit per fix**, each referencing the GHSA in the body (`Refs GHSA-xxxx-xxxx-xxxx`).
- When several advisories share a root cause because parallel code paths diverged, **centralize** the logic
  (e.g. `settings.CreateUserHome` used by signup, proxy, and hook provisioning) so they cannot drift again.
- Keep it surgical; match surrounding style.

## 4. Add a regression test per fix

- Reuse the existing harnesses; add a focused test asserting the fixed behavior (and that the legitimate
  path still works). Commit tests alongside the fixes (`test(scope): …`, `Refs GHSA-…`).
- Run `go test ./... && go vet ./...` before finishing.

## 5. Severity — set a CVSS vector, don't hand-assert

Set a CVSS v3.1 vector; GitHub derives the score **and** severity from it (overriding the plain `severity`
field). Encode the real preconditions in the vector so the band is defensible:

- Requires a specific target configuration/platform (e.g. case-insensitive FS) → **AC:H**.
- Unauthenticated vs needs an account → **PR:N / PR:L**.
- Keep it consistent with the **predecessor CVE's** rating (an incomplete-fix follow-up should not outrank
  its parent). Bands: `0.1–3.9` low, `4.0–6.9` medium, `7.0–8.9` high, `9.0–10.0` critical.

```bash
gh api -X PATCH /repos/filebrowser/filebrowser/security-advisories/GHSA-xxxx-xxxx-xxxx \
  -f cvss_vector_string='CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H' \
  --jq '{ghsa_id, severity, score: .cvss.score, vector: .cvss.vector_string}'
```

## 6. Clean up the title

Rewrite reporter titles to the concise, sentence-case, backtick-free style of the published advisories
(state the vuln; drop "Incomplete fix of…" prefixes and jargon). The title is the `summary` field:

```bash
gh api -X PATCH .../security-advisories/GHSA-xxxx-xxxx-xxxx \
  -f summary='Proxy-auth auto-provisioning ignores createUserDir and grants the server root scope'
```

## 7. Rewrite the body into the standard structure

Reporter reports arrive in whatever shape the reporter used. Before drafting, rewrite the
`description` into the sections below — **reusing the reporter's own wording wherever it is
accurate**, rather than paraphrasing it. Drop the greeting, the offer to help, and any claim the
verification in step 2 disproved. Keep sections in this order and omit the ones that don't apply:

| Section | Contents |
| --- | --- |
| `## Summary` | What the defect is, in two or three sentences. Note the version it was reported against and the range it was verified over. |
| `## Details` | Root cause, naming `file.go`, the function, and a short quote of the **pre-fix** code. |
| `## PoC` | The reporter's reproduction steps and observed result, trimmed to the essentials. |
| `## Impact` | Who can exploit it (privilege level, preconditions) and what they get. |
| `## Patches` | Fixed version plus a link to the fix commit, and one sentence on what the fix does. Mention it if the reporter re-tested and confirmed. |
| `## Workarounds` | Real mitigations, or `None. Upgrade to vX.Y.Z.` |
| `## Out of scope` | Anything in the original report deliberately **not** treated as a vulnerability, with the reasoning. Needed whenever the advisory is narrower than the report. |
| `## References` | Fix and regression-test commit links. |

Use `##` headings (matching the published advisories) and keep the body in the maintainer's voice —
first person ("I reproduced…") belongs only inside quoted PoC steps.

Send it as a file so the Markdown survives shell quoting:

```bash
jq -Rs '{description: .}' desc.md \
  | gh api -X PATCH .../security-advisories/GHSA-xxxx-xxxx-xxxx --input -
```

## 8. Set affected & patched versions

The package is always `{ecosystem: "go", name: "github.com/filebrowser/filebrowser/v2"}`.

- `vulnerable_version_range` — `<= <latest released version>` (the newest tag; find it with
  `git tag --list 'v2.*' --sort=-version:refname | head -1`).
- `patched_versions` — the **next** release that will actually ship the fix. It doesn't exist just
  because the branch does; it still has to be cut. **When unsure which version that will be, ask.**

Replace the placeholder versions below before sending:

```bash
printf '%s' '{"vulnerabilities":[{"package":{"ecosystem":"go","name":"github.com/filebrowser/filebrowser/v2"},"vulnerable_version_range":"<= <latest>","patched_versions":"<next>","vulnerable_functions":[]}]}' \
  | gh api -X PATCH .../security-advisories/GHSA-xxxx-xxxx-xxxx --input -
```

## 9. Move state / close, by disposition

```bash
gh api -X PATCH .../security-advisories/GHSA-xxxx-xxxx-xxxx -f state=draft    # fixed, awaiting release
gh api -X PATCH .../security-advisories/GHSA-xxxx-xxxx-xxxx -f state=closed   # duplicate / N-A / not-exploitable
```

- **CONFIRMED & fixed** → metadata done → **draft**. Publish after the patched release ships.
- **DUPLICATE / NOT APPLICABLE / NOT EXPLOITABLE** → **closed**.
- **DEFERRED** (confirmed but not yet fixed) → leave in **triage** with a note.
- The REST API **cannot post advisory comments** — replies to reporters (dup notice, evidence, links to
  the relevant tracking issue) must be posted manually in the advisory UI. Draft the text for the maintainer.

## 10. Release & publish

Push the branch, open a PR, merge, tag/release the `patched_versions` you set, then publish the drafts.
Confirm outward-facing/irreversible advisory actions (close, publish) with the maintainer before doing them.
