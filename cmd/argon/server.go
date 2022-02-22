package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cli "github.com/urfave/cli/v2"

	"github.com/peertechde/argon/pkg/server"
)

var (
	FlagServerId = &cli.StringFlag{
		Name:  "id",
		Usage: "TODO",
	}
	FlagServerAddr = &cli.StringFlag{
		Name:  "addr",
		Value: "0.0.0.0",
		Usage: "TODO",
	}
	FlagServerPort = &cli.IntFlag{
		Name:  "port",
		Value: 8080,
		Usage: "TODO",
	}
	FlagServerStoragePath = &cli.StringFlag{
		Name:  "path",
		Usage: "TODO",
	}
	FlagPrometheusAddr = &cli.StringFlag{
		Name:  "prometheus_addr",
		Value: "0.0.0.0",
		Usage: "TODO",
	}
	FlagPrometheusPort = &cli.IntFlag{
		Name:  "prometheus_port",
		Value: 9090,
		Usage: "TODO",
	}
)

func ServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Start and serve the server",
		Flags: []cli.Flag{
			FlagServerId,
			FlagServerAddr,
			FlagServerPort,
			FlagServerStoragePath,
			FlagPrometheusAddr,
			FlagPrometheusPort,
		},
		Action: serverCommand,
	}
}

func serverCommand(clictx *cli.Context) error {
	if !clictx.IsSet("id") {
		return requiredFlag(clictx, "id")
	}
	if !clictx.IsSet("path") {
		return requiredFlag(clictx, "path")
	}

	var g run.Group
	{
		// termination handler
		termc := make(chan os.Signal, 1)
		signal.Notify(termc, os.Interrupt, syscall.SIGTERM)
		cancelc := make(chan struct{})

		g.Add(
			func() error {
				select {
				case <-termc:
					log.Warnf("Received SIGTERM, exiting gracefully...")
				case <-cancelc:
					break
				}
				return nil
			},
			func(err error) {
				close(cancelc)
			},
		)
	}
	{
		addr := fmt.Sprintf("%s:%d", clictx.String("prometheus_addr"), clictx.Int("prometheus_port"))
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		defer ln.Close()

		httpServer := &http.Server{
			Handler: promhttp.Handler(),
		}

		g.Add(
			func() error {
				return httpServer.Serve(ln)
			},
			func(err error) {
				httpServer.Shutdown(context.Background())
				log.Info("Succesfully shutdown http server")
			},
		)
	}
	{
		srv, err := server.New(
			server.WithId(clictx.String("id")),
			server.WithAddr(clictx.String("addr")),
			server.WithPort(clictx.Int("port")),
			server.WithStoragePath(clictx.String("path")),
		)
		if err != nil {
			return err
		}
		g.Add(
			func() error {
				return srv.Serve()
			},
			func(err error) {
				srv.Stop()
			},
		)
	}
	if err := g.Run(); err != nil {
		return err
	}
	return nil
}
