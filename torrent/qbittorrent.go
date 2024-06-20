package torrent

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func login(client *http.Client, qbURL, username, password string) (string, error) {
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", qbURL)
	data := fmt.Sprintf("username=%s&password=%s", username, password)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", qbURL)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("SID cookie not found")
}

func addTorrent(client *http.Client, qbURL, sid, torrentPath, savePath string) error {
	addTorrentURL := fmt.Sprintf("%s/api/v2/torrents/add", qbURL)

	file, err := os.Open(torrentPath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("torrents", filepath.Base(torrentPath))
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, file); err != nil {
		return err
	}

	writer.WriteField("savepath", savePath)
	writer.Close()

	req, err := http.NewRequest("POST", addTorrentURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", fmt.Sprintf("SID=%s", sid))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add torrent with status code: %d", resp.StatusCode)
	}

	return nil
}
