package fs

import (
	"main/assets"
)

type TestFS struct{}

func (fs *TestFS) ReadFile(name string) ([]byte, error) {
	return assets.EmbedFS.ReadFile(name)
}
