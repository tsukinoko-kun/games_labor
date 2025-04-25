package public

import (
	"crypto/md5"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	unixpath "path"
	"strings"
)

var (
	//go:embed *.min.css *.svg js/*.js
	public                 embed.FS
	servedPathToNormalPath map[string]string
	normalPathToServedPath map[string]string
	normalPathToIntegrity  map[string]string
)

func init() {
	servedPathToNormalPath = make(map[string]string)
	normalPathToServedPath = make(map[string]string)
	normalPathToIntegrity = make(map[string]string)

	fs.WalkDir(public, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		f, err := public.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		h1 := md5.Sum(b)
		shortHash := base64.URLEncoding.EncodeToString(h1[:])[:16]
		ext := unixpath.Ext(name)
		hashed := fmt.Sprintf(
			"%s.%s%s",
			name[:len(name)-len(ext)],
			shortHash,
			ext,
		)

		servedPathToNormalPath["/public/"+hashed] = name
		normalPathToServedPath[name] = "/public/" + hashed

		h256 := sha256.Sum256(b)
		integrityHash := base64.StdEncoding.EncodeToString(h256[:])
		normalPathToIntegrity[name] = "sha256-" + integrityHash

		return nil
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var f fs.File

	name, ok := servedPathToNormalPath[r.URL.Path]

	if ok {
		var err error
		f, err = public.Open(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Cache-Control", "max-age=31536000") // 1 year
	} else {
		f, err := public.Open(strings.TrimPrefix(r.URL.Path, "/public/"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Cache-Control", "max-age=300") // 5 minutes
	}

	switch ext := unixpath.Ext(name); ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	}

	_, _ = io.Copy(w, f)
}

func Path(name string) string {
	return normalPathToServedPath[name]
}

func Integrity(name string) string {
	if integrity, ok := normalPathToIntegrity[name]; ok {
		return integrity
	}
	return ""
}
