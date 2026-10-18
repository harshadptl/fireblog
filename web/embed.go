// Package web embeds static assets (CSS, JS, images, vendored Oat UI) into
// the binary, so a single executable is all you need to deploy.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"sync"
)

// Static holds everything under web/static, served at /static/.
//
//go:embed static
var Static embed.FS

// AssetVersion is a short hash of all embedded assets. Templates append it
// to asset URLs (?v=...) so long browser caching never serves stale files
// after a deploy.
var AssetVersion = sync.OnceValue(func() string {
	h := sha256.New()
	_ = fs.WalkDir(Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := Static.ReadFile(path)
		if err == nil {
			h.Write([]byte(path))
			h.Write(b)
		}
		return err
	})
	return hex.EncodeToString(h.Sum(nil))[:10]
})
