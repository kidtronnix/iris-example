package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/project-iris/iris-go.v1"
)

// Before you run this service you need to setup a relay node on your machine.
// This is done by downloading the binary from http://iris.karalabe.com/
// You may need to move the binary somewhere sensible according to your OS.
// Then start it up using `iris -dev`. Simple as that.

// EchoService is a simple example of a service with attached methods for iris messaging
type EchoService struct{}

// Init get called when service is initialised during registration.
func (b *EchoService) Init(conn *iris.Connection) error { return nil }

// HandleBroadcast handles broadcast messages sent by clientConn to service.
// All service instances in the cluster receive each message.
func (b *EchoService) HandleBroadcast(msg []byte) { fmt.Printf("  [broadcast] %s\n", msg) }

// HandleRequest handles request sent by a clientConn who are expecting a response.
// These requests are load balanaced amongst the service instances in the cluster.
func (b *EchoService) HandleRequest(req []byte) ([]byte, error) {
	fmt.Printf("  [request] %s\n", req)
	return append([]byte(" ..."), req...), nil
}

// HandleTunnel handles tunnel type messaging.
// These connections are load balanaced amongst service instances in the cluster.
func (b *EchoService) HandleTunnel(tun *iris.Tunnel) {}

// HandleDrop handles dropped connections
func (b *EchoService) HandleDrop(reason error) {}

// TopicSubscriber is a listener that subscribes to a certain topic.
type TopicSubscriber struct {
	InstanceName string
}

// HandleEvent handles event messages sent by publishers.
// All subscribers will receive each event sent.
func (h *TopicSubscriber) HandleEvent(event []byte) {
	fmt.Printf("  [%s] %s\n", h.InstanceName, event)
}

func main() {

	// Uncomment to turn off logging if it annoys you
	iris.Log.SetHandler(log15.DiscardHandler())

	// register the echo service service
	service, err := iris.Register(55555, "echo", &EchoService{}, nil)
	if err != nil {
		log.Fatalf("failed to register to the Iris relay: %v.", err)
	}
	defer service.Unregister()

	// registers a subscriber to a topic
	sub1Conn, err := iris.Connect(55555)
	if err != nil {
		log.Fatalf("failed to connect to the Iris relay: %v.", err)
	}
	defer sub1Conn.Close()
	sub1Conn.Subscribe("event-topic", &TopicSubscriber{"subscriber 1"}, nil)

	// lets registers another subscriber to the same topic
	sub2Conn, err := iris.Connect(55555)
	if err != nil {
		log.Fatalf("failed to connect to the Iris relay: %v.", err)
	}
	defer sub2Conn.Close()
	sub2Conn.Subscribe("event-topic", &TopicSubscriber{"subscriber 2"}, nil)

	// make a clientConn for testing it all out and sending some messages!
	clientConn, err := iris.Connect(55555)
	if err != nil {
		log.Fatalf("failed to connect to the Iris relay: %v.", err)
	}
	defer clientConn.Close()

	// make a request to echo service
	resp, _ := clientConn.Request("echo", []byte("hello"), time.Second)
	fmt.Printf("  [response] %s\n", resp)

	// broadcast message to all memebers of the echo cluster
	clientConn.Broadcast("echo", []byte("boom"))

	// publish a message to subscribers listening on the topic
	clientConn.Publish("event-topic", []byte("weeeee!"))

}
