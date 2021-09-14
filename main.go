/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nats-io/nats.go"
	corev1beta1 "github.com/wasmCloud/wasmcloud-k8s-operator/api/v1beta1"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(corev1beta1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

type message struct {
	Name        string                            `json:"name"`
	Namespace   string                            `json:"namespace"`
	Application *corev1beta1.WasmCloudApplication `json:"application"`
}
type response struct {
	Status string `json:"status"`
}

func Send(m message) response {
	data, err := json.Marshal(m)
	if err != nil {
		setupLog.Error(err, "error parsing the template")
		os.Exit(1)
	}
	nc, _ := nats.Connect(nats.DefaultURL)
	// TODO: replace default with lattice namespace prefix
	msg, err := nc.Request("wasmbus.alc.default", []byte(data), 1*time.Second)

	if err != nil {
		setupLog.Error(err, "unable to connect to the lattice controller")
		os.Exit(1)
	}

	var response response
	err = json.Unmarshal(msg.Data, &response)
	if err != nil {
		setupLog.Error(err, "invalid json from lattice controller")
		os.Exit(1)
	}

	return response
}

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("wasmcloud-lattice-controller", req.NamespacedName)
	log.V(1).Info("reconciling with wasmcloud-lattice-controller")

	var app corev1beta1.WasmCloudApplication
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		Send(message{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Application: nil,
		})

		return ctrl.Result{}, nil
	}

	response := Send(message{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Application: app.DeepCopy(),
	})

	app.Status.FromLatticeController = response.Status

	log.V(1).Info(app.Status.FromLatticeController)

	err := r.Status().Update(context.Background(), &app)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "854763a3.wasmcloud.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctrl.NewControllerManagedBy(mgr).
		For(&corev1beta1.WasmCloudApplication{}).
		Owns(&corev1beta1.WasmCloudApplication{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
