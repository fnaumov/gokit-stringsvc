package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fnaumov/stringsvc/pb"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		logger = log.NewLogfmtLogger(os.Stderr)
		httpAddr = ":8080"
		grpcAddr = ":8081"
		consulAddr = ":8500"
	)

	rand.Seed(time.Now().UnixNano())

	var svc StringService
	svc = stringService{}
	svc = loggingMiddleware{logger, svc}

	errc := make(chan error)

	// Listen signals
	go func() {
		errc <- interrupt()
	}()

	// HTTP Server
	registrar := ConsulRegister(consulAddr, httpAddr, grpcAddr)
	registrar.Register()
	go func() {
		defer registrar.Deregister()

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

		healthHandler := httptransport.NewServer(
			makeHealthEndpoint(svc),
			decodeHealthRequest,
			encodeResponse,
		)

		mux.Handle("/uppercase", uppercaseHandler)
		mux.Handle("/count", countHandler)
		mux.Handle("/health", healthHandler)

		_ = logger.Log("protocol", "HTTP", "addr", httpAddr)
		errc <- http.ListenAndServe(httpAddr, mux)
	}()

	// GRPC Server
	go func() {
		defer registrar.Deregister()

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

	time.Sleep(1 * time.Second)
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
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure())

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
