package logtalez

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/zeromq/goczmq"
)

func TestMakeTopicList(t *testing.T) {
	hosts := "host1,host2"
	programs := "program1,program2"
	topics := MakeTopicList(hosts, programs)

	expected := []string{
		"host1.program1",
		"host1.program2",
		"host2.program1",
		"host2.program2",
	}

	for i, expect := range expected {
		if topics[i] != expect {
			t.Errorf("expected topic '%s', got '%s'", expect[i], topics[i])
		}
	}
}

func TestMakeEndpointList(t *testing.T) {
	conns := "tcp://incproc1,tcp://inproc2"
	endpoints := MakeEndpointList(conns)

	expected := []string{
		"tcp://incproc1",
		"tcp://inproc2",
	}

	for i, expect := range expected {
		if endpoints[i] != expect {
			t.Errorf("expected endpoint '%s', got '%s'", expect[i], endpoints[i])
		}
	}
}

func TestNew(t *testing.T) {
	endpoints := []string{"inproc://test1"}
	topics := []string{"topic1", "topic2"}
	servercert := "./example_certs/example_curve_server_cert"
	clientcert := "./example_certs/example_curve_client_cert"

	auth := goczmq.NewAuth()
	defer auth.Destroy()

	clientCert, err := goczmq.NewCertFromFile(clientcert)
	if err != nil {
		t.Fatal(err)
	}
	defer clientCert.Destroy()

	server := goczmq.NewSock(goczmq.Pub)

	defer server.Destroy()
	server.SetZapDomain("global")

	serverCert, err := goczmq.NewCertFromFile(servercert)
	defer serverCert.Destroy()
	if err != nil {
		t.Fatal(err)
	}

	serverCert.Apply(server)
	server.SetCurveServer(1)

	err = auth.Curve("./example_certs/")
	if err != nil {
		t.Fatal(err)
	}

	server.Bind(endpoints[0])

	lt, err := New(endpoints, topics, servercert, clientcert)
	if err != nil {
		t.Error("NewLogTalez failed: %s", err)
	}

	server.SendFrame([]byte("topic1:hello world"), 0)

	buf := make([]byte, 65536)

	n, err := lt.Read(buf)
	if err != io.EOF {
		t.Errorf("expected %s, got %s", io.EOF, err)
	}

	if string(buf[:n]) != "topic1:hello world" {
		t.Errorf("expected 'topic1:hello world', got '%s'", buf[:n])
	}

	server.SendFrame([]byte("topic2:hello again"), 0)

	n, err = lt.Read(buf)
	if err != io.EOF {
		t.Errorf("expected %s, got %s", io.EOF, err)
	}

	if string(buf[:n]) != "topic2:hello again" {
		t.Errorf("expected 'topic2:hello again', got '%s'", buf[:n])
	}
}

func ExampleLogTalez() {
	// endpoints is a list of zeromq rsyslog endpoints to attach to.
	endpoints := []string{"tcp://host1.example.com:24444,tcp://host2.example.com:24444"}

	// topics is a list of topics to subscribe to.
	topics := []string{"host1.ssh,host1.nginx,host2.ssh,host2.kernel"}

	// path to the server public certificate
	serverCertPath := "/home/example_user/.curve/server_cert"

	// path to the client public certificate
	clientCertPath := "/home/example_user/.curve/my_cert"

	// create a new logtalez instance
	lt, err := New(endpoints, topics, serverCertPath, clientCertPath)
	if err != nil {
		log.Fatal(err)
	}
	defer lt.Destroy()

	// logtalez exposes an io.Reader interface.  so, here we create a
	// buffer and read log lines.  currently logtalez returns one log
	// line per call to Read.
	buf := make([]byte, 65536)
	for {
		n, err := lt.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		fmt.Println(string(buf[:n]))
	}
}
