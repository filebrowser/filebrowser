package files

import (
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"image"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/spf13/afero"
)

const PermFile = 0664
const PermDir = 0755

// FileInfo describes a file.
type FileInfo struct {
	*Listing
	Fs         afero.Fs          `json:"-"`
	Path       string            `json:"path"`
	Name       string            `json:"name"`
	Size       int64             `json:"size"`
	Extension  string            `json:"extension"`
	ModTime    time.Time         `json:"modified"`
	Mode       os.FileMode       `json:"mode"`
	IsDir      bool              `json:"isDir"`
	IsSymlink  bool              `json:"isSymlink"`
	Type       string            `json:"type"`
	Subtitles  []string          `json:"subtitles,omitempty"`
	Content    string            `json:"content,omitempty"`
	Checksums  map[string]string `json:"checksums,omitempty"`
	Token      string            `json:"token,omitempty"`
	currentDir []os.FileInfo     `json:"-"`
	Resolution *ImageResolution  `json:"resolution,omitempty"`
}

// FileOptions are the options when getting a file info.
type FileOptions struct {
	Fs          afero.Fs
	Path        string
	Modify      bool
	Expand      bool
	ReadHeader  bool
	FolderSize  bool
	Token       string
	Checker     rules.Checker
	Content     bool
	RootPath    string
	AnotherPath string
}

type ImageResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// function that checks if given full path exists or not
// it helps to determine to which NFS we need to go IDC or KFS
func CheckIfExistsInPath(pathToCheck string) bool {
	_, pathExistsErr := os.Stat(pathToCheck)
	return !os.IsNotExist(pathExistsErr)
}

// NewFileInfo creates a File object from a path and a given user. This File
// object will be automatically filled depending on if it is a directory
// or a file. If it's a video file, it will also detect any subtitles.
func NewFileInfo(opts FileOptions) (*FileInfo, error) {
	log.Printf("ROOT PATH - %v:", opts.RootPath)
	log.Printf("ANOTHER PATH - %v:", opts.AnotherPath)
	rootFilePath := opts.RootPath + opts.Path
	if !CheckIfExistsInPath(rootFilePath) {
		opts.Fs = afero.NewBasePathFs(afero.NewOsFs(), opts.AnotherPath)
	} else {
		opts.Fs = afero.NewBasePathFs(afero.NewOsFs(), opts.RootPath)
	}
	if !opts.Checker.Check(opts.Path) {
		return nil, os.ErrPermission
	}

	file, err := stat(opts)
	if err != nil {
		return nil, err
	}
	if file.IsDir && opts.FolderSize {
		size, err := getFolderSize(file.RealPath())
		if err != nil {
			return nil, err
		}
		file.Size = size
	}
	if opts.Expand {
		if file.IsDir {
			if err := file.readListing(opts.Checker, opts.ReadHeader, opts.RootPath, opts.AnotherPath); err != nil { //nolint:govet
				return nil, err
			}
			return file, nil
		}

		err = file.detectType(opts.Modify, opts.Content, true)
		if err != nil {
			return nil, err
		}
	}

	return file, err
}

func getFolderSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return err
	})
	return size, err
}

