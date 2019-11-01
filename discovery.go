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

func ConsulRegister(consulAddr string, checkAddr string, protocol string) sd.Registrar {
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

	check := api.AgentServiceCheck{
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "Basic health checks",
	}

	switch protocol {
	case "HTTP":
		check.HTTP = "http://" + checkAddr + "/health"
		check.Method = "GET"
	case "GRPS":
		check.GRPC = "http://" + checkAddr + "/health"
		check.GRPCUseTLS = true
	}

	checkAddrList := strings.Split(checkAddr, ":")
	port, _ := strconv.Atoi(checkAddrList[1])
	date := time.Now().Format("20060102150405")
	asr := api.AgentServiceRegistration{
		ID:      "stringsvc" + date,
		Name:    "stringsvc",
		Address: checkAddrList[0],
		Port:    port,
		Tags:    []string{"stringsvc", strconv.Itoa(port)},
		Check:   &check,
	}

	return consulsd.NewRegistrar(client, &asr, logger)
}
