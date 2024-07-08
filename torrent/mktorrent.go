package torrent

import (
	"fmt"
)

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
func buildArgs(opts *TorrentOptions) []string {
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
