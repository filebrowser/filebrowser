package handlers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	e "errors"
	"hash"
	"io"
	"net/http"
	"os"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/utils/errors"
)

// Checksum calculates the hash of a file. Supports MD5, SHA1, SHA256 and SHA512.
func Checksum(w http.ResponseWriter, r *http.Request, c *config.Config, i *file.Info) (int, error) {
	query := r.URL.Query().Get("checksum")

	file, err := os.Open(i.Path)
	if err != nil {
		return errors.ErrorToHTTPCode(err, true), err
	}

	defer file.Close()

	var h hash.Hash

	switch query {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return http.StatusBadRequest, e.New("Unknown HASH type")
	}

	_, err = io.Copy(h, file)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	val := hex.EncodeToString(h.Sum(nil))
	w.Write([]byte(val))
	return http.StatusOK, nil
}
