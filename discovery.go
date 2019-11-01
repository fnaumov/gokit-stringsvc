package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func ConsulRegister(consulAddr string, checkAddr string) sd.Registrar {
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
		HTTP:     "http://" + checkAddr + "/health",
		Interval: "10s",
		Timeout:  "2s",
		Notes:    "Basic health checks",
	}

	checkAddrList := strings.Split(checkAddr, ":")
	port, _ := strconv.Atoi(checkAddrList[1])
	num := rand.Intn(10000)
	asr := api.AgentServiceRegistration{
		ID:      "stringsvc" + strconv.Itoa(num),
		Name:    "stringsvc",
		Address: checkAddrList[0],
		Port:    port,
		Tags:    []string{"stringsvc", strconv.Itoa(port)},
		Check:   &check,
	}

	return consulsd.NewRegistrar(client, &asr, logger)
}
