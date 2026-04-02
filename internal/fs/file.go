package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileChange represents a file change for batch operations.
// It's defined here to be shared across different tools.
type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Mode    string `json:"mode"` // "create", "modify", "delete"
}

// containsPath checks whether target is contained within root.
// Uses filepath.Rel instead of strings.HasPrefix to prevent the
// prefix-sibling attack (e.g., /workspace2/x passing a check for /workspace).
func containsPath(root, target string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("could not get absolute path for base %s: %w", root, err)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("could not get absolute path for target %s: %w", target, err)
	}

	// Resolve symlinks on the target path to catch symlinks inside the
	// workspace that point outside. We do NOT resolve the root because
	// OS-level symlinks (e.g., macOS /var -> /private/var) are consistent
	// between root and target, and resolving them asymmetrically breaks
	// the prefix check.
	if resolved, err := filepath.EvalSymlinks(absTarget); err == nil {
		absTarget = resolved
		// Also resolve root so both are on the same real filesystem
		if resolvedRoot, err := filepath.EvalSymlinks(absRoot); err == nil {
			absRoot = resolvedRoot
		}
	}

	// Ensure root ends with separator for exact prefix matching
	rootWithSep := absRoot + string(filepath.Separator)

	// Target must be root itself or start with root + separator
	if absTarget != absRoot && !strings.HasPrefix(absTarget, rootWithSep) {
		return fmt.Errorf("path (%s) is outside the allowed directory (%s)", absTarget, absRoot)
	}

	// Double-check with filepath.Rel: result must not start with ".."
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return fmt.Errorf("could not compute relative path: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("path (%s) escapes the allowed directory (%s)", absTarget, absRoot)
	}

	return nil
}

// ApplyChange applies a single file change relative to a base path.
func ApplyChange(basePath string, change FileChange) error {
	fullPath := filepath.Join(basePath, change.Path)

	if err := containsPath(basePath, fullPath); err != nil {
		return err
	}

	absFullPath, _ := filepath.Abs(fullPath)

	switch change.Mode {
	case "create", "modify":
		dir := filepath.Dir(absFullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(absFullPath, []byte(change.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", absFullPath, err)
		}

	case "delete":
		if err := os.Remove(absFullPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file %s: %w", absFullPath, err)
		}

	default:
		return fmt.Errorf("unsupported file change mode: %s", change.Mode)
	}

	return nil
}

// Read safely reads content from a file, ensuring it's within a specific root dir.
func Read(rootDir, filePath string) (string, error) {
	fullPath := filepath.Join(rootDir, filePath)

	if err := containsPath(rootDir, fullPath); err != nil {
		return "", err
	}

	absFullPath, _ := filepath.Abs(fullPath)
	data, err := os.ReadFile(absFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", absFullPath, err)
	}

	return string(data), nil
}

// Write safely writes content to a file, ensuring it's within a specific root dir.
func Write(rootDir, filePath, content string) error {
	fullPath := filepath.Join(rootDir, filePath)

	if err := containsPath(rootDir, fullPath); err != nil {
		return err
	}

	absFullPath, _ := filepath.Abs(fullPath)

	dir := filepath.Dir(absFullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(absFullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", absFullPath, err)
	}

	return nil
}
