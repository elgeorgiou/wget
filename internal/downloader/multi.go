package downloader

import "sync"

// DownloadAll downloads all urls concurrently, one goroutine per url, using dl
// to download each one. It waits for every download to finish and returns one
// error per url, in the same order as urls (nil if the download succeeded).
func DownloadAll(urls []string, dl func(string) error) []error {
	errs := make([]error, len(urls))
	var wg sync.WaitGroup

	for i, link := range urls {
		wg.Add(1)
		go func(index int, link string) {
			defer wg.Done()
			errs[index] = dl(link)
		}(i, link)
	}
	wg.Wait()
	return errs
}
