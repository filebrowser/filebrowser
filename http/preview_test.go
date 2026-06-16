package fbhttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/files"
)

type fakeVideoThumbService struct {
	output []byte
	err    error
	input  string
	calls  int
}

func (s *fakeVideoThumbService) Thumbnail(_ context.Context, input string, out io.Writer) error {
	s.calls++
	s.input = input
	if s.err != nil {
		return s.err
	}
	_, err := out.Write(s.output)
	return err
}

type memoryFileCache struct {
	values map[string][]byte
}

func (c *memoryFileCache) Store(_ context.Context, key string, value []byte) error {
	c.values[key] = value
	return nil
}

func (c *memoryFileCache) Load(_ context.Context, key string) ([]byte, bool, error) {
	value, ok := c.values[key]
	return value, ok, nil
}

func (c *memoryFileCache) Delete(_ context.Context, key string) error {
	delete(c.values, key)
	return nil
}

func TestHandleVideoPreviewCreatesThumbnail(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/preview/thumb/movie.mp4", nil)
	rec := httptest.NewRecorder()
	cache := &memoryFileCache{values: map[string][]byte{}}
	svc := &fakeVideoThumbService{output: []byte("jpeg")}
	file := &files.FileInfo{
		Fs:      afero.NewMemMapFs(),
		Path:    "/movie.mp4",
		Name:    "movie.mp4",
		ModTime: time.Now(),
	}

	status, err := handleVideoPreview(rec, req, svc, cache, file, PreviewSizeThumb, true)
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf("expected handled status 0, got %d", status)
	}
	if svc.calls != 1 {
		t.Fatalf("expected thumbnail service to be called once, got %d", svc.calls)
	}
	if rec.Body.String() != "jpeg" {
		t.Fatalf("expected thumbnail body, got %q", rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("expected image/jpeg content type, got %q", got)
	}
}

func TestHandleVideoPreviewUsesCache(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/preview/thumb/movie.mp4", nil)
	rec := httptest.NewRecorder()
	file := &files.FileInfo{
		Fs:      afero.NewMemMapFs(),
		Path:    "/movie.mp4",
		Name:    "movie.mp4",
		ModTime: time.Now(),
	}
	cache := &memoryFileCache{
		values: map[string][]byte{
			previewCacheKey(file, PreviewSizeThumb): []byte("cached"),
		},
	}
	svc := &fakeVideoThumbService{output: []byte("generated")}

	status, err := handleVideoPreview(rec, req, svc, cache, file, PreviewSizeThumb, true)
	if err != nil {
		t.Fatal(err)
	}
	if status != 0 {
		t.Fatalf("expected handled status 0, got %d", status)
	}
	if svc.calls != 0 {
		t.Fatalf("expected thumbnail service not to be called, got %d", svc.calls)
	}
	if rec.Body.String() != "cached" {
		t.Fatalf("expected cached thumbnail body, got %q", rec.Body.String())
	}
}

func TestHandleVideoPreviewReportsUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/preview/thumb/movie.mp4", nil)
	rec := httptest.NewRecorder()
	cache := &memoryFileCache{values: map[string][]byte{}}
	svc := &fakeVideoThumbService{err: errors.New("missing ffmpeg")}
	file := &files.FileInfo{
		Fs:      afero.NewMemMapFs(),
		Path:    "/movie.mp4",
		Name:    "movie.mp4",
		ModTime: time.Now(),
	}

	status, err := handleVideoPreview(rec, req, svc, cache, file, PreviewSizeThumb, true)
	if err == nil {
		t.Fatal("expected error")
	}
	if status != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", status)
	}
}
