package logtalez

import (
	"fmt"
	"strings"

	"github.com/zeromq/goczmq"
)

// LogTalez holds the context for a running LogTalez instance
type LogTalez struct {
	topics     []string
	endpoints  []string
	serverCert *goczmq.Cert
	clientCert *goczmq.Cert
	sock       *goczmq.Sock
}

// MakeTopicList is a convenience function that, given a string of comma delimited
// hosts and a string of comma delimited program name tags, generates a slice of
// subscription topics.
func MakeTopicList(hosts, programs string) []string {
	topicList := make([]string, 0)

	if hosts != "" {
		for _, h := range strings.Split(hosts, ",") {
			if programs != "" {
				for _, p := range strings.Split(programs, ",") {
					topicList = append(topicList, fmt.Sprintf("%s.%s", h, p))
				}
			} else {
				topicList = append(topicList, h)
			}
		}
	} else {
		topicList = append(topicList, "")
	}

	return topicList
}

// MakeEndpointList is a convenience function that, given a list of comma delimited
// zeromq endpoints, returns a slice containing the endpoints.
func MakeEndpointList(endpoints string) []string {
	endpointList := make([]string, 0)

	for _, e := range strings.Split(endpoints, ",") {
		endpointList = append(endpointList, e)
	}

	return endpointList
}

// New returns a new running LogTalez instance given a slice of endpoints,
// a slice of topics, and the path to a CURVE server public cert and CURVE
// server client cert.  The TailChan member is a channel that returns
// [][]byte messages from ZeroMQ.
func New(endpoints, topics []string, serverCertPath, clientCertPath string) (*LogTalez, error) {

	lt := &LogTalez{
		topics:    make([]string, 0),
		endpoints: make([]string, 0),
	}

	var err error

	lt.serverCert, err = goczmq.NewCertFromFile(serverCertPath)
	if err != nil {
		return lt, err
	}

	lt.clientCert, err = goczmq.NewCertFromFile(clientCertPath)
	if err != nil {
		return lt, err
	}

	lt.sock = goczmq.NewSock(goczmq.Sub)

	lt.clientCert.Apply(lt.sock)
	lt.sock.SetCurveServerkey(lt.serverCert.PublicText())

	for _, t := range topics {
		lt.topics = append(lt.topics, t)
		lt.sock.SetSubscribe(t)
	}

	for _, e := range endpoints {
		err = lt.sock.Connect(e)
		if err != nil {
			return lt, err
		}
		lt.endpoints = append(lt.endpoints, e)
	}

	return lt, nil
}

func (lt *LogTalez) Read(p []byte) (int, error) {
	return lt.sock.Read(p)
}

// Destroy gracefully shuts down a running LogTalez instance
func (lt *LogTalez) Destroy() {
	lt.serverCert.Destroy()
	lt.clientCert.Destroy()
	lt.sock.Destroy()
}
