package app

import (
	"errors"
	"flag"
	"io"
	"wgetclone/internal/downloader"
)

func Run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("wget", flag.ContinueOnError)
	outputName := flags.String("O", "", "output filename")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	remainingArgs := flags.Args()
	if len(remainingArgs) == 0 {
		return errors.New("no URL provided")
	}
	cfg := downloader.DownloadConfig{
		URL:      remainingArgs[0],
		FileName: *outputName,
		Out:      stdout,
	}
	_, err = downloader.DownloadFile(cfg)
	if err != nil {
		return err
	}

	return nil
}
