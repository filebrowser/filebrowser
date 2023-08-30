package files

import (
	"testing"

	"github.com/spf13/afero"
)

func TestDetectSubtitles(t *testing.T) {
	testCases := []struct {
		path string
		want bool // is file detected as subtitles?
	}{
		{path: "/media/movie.mkv", want: false},
		{path: "/media/movie.vtt", want: true},
		{path: "/media/movie.en.srt", want: true},
		{path: "/media/Subs/movie.pt.srt", want: true},
		{path: "/media/subs/movie.zh-tw.srt", want: true},
		{path: "/media/movie.es-es.srt", want: true},
		{path: "/media/movie.fr.srt", want: true},
		{path: "/media/srt", want: false},
		{path: "/media/movie.dir.vtt", want: false},
		{path: "/media/subs/movie.dir.srt", want: false},
	}

	fs := afero.NewMemMapFs()
	err0 := fs.MkdirAll("/media", PermDir)
	if err0 != nil {
		t.Fatalf("Failed to create directory: %v", err0)
	}
	err1 := fs.MkdirAll("/media/movie.dir.vtt", PermDir)
	if err1 != nil {
		t.Fatalf("Failed to create directory: %v", err1)
	}
	err2 := fs.MkdirAll("/media/subs/movie.dir.srt", PermDir)
	if err2 != nil {
		t.Fatalf("Failed to create directory: %v", err2)
	}

	for _, path := range testCases {
		err := afero.WriteFile(fs, path.path, []byte("data"), PermFile)
		if err != nil {
			return
		}
	}

	file := &FileInfo{
		Fs:        fs,
		Path:      "/media/movie.mkv",
		Name:      "movie.mkv",
		Type:      "video",
		Size:      42,
		Extension: ".mkv",
	}

	file.detectSubtitles()

	for _, tt := range testCases {
		t.Run(tt.path, func(t *testing.T) {
			if got := contains(file.Subtitles, tt.path); got != tt.want {
				t.Errorf("detectSubtitles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
