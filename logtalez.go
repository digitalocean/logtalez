package logtalez

import (
	"fmt"
	"strings"
	"time"

	"github.com/zeromq/goczmq"
)

type LogTalez struct {
	topics     []string
	endpoints  []string
	serverCert *goczmq.Cert
	clientCert *goczmq.Cert
	sock       *goczmq.Sock
	channeler  *goczmq.Channeler
	TailChan   <-chan [][]byte
}

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

func MakeEndpointList(endpoints string) []string {
	endpointList := make([]string, 0)

	for _, e := range strings.Split(endpoints, ",") {
		endpointList = append(endpointList, e)
	}

	return endpointList
}

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

	lt.sock = goczmq.NewSock(goczmq.SUB)

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

	lt.channeler = goczmq.NewChanneler(lt.sock, false)
	lt.TailChan = lt.channeler.RecvChan
	return lt, nil
}

func (lt *LogTalez) Destroy() {
	for _, t := range lt.topics {
		lt.sock.SetUnsubscribe(t)
	}

	for _, e := range lt.endpoints {
		err := lt.sock.Disconnect(e)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	lt.serverCert.Destroy()
	lt.clientCert.Destroy()
	lt.sock.Destroy()
}
