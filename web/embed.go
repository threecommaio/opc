package web

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

// NewFS returns a http.FileSystem for the public directory
func NewFS(fsys fs.FS, root string) (http.FileSystem, error) {
	if gin.Mode() == gin.ReleaseMode {
		f, err := fs.Sub(fsys, root)
		if err != nil {
			return nil, fmt.Errorf("New: %w", err)
		}

		return http.FS(f), nil
	}

	_, filename, _, _ := runtime.Caller(1)
	exPath, err := filepath.Abs(filepath.Join(filepath.Dir(filename), root))

	if err != nil {
		return nil, err
	}

	return http.Dir(exPath), nil
}
