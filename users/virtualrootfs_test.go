package users

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestScopes(t *testing.T) (string, *VirtualRootFs) {
	t.Helper()

	base := t.TempDir()

	// Create scope directories with files
	projectsDir := filepath.Join(base, "projects")
	mediaDir := filepath.Join(base, "media")

	require.NoError(t, os.MkdirAll(projectsDir, 0o755))
	require.NoError(t, os.MkdirAll(mediaDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(projectsDir, "readme.txt"), []byte("hello"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(mediaDir, "photo.jpg"), []byte("image"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(projectsDir, "src"), 0o755))

	vfs := NewVirtualRootFs(base, []string{"/projects", "/media"})
	return base, vfs
}

func TestVirtualRootFs_StatRoot(t *testing.T) {
	_, vfs := setupTestScopes(t)

	info, err := vfs.Stat("/")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestVirtualRootFs_StatScope(t *testing.T) {
	_, vfs := setupTestScopes(t)

	info, err := vfs.Stat("/projects")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	assert.Equal(t, "projects", info.Name())
}

func TestVirtualRootFs_StatFileInScope(t *testing.T) {
	_, vfs := setupTestScopes(t)

	info, err := vfs.Stat("/projects/readme.txt")
	require.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Equal(t, "readme.txt", info.Name())
}

func TestVirtualRootFs_StatNonExistent(t *testing.T) {
	_, vfs := setupTestScopes(t)

	_, err := vfs.Stat("/nonexistent")
	assert.Error(t, err)
}

func TestVirtualRootFs_OpenRoot(t *testing.T) {
	_, vfs := setupTestScopes(t)

	f, err := vfs.Open("/")
	require.NoError(t, err)
	defer f.Close()

	entries, err := f.Readdir(-1)
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	assert.Contains(t, names, "projects")
	assert.Contains(t, names, "media")
}

func TestVirtualRootFs_OpenFileInScope(t *testing.T) {
	_, vfs := setupTestScopes(t)

	f, err := vfs.Open("/projects/readme.txt")
	require.NoError(t, err)
	defer f.Close()

	buf := make([]byte, 100)
	n, err := f.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(buf[:n]))
}

func TestVirtualRootFs_CreateFile(t *testing.T) {
	_, vfs := setupTestScopes(t)

	f, err := vfs.Create("/media/newfile.txt")
	require.NoError(t, err)
	_, err = f.WriteString("new content")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// Verify it exists
	info, err := vfs.Stat("/media/newfile.txt")
	require.NoError(t, err)
	assert.Equal(t, "newfile.txt", info.Name())
}

func TestVirtualRootFs_CreateInRootDenied(t *testing.T) {
	_, vfs := setupTestScopes(t)

	_, err := vfs.Create("/rootfile.txt")
	assert.Error(t, err)
}

func TestVirtualRootFs_RemoveFile(t *testing.T) {
	_, vfs := setupTestScopes(t)

	err := vfs.Remove("/projects/readme.txt")
	require.NoError(t, err)

	_, err = vfs.Stat("/projects/readme.txt")
	assert.Error(t, err)
}

func TestVirtualRootFs_RenameWithinScope(t *testing.T) {
	_, vfs := setupTestScopes(t)

	err := vfs.Rename("/projects/readme.txt", "/projects/renamed.txt")
	require.NoError(t, err)

	_, err = vfs.Stat("/projects/renamed.txt")
	assert.NoError(t, err)
}

func TestVirtualRootFs_RenameCrossScopeDenied(t *testing.T) {
	_, vfs := setupTestScopes(t)

	err := vfs.Rename("/projects/readme.txt", "/media/readme.txt")
	assert.Error(t, err)
}

func TestVirtualRootFs_RemoveScopeRootDenied(t *testing.T) {
	_, vfs := setupTestScopes(t)

	assert.Error(t, vfs.Remove("/projects"))
	assert.Error(t, vfs.RemoveAll("/projects"))
}

func TestVirtualRootFs_RenameScopeRootDenied(t *testing.T) {
	_, vfs := setupTestScopes(t)

	assert.Error(t, vfs.Rename("/projects", "/renamed"))
}

func TestVirtualRootFs_RealPath(t *testing.T) {
	base, vfs := setupTestScopes(t)

	realPath, err := vfs.RealPath("/projects/readme.txt")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(base, "projects", "readme.txt"), realPath)
}

func TestVirtualRootFs_RealPathRoot(t *testing.T) {
	_, vfs := setupTestScopes(t)

	_, err := vfs.RealPath("/")
	assert.Error(t, err)
}
