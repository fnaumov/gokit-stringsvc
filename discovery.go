package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"os"
	"time"
)

type DiscoveryProtocol int

const (
	DiscoveryProtocolHTTP DiscoveryProtocol = iota
	DiscoveryProtocolGRPC
)

func ConsulRegister(client consulsd.Client, addr string, protocol DiscoveryProtocol) sd.Registrar {
	logger := log.NewLogfmtLogger(os.Stderr)
	var check api.AgentServiceCheck
	var serviceName string

	switch protocol {
	case DiscoveryProtocolHTTP:
		check = api.AgentServiceCheck{
			Interval: "10s",
			Timeout:  "1s",
			Notes:  "HTTP health checks",
			HTTP:  "http://" + addr + "/health",
			Method:  "GET",
		}
		serviceName = "stringsvcHTTP"
	case DiscoveryProtocolGRPC:
		check = api.AgentServiceCheck{
			Interval: "10s",
			Timeout:  "1s",
			Notes:  "GRPC health checks",
			GRPC:  addr,
			GRPCUseTLS: false,
		}
		serviceName = "stringsvcGRPC"
	}

	date := time.Now().Format("20060102150405")
	asr := api.AgentServiceRegistration{
		ID:      serviceName + date,
		Name:    serviceName,
		Tags:    []string{"stringsvc", addr},
		Check:   &check,
	}

	return consulsd.NewRegistrar(client, &asr, logger)
}

func ConsulClient(consulAddr string) consulsd.Client {
	logger := log.NewLogfmtLogger(os.Stderr)

	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		_ = logger.Log("err", err)
		os.Exit(1)
	}
	client := consulsd.NewClient(consulClient)

	return client
}
