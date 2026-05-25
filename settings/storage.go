package settings

import (
	"os"
	"strconv"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/users"
)

// StorageBackend is a settings storage backend.
type StorageBackend interface {
	Get() (*Settings, error)
	Save(*Settings) error
	GetServer() (*Server, error)
	SaveServer(*Server) error
}

// Storage is a settings storage.
type Storage struct {
	back StorageBackend
}

// NewStorage creates a settings storage from a backend.
func NewStorage(back StorageBackend) *Storage {
	return &Storage{back: back}
}

// Get returns the settings for the current instance.
func (s *Storage) Get() (*Settings, error) {
	set, err := s.back.Get()
	if err != nil {
		return nil, err
	}

	if set.UserHomeBasePath == "" {
		set.UserHomeBasePath = DefaultUsersHomeBasePath
	}

	if set.LogoutPage == "" {
		set.LogoutPage = DefaultLogoutPage
	}

	if set.MinimumPasswordLength == 0 {
		set.MinimumPasswordLength = DefaultMinimumPasswordLength
	}

	if set.Tus == (Tus{}) {
		set.Tus = Tus{
			ChunkSize:  DefaultTusChunkSize,
			RetryCount: DefaultTusRetryCount,
		}
	}

	if set.FileMode == 0 {
		set.FileMode = DefaultFileMode
	}

	if set.DirMode == 0 {
		set.DirMode = DefaultDirMode
	}

	if set.ClamAV.URL == "" {
		set.ClamAV.URL = strings.TrimSpace(os.Getenv("CLAMAV_URL"))
	}
	if set.ClamAV.ScanDepth == 0 {
		if depth, err := strconv.Atoi(strings.TrimSpace(os.Getenv("CLAMAV_SCAN_DEPTH"))); err == nil && depth > 0 {
			set.ClamAV.ScanDepth = depth
		}
	}

	if !set.ConvertX.Configured {
		if set.ConvertX.URL == "" {
			set.ConvertX.URL = strings.TrimRight(strings.TrimSpace(firstNonEmptyEnv("FB_CONVERTX_URL", "CONVERTX_URL")), "/")
		}
		if set.ConvertX.APIKey == "" {
			set.ConvertX.APIKey = strings.TrimSpace(firstNonEmptyEnv("FB_CONVERTX_API_KEY", "CONVERTX_API_KEY", "CONVERTX_API_TOKEN"))
		}
		if set.ConvertX.Timeout == "" {
			set.ConvertX.Timeout = strings.TrimSpace(firstNonEmptyEnv("FB_CONVERTX_TIMEOUT", "CONVERTX_TIMEOUT"))
		}
		if enabled := strings.TrimSpace(firstNonEmptyEnv("FB_CONVERTX_ENABLED", "CONVERTX_ENABLED")); enabled != "" {
			if parsed, err := strconv.ParseBool(enabled); err == nil {
				set.ConvertX.Enabled = parsed
			}
		}
	}
	if set.ConvertX.Timeout == "" {
		set.ConvertX.Timeout = "2m"
	}

	return set, nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

var defaultEvents = []string{
	"save",
	"copy",
	"rename",
	"upload",
	"delete",
}

// Save saves the settings for the current instance.
func (s *Storage) Save(set *Settings) error {
	if len(set.Key) == 0 {
		return fberrors.ErrEmptyKey
	}

	if set.Defaults.Locale == "" {
		set.Defaults.Locale = "en"
	}

	if set.Defaults.Commands == nil {
		set.Defaults.Commands = []string{}
	}

	if set.Defaults.ViewMode == "" {
		set.Defaults.ViewMode = users.MosaicViewMode
	}

	if set.Rules == nil {
		set.Rules = []rules.Rule{}
	}

	if set.Shell == nil {
		set.Shell = []string{}
	}

	if set.Commands == nil {
		set.Commands = map[string][]string{}
	}

	for _, event := range defaultEvents {
		if _, ok := set.Commands["before_"+event]; !ok {
			set.Commands["before_"+event] = []string{}
		}

		if _, ok := set.Commands["after_"+event]; !ok {
			set.Commands["after_"+event] = []string{}
		}
	}

	err := s.back.Save(set)
	if err != nil {
		return err
	}

	return nil
}

// GetServer wraps StorageBackend.GetServer.
func (s *Storage) GetServer() (*Server, error) {
	return s.back.GetServer()
}

// SaveServer wraps StorageBackend.SaveServer and adds some verification.
func (s *Storage) SaveServer(ser *Server) error {
	ser.Clean()
	return s.back.SaveServer(ser)
}
