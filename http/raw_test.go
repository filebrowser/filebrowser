package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/filebrowser/filebrowser/v2/files"
)

func TestSetContentDisposition(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		filename string
		inline   bool
		expected string
	}{
		"inline simple filename": {
			filename: "document.pdf",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("document.pdf"),
		},
		"attachment simple filename": {
			filename: "document.pdf",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("document.pdf"),
		},
		"inline non-ASCII filename": {
			filename: "日本語.txt",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("日本語.txt"),
		},
		"attachment non-ASCII filename": {
			filename: "日本語.txt",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("日本語.txt"),
		},
		"inline filename with spaces": {
			filename: "my file.txt",
			inline:   true,
			expected: "inline; filename*=utf-8''" + url.PathEscape("my file.txt"),
		},
		"attachment filename with spaces": {
			filename: "my file.txt",
			inline:   false,
			expected: "attachment; filename*=utf-8''" + url.PathEscape("my file.txt"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "/test", http.NoBody)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			if tc.inline {
				req.URL.RawQuery = "inline=true"
			}

			file := &files.FileInfo{Name: tc.filename}

			setContentDisposition(recorder, req, file)

			got := recorder.Header().Get("Content-Disposition")
			if got != tc.expected {
				t.Errorf("Content-Disposition = %q, want %q", got, tc.expected)
			}
		})
	}
}
