package fileutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

const RootTestDir string = "testing_dir_func"

func createFileStructure() error {
	childDir := filepath.Join(RootTestDir, "child_dir")

	err := os.MkdirAll(childDir, 0755)
	if err != nil {
		return err
	}

	data := []byte("test_data")

	err = os.WriteFile(filepath.Join(childDir, "test_file"), data, 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(RootTestDir, "test_file"), data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func cleanupFileStructure() {
	os.RemoveAll(RootTestDir)
}

func TestDiskUsageOnFile(t *testing.T) {
	err := createFileStructure()
	if err != nil {
		t.Errorf("createFileStructure() failed: %s", err)
	}
	defer cleanupFileStructure()

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Getwd() failed: %s", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), cwd)
	size, inodes, err := DiskUsage(fs, filepath.Join(RootTestDir, "test_file"))

	require.NoError(t, err)
	require.Equal(t, int64(9), size)
	require.Equal(t, int64(1), inodes)
}

func TestDiskUsageOnNestedDir(t *testing.T) {
	err := createFileStructure()
	if err != nil {
		t.Errorf("createFileStructure() failed: %s", err)
	}
	defer cleanupFileStructure()

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Getwd() failed: %s", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), cwd)
	size, inodes, err := DiskUsage(fs, filepath.Join(RootTestDir, "child_dir"))

	require.NoError(t, err)
	require.GreaterOrEqual(t, size, int64(105))
	require.Equal(t, int64(2), inodes)
}

func TestDiskUsageOnRootDir(t *testing.T) {
	err := createFileStructure()
	if err != nil {
		t.Errorf("createFileStructure() failed: %s", err)
	}
	defer cleanupFileStructure()

	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Getwd() failed: %s", err)
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), cwd)
	size, inodes, err := DiskUsage(fs, RootTestDir)

	require.NoError(t, err)
	require.GreaterOrEqual(t, size, int64(242))
	require.Equal(t, int64(4), inodes)
}
