package server

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

var (
	// general metrics
	modInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "go_mod_info",
	}, []string{"name", "version"})

	// grpc related metrics
	grpcConnsOpen = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grpc",
		Name:      "connections_open",
	})
	grpcConnsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "grpc",
		Name:      "connections_total",
	})
	grpcRequestsPending = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "grpc",
		Name:      "requests_pending",
	})
	grpcRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "grpc",
		Name:      "requests_total",
	})
)

type grpcStatsHandler struct{}

func (*grpcStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (*grpcStatsHandler) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	switch stat.(type) {
	case *stats.Begin:
		grpcRequestsPending.Inc()
	case *stats.End:
		grpcRequestsPending.Dec()
		grpcRequestsTotal.Inc()
	case *stats.OutHeader, *stats.InHeader, *stats.InTrailer, *stats.OutTrailer:
		// do nothing
	case *stats.OutPayload:
		// todo
	case *stats.InPayload:
		// todo
	default:
		log.Warn("unexpected grpc stats handler type")
	}
}

func (*grpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (*grpcStatsHandler) HandleConn(ctx context.Context, stat stats.ConnStats) {
	switch stat.(type) {
	case *stats.ConnBegin:
		grpcConnsOpen.Inc()
		grpcConnsTotal.Inc()
	case *stats.ConnEnd:
		grpcConnsOpen.Dec()
	}
}
