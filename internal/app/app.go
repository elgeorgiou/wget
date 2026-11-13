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
	"wgetclone/internal/mirror"
)

func Run(args []string, stdout, stderr io.Writer) error {
	var output io.Writer
	output = stdout
	flags := flag.NewFlagSet("wget", flag.ContinueOnError)
	outputName := flags.String("O", "", "output filename")
	background := flags.Bool("B", false, "download in the background")
	ratelimit := flags.String("rate-limit", "", "ratelimit")
	inputFile := flags.String("i", "", "inputfile")
	outputDir := flags.String("P", "", "output directory")
	mirrorFlag := flags.Bool("mirror", false, "mirror a website")
	reject := flags.String("R", "", "file suffixes to reject")
	rejectLong := flags.String("reject", "", "file suffixes to reject")
	exclude := flags.String("X", "", "exclude directories")
	excludeLong := flags.String("exclude", "", "exclude directories")
	convertLinks := flags.Bool("convert-links", false, "convert links for offline viewing")
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
		part := (*ratelimit)
		switch suffix {
		case 'k':
			multiplier = 1024
			part = (*ratelimit)[:len(*ratelimit)-1]
		case 'M':
			multiplier = 1024 * 1024
			part = (*ratelimit)[:len(*ratelimit)-1]
		default:
			if suffix < '0' || suffix > '9' {
				return errors.New("invalid rate-limit suffix, use k or M")
			}
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
			RateLimit:  rateBytes,
			Out:        output,
			NoProgress: *background,
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
	if *mirrorFlag {
		var rejectSuffixes []string
		if *reject != "" {
			rejectSuffixes = strings.Split(*reject, ",")
		}
		if *rejectLong != "" {
			rejectSuffixes = append(rejectSuffixes, strings.Split(*rejectLong, ",")...)
		}
		var excludePaths []string
		if *exclude != "" {
			excludePaths = strings.Split(*exclude, ",")
		}
		if *excludeLong != "" {
			excludePaths = append(excludePaths, strings.Split(*excludeLong, ",")...)
		}
		opts := mirror.MirrorOptions{
			RejectSuffixes: rejectSuffixes,
			ExcludePaths:   excludePaths,
			ConvertLinks:   *convertLinks,
		}
		err = mirror.Mirror(remainingArgs[0], opts, downloader.DownloadFile)
		if err != nil {
			return err
		}
		return nil
	}
	cfg := downloader.DownloadConfig{
		URL:        remainingArgs[0],
		OutputDir:  *outputDir,
		FileName:   *outputName,
		RateLimit:  rateBytes,
		Out:        output,
		NoProgress: *background,
	}

	_, err = downloader.DownloadFile(cfg)
	if err != nil {
		return err
	}

	return nil
}
