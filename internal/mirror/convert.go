package mirror

import (
	"os"
	"path/filepath"
	"strings"
)

// RewriteAll walks hostDir and rewrites absolute href/src values in every
// .html and .css file to relative local paths. Other files are left unchanged.
// Links are rewritten relative to hostDir, so pages in subfolders are not
// adjusted.
func RewriteAll(hostDir string) error {
	host := filepath.Base(hostDir)

	return filepath.Walk(hostDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".html" && ext != ".css" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(data)
		text = strings.ReplaceAll(text, "http://"+host+"/", "")
		text = strings.ReplaceAll(text, "https://"+host+"/", "")
		text = strings.ReplaceAll(text, `href="/`, `href="`)
		text = strings.ReplaceAll(text, `src="/`, `src="`)

		return os.WriteFile(path, []byte(text), info.Mode())
	})
}
