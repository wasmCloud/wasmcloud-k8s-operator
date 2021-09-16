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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	corev1beta1 "github.com/wasmCloud/wasmcloud-k8s-operator/api/v1beta1"
	"github.com/wasmCloud/wasmcloud-k8s-operator/request"
)

// WasmCloudApplicationReconciler reconciles a WasmCloudApplication object
type WasmCloudApplicationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *WasmCloudApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("wasmcloud-lattice-controller", req.NamespacedName)

	var app corev1beta1.WasmCloudApplication
	log.Info("reconciling the requested manifest", "request", req)

	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		(&request.Sender{
			Log: log,
		}).Send(request.Message{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Application: nil,
		})

		return ctrl.Result{}, nil
	}

	response, err := (&request.Sender{
		Log: log,
	}).Send(request.Message{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Application: app.DeepCopy(),
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	app.Status.FromLatticeController = response.Status

	log.Info("response from the lattice controller", "application", response.Status)

	if err := r.Status().Update(context.Background(), &app); err != nil {
		log.Info("error updating the status", "error", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WasmCloudApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1beta1.WasmCloudApplication{}).
		Owns(&corev1beta1.WasmCloudApplication{}).
		Complete(r)
}
