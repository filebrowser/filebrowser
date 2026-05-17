package cnc

// Chapter extraction — scans an NC file for operation-header comments
// so the operator can jump around a long program like chapters in a
// book. CAM postprocessors typically emit one comment line per
// operation ("2D CONTOUR1", "DRILL D=0.25", "ROUGH XY") interleaved
// with the actual G-code; this picks those out so the dashboard can
// render a TOC.
//
// Heuristic, not perfect:
//   - Pull every `( … )` or `; …` line that stands alone on its own
//     block (no G-code on the same line).
//   - Drop lines from the file's opening metadata block (everything
//     before the first non-comment line). Those carry post info,
//     date/time, units, the tool list — operator already sees the
//     tool list on the Tools tab.
//   - Drop the standard metadata words ("DATE", "TIME", "T<n>" tool
//     headers, "POSTPROCESSOR", "GENERATED").
//   - Keep the rest. Even if a comment isn't strictly a "chapter,"
//     it's still useful jump-bait.

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Chapter is one entry in the TOC. Line is 1-based to match the
// existing GcodeFollow.machineLine convention.
type Chapter struct {
	Line    int    `json:"line"`
	Comment string `json:"comment"`
}

// ChapterList is the response shape from GET /api/cnc/chapters.
type ChapterList struct {
	FilePath string    `json:"file_path"`
	Total    int       `json:"total"`
	Chapters []Chapter `json:"chapters"`
}

// Lines like "(T5 D=0.5 ... — bullnose)" or "(T5)" are tool-list
// headers from the preamble. Drop them — they're already on the Tools
// tab and would otherwise drown out real operation headers.
var chapterToolHeaderRe = regexp.MustCompile(`(?i)^\s*T\d{1,3}\b`)

// metadata words that show up in CAM-emitted file preamble. Lines
// matching are skipped even when they appear mid-program (Fusion
// occasionally repeats DATE / TIME inside).
var chapterMetadataRe = regexp.MustCompile(`(?i)^\s*(date|time|generated|postprocessor|units|t\d+|stock|wcs|file|customer|operator|partid|part id|toolpath)\b`)

// Inline comments — keep "( ... )" and "; ..." groups. We only count a
// comment as a chapter when it stands alone (no G-code follows on the
// same line); inline annotations next to a real move would be noise.
var chapterCommentRe = regexp.MustCompile(`^\s*\((.*)\)\s*$|^\s*;\s*(.*)$`)

// commentInGcodeRe matches a line that has G-code AND a trailing
// comment — we explicitly do NOT extract these as chapters.
var hasGCodeRe = regexp.MustCompile(`(?i)\b([GMSTNF]\d+|X-?\d+|Y-?\d+|Z-?\d+|I-?\d+|J-?\d+|K-?\d+|R-?\d+)\b`)

// BuildChapters scans the file at absPath and returns a Chapter list.
// Never errors on parse — malformed lines just produce a shorter
// list. Errors only when the file can't be opened.
func BuildChapters(absPath, displayPath string) (*ChapterList, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", absPath, err)
	}
	defer f.Close()

	out := &ChapterList{FilePath: displayPath, Chapters: []Chapter{}}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	lineNum := 0
	seenGcode := false // flips true after first non-comment line
	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		isComment := chapterCommentRe.MatchString(raw)
		if !isComment {
			// G-code line — flag that we've crossed into the body and
			// skip. The preamble is everything BEFORE the first non-
			// comment line, so once we trip seenGcode we accept further
			// stand-alone comments as chapters.
			if hasGCodeRe.MatchString(raw) {
				seenGcode = true
			}
			continue
		}

		// Pre-body comments are metadata — operator sees this stuff
		// elsewhere (Tools tab, file header). Skip until we've left
		// the preamble.
		if !seenGcode {
			continue
		}

		body := extractCommentBody(raw)
		if body == "" {
			continue
		}
		if chapterToolHeaderRe.MatchString(body) {
			continue
		}
		if chapterMetadataRe.MatchString(body) {
			continue
		}
		out.Chapters = append(out.Chapters, Chapter{
			Line:    lineNum,
			Comment: body,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", absPath, err)
	}
	out.Total = len(out.Chapters)
	return out, nil
}

// extractCommentBody pulls the text out of a `( ... )` or `; ...` line
// and tidies whitespace. Returns "" when the comment is empty.
func extractCommentBody(line string) string {
	m := chapterCommentRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	for _, g := range m[1:] {
		if g != "" {
			return strings.TrimSpace(g)
		}
	}
	return ""
}

// ChapterAt returns the chapter that contains line `at` (1-based) —
// i.e. the chapter with the highest Line value that is <= at. Used to
// surface the "current operation" indicator from cnc.lineCurrent.
// Returns nil when no chapter precedes `at` (operator is still in the
// preamble) or the list is empty.
func ChapterAt(chapters []Chapter, at int) *Chapter {
	if len(chapters) == 0 || at <= 0 {
		return nil
	}
	// Binary search — chapters are emitted in file order so the slice
	// is already sorted by Line ascending.
	lo, hi := 0, len(chapters)-1
	var match *Chapter
	for lo <= hi {
		mid := (lo + hi) / 2
		if chapters[mid].Line <= at {
			match = &chapters[mid]
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return match
}
