package fakelatticecontroller

import (
	natsserver "github.com/nats-io/nats-server/test"
	"github.com/nats-io/nats.go"
)

func Setup() *nats.Conn {
	server := natsserver.RunDefaultServer()
	server.Start()

	connection, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	connection.Subscribe("wasmbus.alc.default.*", func(req *nats.Msg) {
		connection.Publish(req.Reply, []byte("{\"status\": \"received\"}"))
	})

	return connection
}
