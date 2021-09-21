package fakelatticecontroller

import (
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func Setup() *server.Server {
	server := natsserver.RunDefaultServer()
	server.Start()

	return server

}

type FakeLatticeController struct {
	conn     *nats.Conn
	messages chan *nats.Msg
}

// Finish the test and close connections.
func (f *FakeLatticeController) Close() {
	f.conn.Close()
	close(f.messages)
}

// Return the next message that was received over NATS.
// If you SpyNextMessage() and there are no more messages, this function
// will return nil.
// Be sure to call FakeLatticeController.Close() before using this
// method, and do all of your assertions at the end of the test,
// to avoid hard-to-debug timeouts. If you do not call Close, this
// method may time out, rather than returning nil.
func (f *FakeLatticeController) SpyNextMessage() *nats.Msg {
	result := <-f.messages
	return result
}

// Set up a fake lattice controller.
// Returns the nats connection for the lattice controller,
// and a channel for spying on the received messages.
func SetupSubscriber() *FakeLatticeController {
	connection, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic("Unable to connect to NATS server")
	}
	messages := make(chan *nats.Msg, 1000)
	connection.Subscribe("wasmbus.alc.default.*", func(req *nats.Msg) {
		messages <- req
		connection.Publish(req.Reply, []byte("{\"status\": \"received\"}"))
	})

	return &FakeLatticeController{connection, messages}
}
