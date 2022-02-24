package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	FlagFileName = &cli.StringFlag{
		Name:  "name",
		Usage: "TODO",
	}
	FlagTo = &cli.StringFlag{
		Name:  "to",
		Usage: "TODO",
	}
	FlagOldFile = &cli.StringFlag{
		Name:  "old",
		Usage: "TODO",
	}
	FlagNewFile = &cli.StringFlag{
		Name:  "new",
		Usage: "TODO",
	}
)

func WriteCommand() *cli.Command {
	return &cli.Command{
		Name:  "write",
		Usage: "Write",
		Flags: []cli.Flag{
			FlagTarget,
			FlagFileName,
		},
		Action: writeCommand,
	}
}
func ReadCommand() *cli.Command {
	return &cli.Command{
		Name:  "read",
		Usage: "Read",
		Flags: []cli.Flag{
			FlagTarget,
			FlagFileName,
			FlagTo,
		},
		Action: readCommand,
	}
}

func ListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List",
		Flags: []cli.Flag{
			FlagTarget,
		},
		Action: listCommand,
	}
}

func StatCommand() *cli.Command {
	return &cli.Command{
		Name:  "stat",
		Usage: "Stat",
		Flags: []cli.Flag{
			FlagTarget,
			FlagFileName,
		},
		Action: statCommand,
	}
}

func RemoveCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove",
		Flags: []cli.Flag{
			FlagTarget,
			FlagFileName,
		},
		Action: removeCommand,
	}
}

func RenameCommand() *cli.Command {
	return &cli.Command{
		Name:  "rename",
		Usage: "Rename",
		Flags: []cli.Flag{
			FlagTarget,
			FlagOldFile,
			FlagNewFile,
		},
		Action: renameCommand,
	}
}

func writeCommand(clictx *cli.Context) error {
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

	return c.Write(opctx, clictx.String("name"))
}

func readCommand(clictx *cli.Context) error {
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

	return c.Read(opctx, clictx.String("name"), clictx.String("to"))
}

func listCommand(clictx *cli.Context) error {
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

	files, err := c.List(opctx)
	if err != nil {
		return err
	}
	out, err := json.Marshal(files)
	if err != nil {
		return errors.Wrap(err, "failed to marshal file info")
	}

	fmt.Println(string(out))
	return nil
}

func statCommand(clictx *cli.Context) error {
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

	fileInfo, err := c.Stat(opctx, clictx.String("name"))
	if err != nil {
		return err
	}
	out, err := json.Marshal(fileInfo)
	if err != nil {
		return errors.Wrap(err, "failed to marshal file info")
	}

	fmt.Println(string(out))
	return nil
}

func removeCommand(clictx *cli.Context) error {
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

	err := c.Remove(opctx, clictx.String("name"))
	if err != nil {
		return err
	}

	return nil
}

func renameCommand(clictx *cli.Context) error {
	if !clictx.IsSet("old") {
		return requiredFlag(clictx, "old")
	}
	if !clictx.IsSet("new") {
		return requiredFlag(clictx, "new")
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

	err := c.Rename(opctx, clictx.String("old"), clictx.String("new"))
	if err != nil {
		return err
	}

	return nil
}
