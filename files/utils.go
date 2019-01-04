package files

func isBinary(content string) bool {
	for _, b := range content {
		// 65533 is the unknown char
		// 8 and below are control chars (e.g. backspace, null, eof, etc)
		if b <= 8 || b == 65533 {
			return true
		}
	}
	return false
}
