package users

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

const (
	// KB represents kilobyte in bytes
	KB = 1024
	// MB represents megabyte in bytes
	MB = 1024 * KB
	// GB represents gigabyte in bytes
	GB = 1024 * MB
	// TB represents terabyte in bytes
	TB = 1024 * GB
	// QuotaCalculationTimeout is the maximum time allowed for quota calculation
	QuotaCalculationTimeout = 30 * time.Second
)

// CalculateUserQuota calculates the total disk usage for a user's scope.
// It walks the entire directory tree and sums all file sizes.
func CalculateUserQuota(fs afero.Fs, scope string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), QuotaCalculationTimeout)
	defer cancel()

	var totalSize uint64
	done := make(chan error, 1)

	go func() {
		err := afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip files/directories that can't be accessed
				return nil
			}

			// Skip directories, only count files
			if !info.IsDir() {
				totalSize += uint64(info.Size())
			}

			return nil
		})
		done <- err
	}()

	select {
	case <-ctx.Done():
		return 0, fmt.Errorf("quota calculation timed out after %v", QuotaCalculationTimeout)
	case err := <-done:
		return totalSize, err
	}
}

// ConvertToBytes converts a value with a unit (KB, MB, GB or TB) to bytes.
func ConvertToBytes(value float64, unit string) (uint64, error) {
	switch unit {
	case "KB":
		return uint64(value * float64(KB)), nil
	case "MB":
		return uint64(value * float64(MB)), nil
	case "GB":
		return uint64(value * float64(GB)), nil
	case "TB":
		return uint64(value * float64(TB)), nil
	default:
		return 0, fberrors.ErrInvalidQuotaUnit
	}
}

// ConvertFromBytes converts bytes to a value in the specified unit (KB, MB, GB or TB).
func ConvertFromBytes(bytes uint64, unit string) (float64, error) {
	switch unit {
	case "KB":
		return float64(bytes) / float64(KB), nil
	case "MB":
		return float64(bytes) / float64(MB), nil
	case "GB":
		return float64(bytes) / float64(GB), nil
	case "TB":
		return float64(bytes) / float64(TB), nil
	default:
		return 0, fberrors.ErrInvalidQuotaUnit
	}
}

// FormatQuotaDisplay formats bytes for display, automatically selecting the appropriate unit.
func FormatQuotaDisplay(bytes uint64) string {
	if bytes >= TB {
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	}
	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	}
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%d B", bytes)
}

// CheckQuotaExceeded checks if a file operation would exceed the user's quota.
// Returns true if the quota would be exceeded, false otherwise.
func CheckQuotaExceeded(user *User, additionalSize int64) (bool, uint64, error) {
	// If quota is not set (0) or not enforced, never exceeded
	if user.QuotaLimit == 0 {
		return false, 0, nil
	}

	// Calculate current usage
	currentUsage, err := CalculateUserQuota(user.Fs, user.Scope)
	if err != nil {
		return false, 0, fmt.Errorf("failed to calculate quota: %w", err)
	}

	// Check if adding the new file would exceed the limit
	newTotal := currentUsage + uint64(additionalSize)
	exceeded := newTotal > user.QuotaLimit

	return exceeded, currentUsage, nil
}

// GetQuotaInfo returns detailed quota information for a user.
type QuotaInfo struct {
	Limit      uint64  `json:"limit"`      // Quota limit in bytes
	Used       uint64  `json:"used"`       // Current usage in bytes
	Unit       string  `json:"unit"`       // Display unit (GB or TB)
	Enforce    bool    `json:"enforce"`    // Whether quota is enforced
	Percentage float64 `json:"percentage"` // Usage percentage
}

// GetQuotaInfo retrieves quota information for a user.
func GetQuotaInfo(user *User) (*QuotaInfo, error) {
	// Calculate current usage
	used, err := CalculateUserQuota(user.Fs, user.Scope)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate quota usage: %w", err)
	}

	// Calculate percentage
	var percentage float64
	if user.QuotaLimit > 0 {
		percentage = (float64(used) / float64(user.QuotaLimit)) * 100
	}

	return &QuotaInfo{
		Limit:      user.QuotaLimit,
		Used:       used,
		Unit:       user.QuotaUnit,
		Enforce:    user.EnforceQuota,
		Percentage: percentage,
	}, nil
}
