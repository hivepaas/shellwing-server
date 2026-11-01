package assets

import (
	"embed"
	"io/fs"
)

//go:embed templates
var templateEmbedFS embed.FS

// GetTemplatesFS returns a filesystem scoped to the templates directory
func GetTemplatesFS() fs.FS {
	subFS, err := fs.Sub(templateEmbedFS, "templates")
	if err != nil {
		panic(err)
	}
	return subFS
}
