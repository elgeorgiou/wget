package downloader

import (
	"bytes"
	"sync"
)

// DownloadAll downloads all urls concurrently, one goroutine per url. Each
// download gets its own copy of baseCfg with URL set to that url. Output of
// each download is buffered and written to baseCfg.Out as one block once the
// download finishes, so concurrent downloads never interleave their output.
// It returns one error per url, in the same order as urls (nil on success).
func DownloadAll(urls []string, baseCfg DownloadConfig) []error {
	errs := make([]error, len(urls))
	var wg sync.WaitGroup
	var outMu sync.Mutex

	for i, link := range urls {
		wg.Add(1)
		go func(index int, link string) {
			defer wg.Done()

			cfg := baseCfg
			cfg.URL = link

			var buf bytes.Buffer
			if baseCfg.Out != nil {
				cfg.Out = &buf
			}

			_, errs[index] = DownloadFile(cfg)

			if baseCfg.Out != nil {
				outMu.Lock()
				baseCfg.Out.Write(buf.Bytes())
				outMu.Unlock()
			}
		}(i, link)
	}
	wg.Wait()
	return errs
}
