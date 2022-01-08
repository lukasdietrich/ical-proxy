package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/lukasdietrich/ical-proxy/internal/proxy"
)

func main() {
	var filename string

	flag.StringVar(&filename, "config", "config.yml", "The filename of the configuration file")
	flag.Parse()

	config, err := parseConfig(filename)
	if err != nil {
		log.Fatalf("could not read config file %q: %v", filename, err)
	}

	mux, err := proxy.NewMux(config.Calendars)
	if err != nil {
		log.Fatalf("could not build calendar mux: %w", err)
	}

	address := net.JoinHostPort(config.HTTP.Host, config.HTTP.Port)
	log.Printf("starting ical-proxy on %q", address)

	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("ical-proxy stopped: %v", err)
	}
}
