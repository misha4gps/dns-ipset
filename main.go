package main

import (
	"flag"
	"github.com/miekg/dns"
	"log"
	"strconv"
)

var configFile = flag.String("c", "config.yaml", "configuration file")
var config = &Config{}
var ipSet = NewIpSet()

var cache = NewMemoryCache()

func main() {
	flag.Parse()
	var err error
	config, err = loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// attach request handler func
	dns.HandleFunc(".", handleDnsRequest)

	// start server
	//port := 5354
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	log.Printf("Starting at %d\n", port)
	err = server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %s\n ", err.Error())
	}
}
