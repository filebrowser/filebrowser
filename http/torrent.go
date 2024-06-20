package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/torrent"
)

func withPermTorrent(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Torrent {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

/*
mktorrent 1.1 (c) 2007, 2009 Emil Renner Berthing

Usage: mktorrent [OPTIONS] <target directory or filename>

Options:
-a, --announce=<url>[,<url>]* : specify the full announce URLs
                                at least one is required
                                additional -a adds backup trackers
-c, --comment=<comment>       : add a comment to the metainfo
-d, --no-date                 : don't write the creation date
-h, --help                    : show this help screen
-l, --piece-length=<n>        : set the piece length to 2^n bytes,
                                default is 18, that is 2^18 = 256kb
-n, --name=<name>             : set the name of the torrent
                                default is the basename of the target
-o, --output=<filename>       : set the path and filename of the created file
                                default is <name>.torrent
-p, --private                 : set the private flag
-s, --source=<source>         : add source string embedded in infohash
-t, --threads=<n>             : use <n> threads for calculating hashes
                                default is the number of CPU cores
-v, --verbose                 : be verbose
-w, --web-seed=<url>[,<url>]* : add web seed URLs
                                additional -w adds more URLs
*/
// TorrentOptions 结构体定义了 mktorrent 命令的选项
type TorrentOptions struct {
	Target     string   // 目标文件或目录路径
	Name       string   // 种子名称
	OutputFile string   // 输出种子文件的路径和文件名
	Announces  []string // 主要的 tracker URLs
	Comment    string   // 添加的注释
	Date       bool     // 是否写入创建日期
	PieceLen   int      // 设置块大小
	Private    bool     // 是否设置为私有种子
	Source     string   // 添加到 infohash 中的源字符串
	Threads    int      // 使用的线程数
	WebSeeds   []string // Web seed URLs
}

// buildArgs 函数根据 TorrentOptions 构建 mktorrent 命令的参数列表
func buildArgs(opts TorrentOptions) []string {
	args := []string{}

	for _, announce := range opts.Announces {
		args = append(args, "-a", announce)
	}

	if opts.Comment != "" {
		args = append(args, "-c", opts.Comment)
	}

	if !opts.Date {
		args = append(args, "-d")
	}

	if opts.PieceLen > 0 {
		args = append(args, "-l", fmt.Sprintf("%d", opts.PieceLen))
	}

	if opts.Name != "" {
		args = append(args, "-n", opts.Name)
	}

	if opts.OutputFile != "" {
		args = append(args, "-o", opts.OutputFile)
	}

	if opts.Private {
		args = append(args, "-p")
	}

	if opts.Source != "" {
		args = append(args, "-s", opts.Source)
	}

	if opts.Threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", opts.Threads))
	}

	if len(opts.WebSeeds) > 0 {
		for _, ws := range opts.WebSeeds {
			args = append(args, "-w", ws)
		}
	}

	args = append(args, opts.Target)

	return args
}

var torrentPostHandler = withPermTorrent(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
	})
	if err != nil {
		return errToStatus(err), err
	}
	fPath := file.RealPath()

	var s *torrent.Torrent
	var body torrent.CreateBody
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return http.StatusBadRequest, fmt.Errorf("failed to decode body: %w", err)
		}
		defer r.Body.Close()
	}

	// 设置 mktorrent 命令的选项
	options := TorrentOptions{
		Target:     fPath,
		Announces:  body.Announces,
		Comment:    body.Comment,
		Date:       body.Date,
		Name:       body.Name,
		OutputFile: fPath + ".torrent",
		PieceLen:   body.PieceLen,
		Private:    body.Private,
		Source:     body.Source,
		WebSeeds:   body.WebSeeds,
	}

	// 构建 mktorrent 命令的参数列表
	args := buildArgs(options)

	cmd := exec.Command("mktorrent", args...)

	err = cmd.Run()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	s = &torrent.Torrent{
		Path: fPath + ".torrent",
	}

	return renderJSON(w, r, s)
})

var publishPostHandler = withPermTorrent(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
	})
	if err != nil {
		return errToStatus(err), err
	}
	tPath := file.RealPath()
	// only folder path
	fPath := filepath.Dir(tPath)

	qbURL := "http://localhost:8081" // 修改为你的qBittorrent URL
	username := "moezakura"          // 修改为你的用户名
	password := "moezakura"          // 修改为你的密码
	torrentPath := tPath             // 修改为你的本地torrent文件路径
	savePath := fPath                // 修改为你的保存路径

	client := &http.Client{}

	sid, err := torrent.Login(client, qbURL, username, password)
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
	}

	err = torrent.AddTorrent(client, qbURL, sid, torrentPath, savePath)
	if err != nil {
		fmt.Printf("Error adding torrent: %v\n", err)
	}

	fmt.Println("Torrent added successfully!")

	return renderJSON(w, r, nil)
})
