package torrent

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type Torrent struct {
	*settings.Settings
	*users.User
}

func (t *Torrent) MakeTorrent(fPath string, body users.CreateTorrentBody) error {
	tPath := fPath + ".torrent"

	// 设置 mktorrent 命令的选项
	opts := &TorrentOptions{
		Target:     fPath,
		Announces:  body.Announces,
		Comment:    body.Comment,
		Date:       body.Date,
		Name:       body.Name,
		OutputFile: tPath,
		PieceLen:   body.PieceLen,
		Private:    body.Private,
		Source:     body.Source,
		WebSeeds:   body.WebSeeds,
	}

	args := buildArgs(opts)

	cmd := exec.Command("mktorrent", args...)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (t *Torrent) PublishTorrent(torrentPath string, savePath string) error {
	qbURL := t.Settings.Torrent.QbUrl
	username := t.Settings.Torrent.QbUsername
	password := t.Settings.Torrent.QbPassword

	client := &http.Client{}
	sid, err := login(client, qbURL, username, password)
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}

	err = addTorrent(client, qbURL, sid, torrentPath, savePath)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %v", err)
	}

	return nil
}

func (t *Torrent) GetDefaultCreateBody(createBody users.CreateTorrentBody) (*users.CreateTorrentBody, error) {
	announces, err := fetchTrackerList(t.Settings.Torrent.TrackersListUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tracker list: %v", err)
	}

	if createBody.Announces == nil {
		createBody.Announces = []string{}
	}

	if createBody.WebSeeds == nil {
		createBody.WebSeeds = []string{}
	}

	return &users.CreateTorrentBody{
		Announces: announces,
		Comment:   createBody.Comment,
		Date:      createBody.Date,
		Name:      "",
		PieceLen:  createBody.PieceLen,
		Private:   createBody.Private,
		Source:    createBody.Source,
		WebSeeds:  createBody.WebSeeds,
	}, nil
}

func fetchTrackerList(url string) ([]string, error) {
	// 发送HTTP GET请求
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trackers: %v", err)
	}
	defer response.Body.Close()

	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch trackers: status code %d", response.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 移除空行
	text := string(body)
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")
	var filteredLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	return filteredLines, nil
}
