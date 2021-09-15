package request

import (
	"encoding/json"
	"time"

	"github.com/go-logr/logr"
	"github.com/nats-io/nats.go"

	corev1beta1 "github.com/wasmCloud/wasmcloud-k8s-operator/api/v1beta1"
)

type Message struct {
	Name        string                            `json:"name"`
	Namespace   string                            `json:"namespace"`
	Application *corev1beta1.WasmCloudApplication `json:"application"`
}

type response struct {
	Status string `json:"status"`
}

type Sender struct {
	Log logr.Logger
}

func (s *Sender) Send(m Message) (response, error) {
	log := s.Log.WithValues("requesting wasmcloud-lattice-controller to reconcile", m)

	data, err := json.Marshal(m)
	if err != nil {
		log.Info("error parsing the template", "error", err)
		return response{}, err
	}
	nc, _ := nats.Connect(nats.DefaultURL)
	// TODO: replace default with lattice namespace prefix
	msg, err := nc.Request("wasmbus.alc.default", []byte(data), 1*time.Second)

	if err != nil {
		log.Info("unable to connect to the lattice controller", "error", err)
		return response{}, err
	}

	var response response
	err = json.Unmarshal(msg.Data, &response)
	if err != nil {
		log.Info("invalid json from lattice controller", "error", err)
		return response, err
	}

	return response, nil
}