func stat(opts FileOptions) (*FileInfo, error) {
	var file *FileInfo

	if lstaterFs, ok := opts.Fs.(afero.Lstater); ok {
		info, _, err := lstaterFs.LstatIfPossible(opts.Path)
		if err != nil {
			log.Printf("stat current path error - %v:", err)
			return nil, err
		}
		file = &FileInfo{
			Fs:        opts.Fs,
			Path:      opts.Path,
			Name:      info.Name(),
			ModTime:   info.ModTime(),
			Mode:      info.Mode(),
			IsDir:     info.IsDir(),
			IsSymlink: IsSymlink(info.Mode()),
			Size:      info.Size(),
			Extension: filepath.Ext(info.Name()),
			Token:     opts.Token,
		}
	}

	// regular file
	if file != nil && !file.IsSymlink {
		return file, nil
	}

	// fs doesn't support afero.Lstater interface or the file is a symlink
	info, err := opts.Fs.Stat(opts.Path)
	if err != nil {
		// can't follow symlink
		if file != nil && file.IsSymlink {
			return file, nil
		}
		return nil, err
	}

	// set correct file size in case of symlink
	if file != nil && file.IsSymlink {
		file.Size = info.Size()
		file.IsDir = info.IsDir()
		return file, nil
	}

	file = &FileInfo{
		Fs:        opts.Fs,
		Path:      opts.Path,
		Name:      info.Name(),
		ModTime:   info.ModTime(),
		Mode:      info.Mode(),
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		Extension: filepath.Ext(info.Name()),
		Token:     opts.Token,
	}

	return file, nil
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

	//nolint:gosec
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

func (i *FileInfo) RealPath() string {
	if realPathFs, ok := i.Fs.(interface {
		RealPath(name string) (fPath string, err error)
	}); ok {
		realPath, err := realPathFs.RealPath(i.Path)
		if err == nil {
			return realPath
		}
	}

	return i.Path
}

// TODO: use constants
//
//nolint:goconst
func (i *FileInfo) detectType(modify, saveContent, readHeader bool) error {
	if IsNamedPipe(i.Mode) {
		i.Type = "blob"
		return nil
	}
	// failing to detect the type should not return error.
	// imagine the situation where a file in a dir with thousands
	// of files couldn't be opened: we'd have immediately
	// a 500 even though it doesn't matter. So we just log it.

	mimetype := mime.TypeByExtension(i.Extension)

	var buffer []byte
	if readHeader {
		buffer = i.readFirstBytes()

		if mimetype == "" {
			mimetype = http.DetectContentType(buffer)
		}
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
		resolution, err := calculateImageResolution(i.Fs, i.Path)
		if err != nil {
			log.Printf("Error calculating image resolution: %v", err)
		} else {
			i.Resolution = resolution
		}
		return nil
	case strings.HasSuffix(mimetype, "pdf"):
		i.Type = "pdf"
		return nil
	case (strings.HasPrefix(mimetype, "text") || !isBinary(buffer)) && i.Size <= 10*1024*1024: // 10 MB
		i.Type = "text"

		if !modify {
			i.Type = "textImmutable"
		}

		if saveContent {
			afs := &afero.Afero{Fs: i.Fs}
			content, err := afs.ReadFile(i.Path)
			if err != nil {
				return err
			}

			i.Content = string(content)
		}
		return nil
	default:
		i.Type = "blob"
	}

	return nil
}

func calculateImageResolution(fs afero.Fs, filePath string) (*ImageResolution, error) {
	file, err := fs.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			log.Printf("Failed to close file: %v", cErr)
		}
	}()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	return &ImageResolution{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func (i *FileInfo) readFirstBytes() []byte {
	reader, err := i.Fs.Open(i.Path)
	if err != nil {
		log.Print(err)
		i.Type = "blob"
		return nil
	}
	defer reader.Close()

	buffer := make([]byte, 512) //nolint:gomnd
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		log.Print(err)
		i.Type = "blob"
		return nil
	}

	return buffer[:n]
}

func (i *FileInfo) detectSubtitles() {
	if i.Type != "video" {
		return
	}

	i.Subtitles = []string{}
	ext := filepath.Ext(i.Path)

	// detect multiple languages. Base*.vtt
	// TODO: give subtitles descriptive names (lang) and track attributes
	parentDir := strings.TrimRight(i.Path, i.Name)
	var dir []os.FileInfo
	if len(i.currentDir) > 0 {
		dir = i.currentDir
	} else {
		var err error
		dir, err = afero.ReadDir(i.Fs, parentDir)
		if err != nil {
			return
		}
	}

	base := strings.TrimSuffix(i.Name, ext)
	for _, f := range dir {
		if !f.IsDir() && strings.HasPrefix(f.Name(), base) && strings.HasSuffix(f.Name(), ".vtt") {
			i.Subtitles = append(i.Subtitles, path.Join(parentDir, f.Name()))
		}
	}
}

