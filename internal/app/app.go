package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"wgetclone/internal/downloader"
)

func Run(args []string, stdout, stderr io.Writer) error {
	var output io.Writer
	output = stdout
	flags := flag.NewFlagSet("wget", flag.ContinueOnError)
	outputName := flags.String("O", "", "output filename")
	background := flags.Bool("B", false, "download in the background")
	ratelimit := flags.String("rate-limit", "", "ratelimit")
	inputFile := flags.String("i", "", "inputfile")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *background {
		file, err := os.Create("wget-log")
		if err != nil {
			return err
		}
		output = file
		fmt.Fprintf(stdout, "Output will be written to \"wget-log\".\n")
		defer file.Close()
	}
	var rateBytes int64
	if *ratelimit != "" {
		suffix := (*ratelimit)[len(*ratelimit)-1]
		multiplier := int64(1)
		part := (*ratelimit)[:len(*ratelimit)]
		switch suffix {
		case 'k':
			multiplier = 1024
			part = (*ratelimit)[:len(*ratelimit)-1]
		case 'M':
			multiplier = 1024 * 1024
			part = (*ratelimit)[:len(*ratelimit)-1]
		}
		partNumber, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		rateBytes = partNumber * multiplier
	}
	if *inputFile != "" {
		content, err := os.ReadFile(*inputFile)
		if err != nil {
			return err
		}
		urls := strings.Fields(string(content))
		basecfg := downloader.DownloadConfig{
			RateLimit: rateBytes,
			Out:       output,
		}
		errs := downloader.DownloadAll(urls, basecfg)
		for _, err := range errs {
			if err != nil {
				return err
			}
		}
		return nil
	}
	remainingArgs := flags.Args()
	if len(remainingArgs) == 0 {
		return errors.New("no URL provided")
	}
	cfg := downloader.DownloadConfig{
		URL:       remainingArgs[0],
		FileName:  *outputName,
		RateLimit: rateBytes,
		Out:       output,
	}

	_, err = downloader.DownloadFile(cfg)
	if err != nil {
		return err
	}

	return nil
}
