package fakelatticecontroller

import (
	"time"

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
	messages []*nats.Msg
}

// Finish the test and close connections.
func (f *FakeLatticeController) Close() {
	f.conn.Close()
}

// Return the next message that was received over NATS.
// If you SpyNextMessage() and there are no more messages, this function
// will return nil.
func (f *FakeLatticeController) SpyNextMessage() *nats.Msg {
	if len(f.messages) == 0 {
		return nil
	}

	result, messages := f.messages[0], f.messages[1:]
	f.messages = messages

	return result
}

// wait for up to 10 seconds for a message to arrive
// or return nil
func (f *FakeLatticeController) WaitForMessage() *nats.Msg {
	message := f.SpyNextMessage()
	counter := 0

	for message == nil && counter < 10 {
		time.Sleep(1 * time.Second)
		message = f.SpyNextMessage()

		counter++
	}

	return message
}

// Set up a fake lattice controller.
// Returns the nats connection for the lattice controller,
// and a channel for spying on the received messages.
func SetupSubscriber() *FakeLatticeController {
	connection, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic("Unable to connect to NATS server")
	}
	messages := make([]*nats.Msg, 0)

	f := &FakeLatticeController{connection, messages}

	connection.Subscribe("wasmbus.alc.default.*", func(req *nats.Msg) {
		f.messages = append(f.messages, req)
		connection.Publish(req.Reply, []byte("{\"status\": \"received\"}"))
	})

	return f
}
