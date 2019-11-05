package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"os"
	"strconv"
	"strings"
	"time"
)

func ConsulRegister(consulAddr string, httpAddr string, grpcAddr string) sd.Registrar {
	logger := log.NewLogfmtLogger(os.Stderr)

	// Service discovery domain.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddr
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			_ = logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	checkHTTP := api.AgentServiceCheck{
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "HTTP health checks",
		HTTP:     "http://" + httpAddr + "/health",
		Method:   "GET",
	}

	checkGRPS := api.AgentServiceCheck{
		Interval:   "10s",
		Timeout:    "1s",
		Notes:      "GRPS health checks",
		GRPC:       grpcAddr,
		GRPCUseTLS: false,
	}

	httpAddrList := strings.Split(httpAddr, ":")
	httpPort, _ := strconv.Atoi(httpAddrList[1])
	grpcAddrList := strings.Split(grpcAddr, ":")
	grpcPort, _ := strconv.Atoi(grpcAddrList[1])
	date := time.Now().Format("20060102150405")
	asr := api.AgentServiceRegistration{
		ID:      "stringsvc" + date,
		Name:    "stringsvc",
		Tags:    []string{"stringsvc", strconv.Itoa(httpPort), strconv.Itoa(grpcPort)},
		Checks:  []*api.AgentServiceCheck{&checkHTTP, &checkGRPS},
	}

	return consulsd.NewRegistrar(client, &asr, logger)
}
