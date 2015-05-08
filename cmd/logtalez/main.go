package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/digitalocean/logtalez"
)

func main() {

	endpointsPtr := flag.String("endpoints", "", "comma delimited list of zeromq endpoints")
	hostsPtr := flag.String("hosts", "", "comma delimited list of hostnames to get logs from")
	programsPtr := flag.String("programs", "", "comma delimited list of programs to get logs from")
	serverCertPathPtr := flag.String("servercertpath", "", "path to server public cert")
	clientCertPathPtr := flag.String("clientcertpath", "", "path to client public cert")

	flag.Parse()

	if *endpointsPtr == "" {
		log.Fatal("--endpoints is mandatory")
	}

	if *serverCertPathPtr == "" {
		log.Fatal("--servercertpath is mandatory")
	}

	if *clientCertPathPtr == "" {
		log.Fatal("--clientcertpath is mandatory")
	}

	topicList := logtalez.MakeTopicList(*hostsPtr, *programsPtr)
	endpointList := logtalez.MakeEndpointList(*endpointsPtr)

	lt, err := logtalez.New(endpointList, topicList, *serverCertPathPtr, *clientCertPathPtr)
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	for {
		select {
		case msg := <-lt.TailChan:
			logline := strings.Split(string(msg[0]), "@cee:")[1]
			fmt.Println(logline)
		case <-sigChan:
			lt.Destroy()
			time.Sleep(100 * time.Millisecond)
			os.Exit(0)
		}
	}
}
