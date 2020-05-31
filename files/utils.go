package files

import (
	"unicode/utf8"
)

func isBinary(content []byte, _ int) bool {
	maybeStr := string(content)
	runeCnt := utf8.RuneCount(content)
	runeIndex := 0
	gotRuneErrCnt := 0
	firstRuneErrIndex := -1

	const (
		// 8 and below are control chars (e.g. backspace, null, eof, etc)
		maxControlCharsCode = 8
		// 0xFFFD(65533) is  the "error" Rune or "Unicode replacement character"
		// see https://golang.org/pkg/unicode/utf8/#pkg-constants
		unicodeReplacementChar = 0xFFFD
	)

	for _, b := range maybeStr {
		if b <= maxControlCharsCode {
			return true
		}

		if b == unicodeReplacementChar {
			// if it is not the last (utf8.UTFMax - x) rune
			if runeCnt > utf8.UTFMax && runeIndex < runeCnt-utf8.UTFMax {
				return true
			}
			// else it is the last (utf8.UTFMax - x) rune
			// there maybe Vxxx, VVxx, VVVx, thus, we may got max 3 0xFFFD rune (assume V is the byte we got)
			// for Chinese, it can only be Vxx, VVx, we may got max 2 0xFFFD rune
			gotRuneErrCnt++

			// mark the first time
			if firstRuneErrIndex == -1 {
				firstRuneErrIndex = runeIndex
			}
		}
		runeIndex++
	}

	// if last (utf8.UTFMax - x ) rune has the "error" Rune, but not all
	if firstRuneErrIndex != -1 && gotRuneErrCnt != runeCnt-firstRuneErrIndex {
		return true
	}
	return false
}
