package utils

import (
	"os"
	"path/filepath"

	"go-micro.dev/v4/errors"
)

func IsDirectoryAndWriteable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.InternalServerError("NOT_A_DIRECTORY", "'%s' is not a directory", path)
	}

	f := filepath.Join(path, ".touch")
	if err := os.WriteFile(f, []byte(""), 0o600); err != nil {
		return errors.InternalServerError("DIRECTORY_NOT_WRITEABLE", "directory '%s' is not writeable: %s", path, err)
	}
	return os.Remove(f)
}
