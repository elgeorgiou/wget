package downloader

import (
	"bytes"
	"sync"
)

// DownloadAll downloads all urls concurrently, one goroutine per url. Each
// download gets its own copy of baseCfg with URL set to that url and FileName
// cleared, so every file keeps the name from its own url. Output of each
// download is buffered and written to baseCfg.Out as one block once the
// download finishes, so concurrent downloads never interleave their output.
// It returns one error per url, in the same order as urls (nil on success).
func DownloadAll(urls []string, baseCfg DownloadConfig) []error {
	errs := make([]error, len(urls))
	var wg sync.WaitGroup
	var printing sync.Mutex

	for i, link := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()

			output, err := downloadOne(link, baseCfg)
			errs[i] = err

			if baseCfg.Out != nil {
				// Only one download prints at a time.
				printing.Lock()
				baseCfg.Out.Write(output)
				printing.Unlock()
			}
		}()
	}
	wg.Wait()
	return errs
}

// downloadOne downloads link and returns everything the download printed,
// instead of printing it straight away.
func downloadOne(link string, baseCfg DownloadConfig) ([]byte, error) {
	var output bytes.Buffer

	cfg := baseCfg
	cfg.URL = link
	cfg.FileName = ""
	cfg.Out = &output

	_, err := DownloadFile(cfg)
	return output.Bytes(), err
}
