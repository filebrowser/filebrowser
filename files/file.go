package files

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
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

// FileInfo describes a file.
type FileInfo struct {
	*Listing
	Fs        afero.Fs          `json:"-"`
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

// FileOptions are the options when getting a file info.
type FileOptions struct {
	Fs      afero.Fs
	Path    string
	Modify  bool
	Expand  bool
	Checker rules.Checker
}

// NewFileInfo creates a File object from a path and a given user. This File
// object will be automatically filled depending on if it is a directory
// or a file. If it's a video file, it will also detect any subtitles.
func NewFileInfo(opts FileOptions) (*FileInfo, error) {
	if !opts.Checker.Check(opts.Path) {
		return nil, os.ErrPermission
	}

	info, err := opts.Fs.Stat(opts.Path)
	if err != nil {
		return nil, err
	}

	file := &FileInfo{
		Fs:        opts.Fs,
		Path:      opts.Path,
		Name:      info.Name(),
		ModTime:   info.ModTime(),
		Mode:      info.Mode(),
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		Extension: filepath.Ext(info.Name()),
	}

	if opts.Expand {
		if file.IsDir {
			return file, file.readListing(opts.Checker)
		}

		err = file.detectType(opts.Modify)
		if err != nil {
			return nil, err
		}
	}

	return file, err
}

// Checksum checksums a given File for a given User, using a specific
// algorithm. The checksums data is saved on File object.
func (i *FileInfo) Checksum(algo string) error {
	if i.IsDir {
		return errors.ErrIsDirectory
	}

	if i.Checksums == nil {
		i.Checksums = map[string]string{}
	}

	reader, err := i.Fs.Open(i.Path)
	if err != nil {
		return err
	}
	defer reader.Close()

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
		return errors.ErrInvalidOption
	}

	_, err = io.Copy(h, reader)
	if err != nil {
		return err
	}

	i.Checksums[algo] = hex.EncodeToString(h.Sum(nil))
	return nil
}

func (i *FileInfo) detectType(modify bool) error {
	reader, err := i.Fs.Open(i.Path)
	if err != nil {
		return err
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	mimetype := mime.TypeByExtension(i.Extension)
	if mimetype == "" {
		mimetype = http.DetectContentType(buffer[:n])
	}

	switch {
	case strings.HasPrefix(mimetype, "video"):
		i.Type = "video"
		i.detectSubtitles()
		return nil
	case strings.HasPrefix(mimetype, "audio"):
		i.Type = "audio"
		return nil
	case strings.HasPrefix(mimetype, "image"):
		i.Type = "image"
		return nil
	case isBinary(string(buffer[:n])) || i.Size > 10*1024*1024: // 10 MB
		i.Type = "blob"
		return nil
	default:
		i.Type = "text"
		afs := &afero.Afero{Fs: i.Fs}
		content, err := afs.ReadFile(i.Path)
		if err != nil {
			return err
		}

		if !modify {
			i.Type = "textImmutable"
		}

		i.Content = string(content)
	}

	return nil
}

func (i *FileInfo) detectSubtitles() {
	if i.Type != "video" {
		return
	}

	i.Subtitles = []string{}
	ext := filepath.Ext(i.Path)

	// TODO: detect multiple languages. Base.Lang.vtt

	path := strings.TrimSuffix(i.Path, ext) + ".vtt"
	if _, err := i.Fs.Stat(path); err == nil {
		i.Subtitles = append(i.Subtitles, path)
	}
}

func (i *FileInfo) readListing(checker rules.Checker) error {
	afs := &afero.Afero{Fs: i.Fs}
	dir, err := afs.ReadDir(i.Path)
	if err != nil {
		return err
	}

	listing := &Listing{
		Items:    []*FileInfo{},
		NumDirs:  0,
		NumFiles: 0,
	}

	for _, f := range dir {
		name := f.Name()
		path := path.Join(i.Path, name)

		if !checker.Check(path) {
			continue
		}

		if strings.HasPrefix(f.Mode().String(), "L") {
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead if the target's.
			info, err := i.Fs.Stat(path)
			if err == nil {
				f = info
			}
		}

		file := &FileInfo{
			Fs:        i.Fs,
			Name:      name,
			Size:      f.Size(),
			ModTime:   f.ModTime(),
			Mode:      f.Mode(),
			IsDir:     f.IsDir(),
			Extension: filepath.Ext(name),
			Path:      path,
		}

		if file.IsDir {
			listing.NumDirs++
		} else {
			listing.NumFiles++

			err := file.detectType(true)
			if err != nil {
				return err
			}
		}

		listing.Items = append(listing.Items, file)
	}

	i.Listing = listing
	return nil
}
