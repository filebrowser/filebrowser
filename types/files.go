package types

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// File describes a file.
type File struct {
	*Listing
	user      *User
	Path      string            `json:"path"`
	Name      string            `json:"name"`
	Size      int64             `json:"size"`
	Extension string            `json:"extension"`
	ModTime   time.Time         `json:"modified"`
	Mode      os.FileMode       `json:"mode"`
	IsDir     bool              `json:"isDir"`
	Type      string            `json:"type"`
	Subtitles []string          `json:"subtitles,omitempty"`
	Content   string            `json:"content,omitempty"`
	Checksums map[string]string `json:"checksums,omitempty"`
}

// NewFile generates a new file info from a user and a path.
func NewFile(u *User, path string) (*File, error) {
	f := &File{
		Path: path,
	}

	i, err := u.Fs.Stat(path)
	if err != nil {
		return f, err
	}

	f.user = u
	f.Name = i.Name()
	f.ModTime = i.ModTime()
	f.Mode = i.Mode()
	f.IsDir = i.IsDir()
	f.Size = i.Size()
	f.Extension = filepath.Ext(f.Name)

	if f.IsDir {
		err = f.getDirInfo()
	} else {
		err = f.detectFileType()
	}

	return f, err
}

// Checksum retrieves the checksum of a file.
func (f *File) Checksum(algo string) error {
	if f.IsDir {
		return ErrIsDirectory
	}

	if f.Checksums == nil {
		f.Checksums = map[string]string{}
	}

	i, err := f.user.Fs.Open(f.Path)
	if err != nil {
		return err
	}
	defer i.Close()

	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return ErrInvalidOption
	}

	_, err = io.Copy(h, i)
	if err != nil {
		return err
	}

	f.Checksums[algo] = hex.EncodeToString(h.Sum(nil))
	return nil
}

func (f *File) getDirInfo() error {
	afs := &afero.Afero{Fs: f.user.Fs}
	files, err := afs.ReadDir(f.Path)
	if err != nil {
		return err
	}

	f.Listing = &Listing{
		Items:    []*File{},
		NumDirs:  0,
		NumFiles: 0,
	}

	for _, i := range files {
		name := i.Name()
		path := filepath.Join(f.Path, name)

		if strings.HasPrefix(i.Mode().String(), "L") {
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead if the target's.
			info, err := os.Stat(name)
			if err == nil {
				i = info
			}
		}

		file := &File{
			user:      f.user,
			Name:      name,
			Size:      i.Size(),
			ModTime:   i.ModTime(),
			Mode:      i.Mode(),
			IsDir:     i.IsDir(),
			Extension: filepath.Ext(name),
			Path:      path,
		}

		if file.IsDir {
			f.Listing.NumDirs++
		} else {
			f.Listing.NumFiles++

			err := file.detectFileType()
			if err != nil {
				return err
			}
		}

		f.Listing.Items = append(f.Listing.Items, file)
	}

	return nil
}

func (f *File) detectFileType() error {
	i, err := f.user.Fs.Open(f.Path)
	if err != nil {
		return err
	}
	defer i.Close()

	buffer := make([]byte, 512)
	n, err := i.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	mimetype := mime.TypeByExtension(f.Extension)
	if mimetype == "" {
		mimetype = http.DetectContentType(buffer[:n])
	}

	switch {
	case strings.HasPrefix(mimetype, "video"):
		f.Type = "video"
		return nil
	case strings.HasPrefix(mimetype, "audio"):
		f.Type = "audio"
		return nil
	case strings.HasPrefix(mimetype, "image"):
		f.Type = "image"
		return nil
	case isBinary(string(buffer[:n])) || f.Size > 10*1024*1024: // 10 MB
		f.Type = "blob"
		return nil
	default:
		f.Type = "text"
		afs := &afero.Afero{Fs: f.user.Fs}
		content, err := afs.ReadFile(f.Path)
		if err != nil {
			return err
		}
		f.Content = string(content)
	}

	return nil
}

var (
	subtitleExts = []string{
		".vtt",
	}
)

// DetectSubtitles fills the subtitles field if the file
// is a movie.
func (f *File) DetectSubtitles() {
	f.Subtitles = []string{}
	ext := filepath.Ext(f.Path)
	base := strings.TrimSuffix(f.Path, ext)

	for _, ext := range subtitleExts {
		path := base + ext
		if _, err := f.user.Fs.Stat(path); err == nil {
			f.Subtitles = append(f.Subtitles, path)
		}
	}
}

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
