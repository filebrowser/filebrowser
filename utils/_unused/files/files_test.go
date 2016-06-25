package files

import "testing"

func TestCopyFile(t *testing.T) {
	err := CopyFile("test_data/file_to_copy.txt", "test_data/copied_file.txt")

	if err != nil {
		t.Error("Can't copy the file.")
	}
}
