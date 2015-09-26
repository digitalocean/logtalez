package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/digitalocean/logtalez"
)

func main() {

	endpointsPtr := flag.String("endpoints", "", "comma delimited list of zeromq endpoints")
	topicsPtr := flag.String("topics", "", "comma delimited list of topics to subscribe to")
	serverCertPathPtr := flag.String("servercertpath", "", "path to server public cert")
	clientCertPathPtr := flag.String("clientcertpath", "", "path to client public cert")
	jSONPtr := flag.Bool("json", false, "restrict output to valid JSON")

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

	if _, err := os.Stat(*serverCertPathPtr); err != nil {
		log.Fatalf("error reading server certificate %q: %s", *serverCertPathPtr, err)
	}

	if _, err := os.Stat(*clientCertPathPtr); err != nil {
		log.Fatalf("error reading client certificate %q: %s", *clientCertPathPtr, err)
	}

	topicList := make([]string, 0)
	for _, t := range strings.Split(*topicsPtr, ",") {
		topicList = append(topicList, t)
	}

	endpointList := make([]string, 0)
	for _, e := range strings.Split(*endpointsPtr, ",") {
		endpointList = append(endpointList, e)
	}

	lt, err := logtalez.New(endpointList, topicList, *serverCertPathPtr, *clientCertPathPtr)
	if err != nil {
		log.Fatal(err)
	}
	defer lt.Destroy()

	buf := make([]byte, 65536)
	for {
		n, err := lt.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		output := string(buf[:n])
		if *jSONPtr {
			output = strings.SplitAfterN(output, "@cee:", 2)[1]
		}

		fmt.Println(output)
	}
}
