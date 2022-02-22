package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"

	"github.com/peertechde/argon/pkg/client"
)

var (
	FlagTarget = &cli.StringFlag{
		Name:  "target",
		Value: "0.0.0.0:8080",
		Usage: "TODO",
	}
	FlagName = &cli.StringFlag{
		Name:  "name",
		Usage: "TODO",
	}
	FlagTo = &cli.StringFlag{
		Name:  "to",
		Usage: "TODO",
	}
)

func UploadCommand() *cli.Command {
	return &cli.Command{
		Name:  "upload",
		Usage: "Upload",
		Flags: []cli.Flag{
			FlagTarget,
			FlagName,
		},
		Action: uploadCommand,
	}
}

func uploadCommand(clictx *cli.Context) error {
	if !clictx.IsSet("name") {
		return requiredFlag(clictx, "name")
	}

	// termination handler
	termc := make(chan os.Signal, 1)
	signal.Notify(termc, os.Interrupt, syscall.SIGTERM)

	opctx, opcancel := context.WithCancel(context.Background())
	defer opcancel()

	go func() {
		select {
		case <-termc:
			log.Warnf("Received SIGTERM, exiting gracefully...")
			opcancel()
		}
	}()

	c := client.New()
	if err := c.DialContext(opctx, clictx.String("target")); err != nil {
		return errors.Wrap(err, "failed to dial")
	}

	return c.Upload(opctx, clictx.String("name"))
}

func DownloadCommand() *cli.Command {
	return &cli.Command{
		Name:  "download",
		Usage: "Download",
		Flags: []cli.Flag{
			FlagTarget,
			FlagName,
			FlagTo,
		},
		Action: downloadCommand,
	}
}

func downloadCommand(clictx *cli.Context) error {
	if !clictx.IsSet("name") {
		return requiredFlag(clictx, "name")
	}
	if !clictx.IsSet("to") {
		return requiredFlag(clictx, "to")
	}

	// termination handler
	termc := make(chan os.Signal, 1)
	signal.Notify(termc, os.Interrupt, syscall.SIGTERM)

	opctx, opcancel := context.WithCancel(context.Background())
	defer opcancel()

	go func() {
		select {
		case <-termc:
			log.Warnf("Received SIGTERM, exiting gracefully...")
			opcancel()
		}
	}()

	c := client.New()
	if err := c.DialContext(opctx, clictx.String("target")); err != nil {
		return errors.Wrap(err, "failed to dial")
	}

	return c.Download(opctx, clictx.String("name"), clictx.String("to"))
}
