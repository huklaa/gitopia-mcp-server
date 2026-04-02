package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello world"), 0644))

	content, err := Read(tmpDir, "test.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello world", content)
}

func TestRead_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Read(tmpDir, "../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside the allowed directory")
}

func TestRead_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Read(tmpDir, "nonexistent.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestRead_NestedFile(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "dir")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	testFile := filepath.Join(subDir, "nested.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("nested content"), 0644))

	content, err := Read(tmpDir, "sub/dir/nested.txt")
	require.NoError(t, err)
	assert.Equal(t, "nested content", content)
}

func TestWrite_CreateFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := Write(tmpDir, "new-file.txt", "new content")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "new-file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "new content", string(content))
}

func TestWrite_CreateNestedDirs(t *testing.T) {
	tmpDir := t.TempDir()

	err := Write(tmpDir, "deep/nested/dir/file.txt", "deep content")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "deep/nested/dir/file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "deep content", string(content))
}

func TestWrite_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	err := Write(tmpDir, "../escape.txt", "bad content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside the allowed directory")
}

func TestWrite_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("old"), 0644))

	err := Write(tmpDir, "existing.txt", "new")
	require.NoError(t, err)

	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "new", string(content))
}

func TestApplyChange_Create(t *testing.T) {
	tmpDir := t.TempDir()

	err := ApplyChange(tmpDir, FileChange{Path: "created.txt", Content: "created", Mode: "create"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "created.txt"))
	require.NoError(t, err)
	assert.Equal(t, "created", string(content))
}

func TestApplyChange_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "to-delete.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("bye"), 0644))

	err := ApplyChange(tmpDir, FileChange{Path: "to-delete.txt", Mode: "delete"})
	require.NoError(t, err)

	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestApplyChange_DeleteNonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	err := ApplyChange(tmpDir, FileChange{Path: "nonexistent.txt", Mode: "delete"})
	assert.NoError(t, err) // Should not error for nonexistent files
}

func TestApplyChange_UnsupportedMode(t *testing.T) {
	tmpDir := t.TempDir()

	err := ApplyChange(tmpDir, FileChange{Path: "file.txt", Mode: "invalid"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestApplyChange_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	err := ApplyChange(tmpDir, FileChange{Path: "../escape.txt", Content: "bad", Mode: "create"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outside the allowed directory")
}

func TestRead_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")
	require.NoError(t, os.WriteFile(testFile, []byte(""), 0644))

	content, err := Read(tmpDir, "empty.txt")
	require.NoError(t, err)
	assert.Equal(t, "", content)
}

func TestWrite_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()

	err := Write(tmpDir, "empty.txt", "")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "empty.txt"))
	require.NoError(t, err)
	assert.Equal(t, "", string(content))
}

func TestApplyChange_ModifyExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "existing.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("old"), 0644))

	err := ApplyChange(tmpDir, FileChange{Path: "existing.txt", Content: "updated", Mode: "modify"})
	require.NoError(t, err)

	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "updated", string(content))
}

func TestRead_PrefixSiblingAttack(t *testing.T) {
	// Create /tmp/workspace and /tmp/workspace2
	// A naive HasPrefix check would allow /tmp/workspace2/secret.txt
	// to pass a boundary check for /tmp/workspace
	workspace := t.TempDir() // e.g., /tmp/TestXXX/workspace
	sibling := workspace + "2" // e.g., /tmp/TestXXX/workspace2
	require.NoError(t, os.MkdirAll(sibling, 0755))
	secretFile := filepath.Join(sibling, "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("stolen"), 0644))

	// Try to read the sibling file by computing a relative path that
	// resolves to the sibling directory
	_, err := Read(workspace, "../"+filepath.Base(sibling)+"/secret.txt")
	require.Error(t, err, "prefix-sibling attack should be blocked")
	assert.Contains(t, err.Error(), "outside the allowed directory")
}

func TestWrite_PrefixSiblingAttack(t *testing.T) {
	workspace := t.TempDir()
	sibling := workspace + "2"
	require.NoError(t, os.MkdirAll(sibling, 0755))

	err := Write(workspace, "../"+filepath.Base(sibling)+"/evil.txt", "pwned")
	require.Error(t, err, "prefix-sibling write attack should be blocked")
	assert.Contains(t, err.Error(), "outside the allowed directory")

	// Verify file was NOT created
	_, statErr := os.Stat(filepath.Join(sibling, "evil.txt"))
	assert.True(t, os.IsNotExist(statErr), "evil.txt should not exist")
}

func TestRead_SymlinkTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a symlink pointing outside the allowed directory
	outsideDir := t.TempDir()
	secretFile := filepath.Join(outsideDir, "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("secret data"), 0644))

	symlinkPath := filepath.Join(tmpDir, "link")
	err := os.Symlink(outsideDir, symlinkPath)
	if err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	// Attempt to read through symlink — the path resolves inside tmpDir
	// but the actual file is outside. EvalSymlinks should catch this.
	_, err = Read(tmpDir, "link/secret.txt")
	require.Error(t, err, "symlink traversal should be blocked")
	assert.Contains(t, err.Error(), "outside the allowed directory")
}
