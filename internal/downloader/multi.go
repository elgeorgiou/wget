package downloader

import "sync"

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
