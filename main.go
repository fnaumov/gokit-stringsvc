package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fnaumov/stringsvc/pb"
	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		httpAddr = ":8080"
		grpcAddr = ":8081"
	)

	logger := log.NewLogfmtLogger(os.Stderr)

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
	countResult := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{}) // no fields here

	var svc StringService
	svc = stringService{}
	svc = loggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	errc := make(chan error)

	go func() {
		errc <- interrupt()
	}()

	go func() {
		mux := http.NewServeMux()

		uppercaseHandler := httptransport.NewServer(
			makeUppercaseEndpoint(svc),
			decodeUppercaseRequest,
			encodeResponse,
		)

		countHandler := httptransport.NewServer(
			makeCountEndpoint(svc),
			decodeCountRequest,
			encodeResponse,
		)

		mux.Handle("/uppercase", uppercaseHandler)
		mux.Handle("/count", countHandler)
		mux.Handle("/metrics", promhttp.Handler())

		_ = logger.Log("protocol", "HTTP", "addr", httpAddr)
		errc <- http.ListenAndServe(httpAddr, mux)
	}()

	go func() {
		_ = logger.Log("protocol", "GRPC", "addr", grpcAddr)
		ln, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			errc <- err
			return
		}

		s := grpc.NewServer()
		pb.RegisterStringServiceServer(s, grpcBinding{svc})
		errc <- s.Serve(ln)
	}()

	testHTTPRequest()
	testGRPCRequest()

	_ = logger.Log("fatal", <-errc)
}

func interrupt() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return fmt.Errorf("%s", <-c)
}

func testHTTPRequest() {
	logger := log.NewLogfmtLogger(os.Stderr)
	requestBody, err := json.Marshal(uppercaseRequest{
		S: "Hello, this response for HTTP request!",
	})
	if err != nil {
		_ = logger.Log(err)
	}

	resp, err := http.Post("http://localhost:8080/uppercase", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		_ = logger.Log(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = logger.Log(err)
	}

	m := uppercaseResponse{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		_ = logger.Log(err)
	}

	_ = logger.Log("response", m.V)
}

func testGRPCRequest() {
	logger := log.NewLogfmtLogger(os.Stderr)
	conn, err := grpc.Dial("127.0.0.1:8081", grpc.WithInsecure())

	if err != nil {
		grpclog.Fatalf("Fail to dial: %v", err)
	}

	defer conn.Close()

	client := pb.NewStringServiceClient(conn)
	request := &pb.UppercaseRequest{
		S: "Hello, this response for GRPC request!",
	}
	response, err := client.Uppercase(context.Background(), request)

	if err != nil {
		grpclog.Fatalf("Fail to dial: %v", err)
	}

	_ = logger.Log("response", response.V)
}
