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

func SetupSubscriber() *nats.Conn {
	connection, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic("Unable to connect to NATS server")
	}

	connection.Subscribe("wasmbus.alc.default.*", func(req *nats.Msg) {
		connection.Publish(req.Reply, []byte("{\"status\": \"received\"}"))
	})

	return connection
}
