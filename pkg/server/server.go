package server

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/peertechde/argon/api"
	"github.com/peertechde/argon/pkg/logging"
	"github.com/peertechde/argon/pkg/storage"
	"github.com/peertechde/argon/pkg/storage/local"
)

var log = logging.Logger.WithField(logging.Subsys, "server")

func New(options ...Option) (*Server, error) {
	var opts Options
	opts.Apply(options...)

	if opts.StoragePath == "" {
		return nil, fmt.Errorf("missing storage path")
	}

	fi, err := os.Stat(opts.StoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Errorf("path '%s' doesn't exist", opts.StoragePath)
		}
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.Errorf("path '%s' is not a directory", opts.StoragePath)
	}

	srv := &Server{
		options: opts,
	}

	return srv, nil
}

type Server struct {
	options Options

	grpcServer     *grpc.Server
	storageService *StorageService
	store          storage.Storage
}

func (s *Server) Serve() error {
	log.WithFields(logrus.Fields{
		"id":   s.options.Id,
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"path": s.options.StoragePath,
	}).Info("Starting the server")

	s.registerMetrics()
	s.storageService = NewStorageService(local.New(s.options.StoragePath))

	addr := fmt.Sprintf("%s:%d", s.options.Addr, s.options.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}
	s.grpcServer = NewGRPCServer(
		WithGRPCServerOptions(grpc.StatsHandler(&grpcStatsHandler{})),
	)
	api.RegisterStorageServer(s.grpcServer, s.storageService)

	err = s.grpcServer.Serve(ln)
	if err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	log.Info("Trying to gracefully stop the server...")
	s.grpcServer.GracefulStop()

	log.Info("Successfully stopped the server")
	return nil
}

func (s *Server) registerMetrics() {
	// general metrics
	prometheus.MustRegister(modInfo)

	// grpc metrics
	prometheus.MustRegister(grpcConnsOpen)
	prometheus.MustRegister(grpcConnsTotal)
	prometheus.MustRegister(grpcRequestsPending)
	prometheus.MustRegister(grpcRequestsTotal)

	// go_mod_info; name and version of used modules
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, dep := range buildInfo.Deps {
		d := dep
		if dep.Replace != nil {
			d = dep.Replace
		}
		modInfo.WithLabelValues(d.Path, d.Version)
	}
}
