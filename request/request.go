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

func (s *Sender) Put(app *corev1beta1.WasmCloudApplication) (response, error) {
	r, e := s.send("put", app)
	return r, e
}
func (s *Sender) Delete(app *corev1beta1.WasmCloudApplication) (response, error) {
	r, e := s.send("delete", app)
	return r, e
}

func (s *Sender) send(verb string, app *corev1beta1.WasmCloudApplication) (response, error) {
	log := s.Log.WithValues("requesting wasmcloud-lattice-controller to reconcile", app)

	data, err := json.Marshal(app)
	if err != nil {
		log.Info("error parsing the template", "error", err)
		return response{}, err
	}
	nc, _ := nats.Connect(nats.DefaultURL)
	// TODO: replace default with lattice namespace prefix
	topic := "wasmbus.alc.default." + verb
	msg, err := nc.Request(topic, []byte(data), 1*time.Second)

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
