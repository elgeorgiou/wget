// Package app provides command-line argument parsing and execution
// coordination for the wget clone.
//
// It processes download options, configures output destinations and rate
// limits, and dispatches requests to the appropriate download operation.
//
// The package supports single-file downloads, concurrent downloads from
// a URL list, background logging, and recursive website mirroring.
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

// Run parses command-line arguments and executes the requested download operation.
//
// The function configures command-line flags, selects the output destination,
// validates rate-limit values, and dispatches execution based on the supplied
// options. It supports downloading a single file, downloading multiple URLs
// concurrently from an input file, and mirroring a website.
//
// When background mode is enabled, Run redirects download output to a
// "wget-log" file and writes a notification to standard output.
//
// Parameters:
// - args: Command-line arguments, excluding the executable name.
// - stdout: Writer for standard output and background-mode notifications.
// - stderr: Writer provided for standard error output.
//
// Returns:
// - An error if argument parsing, configuration, or a download operation fails.
// - Nil if the requested operation completes successfully.
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
			multiplier = 1000
			part = (*ratelimit)[:len(*ratelimit)-1]
		case 'M':
			multiplier = 1000 * 1000
			part = (*ratelimit)[:len(*ratelimit)-1]
		default:
			if suffix < '0' || suffix > '9' {
				return errors.New("invalid rate-limit suffix, use k or M")
			}
		}
		partNumber, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return errors.New("invalid rate-limit value")
		}
		if partNumber <= 0 {
			return errors.New("rate-limit must be positive")
		}
		rateBytes = partNumber * multiplier
	}
	// Handle multi-file downloads before single-file and mirror operations.
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
	// Dispatch mirror requests separately from single-file downloads.
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
