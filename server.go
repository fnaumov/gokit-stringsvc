package main

import (
	"fmt"
	"github.com/fnaumov/stringsvc/pb"
	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	logger = log.NewLogfmtLogger(os.Stderr)
	errc = make(chan error)
	httpAddr = ":8080"
	grpcAddr = ":8081"
	consulAddr = "127.0.0.1:8500"
	authConfig = authService{
		key: []byte("secret_key"),
		clients: map[string]string{
			"user1": "passwordOne",
			"user2": "passwordTwo",
		},
	}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var svc StringService
	svc = stringService{authConfig}
	svc = loggingMiddleware{authConfig, logger, svc}

	// Listen signals
	go func() {
		errc <- interrupt()
	}()

	consulClient := ConsulClient(consulAddr)

	runHTTPServer(consulClient, svc, httpAddr)
	runGRPCServer(consulClient, svc, grpcAddr)

	// time.Sleep(1 * time.Second)

	_ = logger.Log("fatal", <-errc)
}

func runHTTPServer(consulClient consulsd.Client, svc StringService, addr string) {
	handler := makeHttpHandler(svc)

	registrarHTTP := ConsulRegister(consulClient, addr, DiscoveryProtocolHTTP)
	go func() {
		registrarHTTP.Register()
		defer registrarHTTP.Deregister()

		_ = logger.Log("output", fmt.Sprintf("Starting HTTP server at %s", addr))
		errc <- http.ListenAndServe(addr, handler)
	}()
}

func runGRPCServer(consulClient consulsd.Client, svc StringService, addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		errc <- err
		os.Exit(1)
	}

	srv := grpc.NewServer()
	healthServer := health.NewServer()
	pb.RegisterStringServiceServer(srv, grpcBinding{svc, healthServer})
	healthpb.RegisterHealthServer(srv, grpcBinding{svc, healthServer})

	registrarGRPC := ConsulRegister(consulClient, addr, DiscoveryProtocolGRPC)
	go func() {
		registrarGRPC.Register()
		defer registrarGRPC.Deregister()

		_ = logger.Log("output", fmt.Sprintf("Starting GRPC server at %s", addr))
		errc <- srv.Serve(ln)
	}()
}

func interrupt() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}