// async read dir and append the data to given FileInfo slice
func readDirAsync(fs afero.Fs, fullPath string, wg *sync.WaitGroup, resultSlice *[]os.FileInfo) {
	defer wg.Done()

	dir, err := afero.ReadDir(fs, fullPath)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return
	}

	// Append the result to the slice
	*resultSlice = append(*resultSlice, dir...)
}

func (i *FileInfo) readListing(checker rules.Checker, readHeader bool, rootPath, anotherPath string) error {
	var wg sync.WaitGroup
	var rootDir []os.FileInfo
	var anotherDir []os.FileInfo
	var finalDir []os.FileInfo
	useAnotherDir := false
	anotherFullPath := anotherPath + i.Path
	rootFullPath := rootPath + i.Path
	existsInRootPath := CheckIfExistsInPath(rootFullPath)
	existsInAnotherPath := CheckIfExistsInPath(anotherFullPath)
	log.Printf("%v %v %v %v", anotherFullPath, existsInAnotherPath, rootFullPath, existsInRootPath)
	// if we aren't in home scenario use idcSite only because we are messing with opth.Path
	// in some cases it can go to KFS site instead of going to IDC, so just to be sure...
	if existsInRootPath && existsInAnotherPath && i.Path != "/" {
		useAnotherDir = true
	}
	if existsInRootPath && !useAnotherDir {
		wg.Add(1)
		go readDirAsync(afero.NewOsFs(), rootFullPath, &wg, &rootDir)
	}
	if existsInAnotherPath {
		wg.Add(1)
		go readDirAsync(afero.NewOsFs(), anotherFullPath, &wg, &anotherDir)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	if len(rootDir) > 0 && len(anotherDir) > 0 {
		log.Printf("combining results")
		finalDir = append(rootDir, anotherDir...)
	} else if len(rootDir) > 0 {
		finalDir = rootDir
	} else {
		finalDir = anotherDir
	}

	//in case of somehow the path exists in both paths due to some mess with opts.path use idc site
	if useAnotherDir {
		finalDir = anotherDir
	}

	listing := &Listing{
		Items:    []*FileInfo{},
		NumDirs:  0,
		NumFiles: 0,
	}

	for _, f := range finalDir {
		name := f.Name()
		fPath := path.Join(i.Path, name)
		if !checker.Check(fPath) {
			continue
		}
		isSymlink, isInvalidLink := false, false
		if IsSymlink(f.Mode()) {
			isSymlink = true
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead of the target's.
			info, err := i.Fs.Stat(fPath)
			if err == nil {
				f = info
			} else {
				isInvalidLink = true
			}
		}
		file := &FileInfo{
			Fs:         i.Fs,
			Name:       name,
			Size:       f.Size(),
			ModTime:    f.ModTime(),
			Mode:       f.Mode(),
			IsDir:      f.IsDir(),
			IsSymlink:  isSymlink,
			Extension:  filepath.Ext(name),
			Path:       fPath,
			currentDir: finalDir,
		}
		if !file.IsDir && strings.HasPrefix(mime.TypeByExtension(file.Extension), "image/") {
			resolution, err := calculateImageResolution(file.Fs, file.Path)
			if err != nil {
				log.Printf("Error calculating resolution for image %s: %v", file.Path, err)
			} else {
				file.Resolution = resolution
			}
		}
		if file.IsDir {
			listing.NumDirs++
		} else {
			listing.NumFiles++

			if isInvalidLink {
				file.Type = "invalid_link"
			} else {
				err := file.detectType(true, false, readHeader)
				if err != nil {
					return err
				}
			}
		}
		listing.Items = append(listing.Items, file)
	}
	i.Listing = listing
	return nil
}
