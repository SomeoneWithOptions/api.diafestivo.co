package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const defaultPort = "3002"

type config struct {
	port        string
	ipInfoToken string
	giphyKey    string
	myCIDR      string

	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	shutdownTimeout   time.Duration
	logTimeout        time.Duration
}

func loadConfig() (config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	if err := validatePort(port); err != nil {
		return config{}, err
	}

	return config{
		port:              port,
		ipInfoToken:       os.Getenv("IP_INFO_TOKEN"),
		giphyKey:          os.Getenv("GIPHY_KEY"),
		myCIDR:            os.Getenv("MY_CIDR"),
		readHeaderTimeout: 5 * time.Second,
		readTimeout:       15 * time.Second,
		writeTimeout:      30 * time.Second,
		idleTimeout:       60 * time.Second,
		shutdownTimeout:   10 * time.Second,
		logTimeout:        3 * time.Second,
	}, nil
}

func validatePort(port string) error {
	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return fmt.Errorf("invalid PORT %q: must be an integer from 1 to 65535", port)
	}
	return nil
}
