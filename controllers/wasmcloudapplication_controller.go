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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	corev1beta1 "github.com/wasmCloud/wasmcloud-k8s-operator/api/v1beta1"
	latticeclient "github.com/wasmCloud/wasmcloud-k8s-operator/lattice_client"
)

// WasmCloudApplicationReconciler reconciles a WasmCloudApplication object
type WasmCloudApplicationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

func (r *WasmCloudApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("NamespacedName", req.NamespacedName)

	var app corev1beta1.WasmCloudApplication
	log.Info("reconciling the requested manifest", "request", req)

	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		log.Info("unable to fetch app. App is possibly deleted or does not exist yet.", "error", err)
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}

	appFinalizerName := "core.wasmcloud.com/finalizer"

	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		if !containsString(app.GetFinalizers(), appFinalizerName) {
			controllerutil.AddFinalizer(&app, appFinalizerName)

			if err := r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if containsString(app.GetFinalizers(), appFinalizerName) {
			if err := r.deleteExternalResources(&app); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&app, appFinalizerName)
			if err := r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, err
			}

			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}
	}

	if err := r.createExternalResources(&app); err != nil {
		// if fail to delete the external dependency here, return with error
		// so that it can be retried
		return ctrl.Result{}, err
	}

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

func (r *WasmCloudApplicationReconciler) createExternalResources(app *corev1beta1.WasmCloudApplication) error {
	sender := latticeclient.Client{Log: r.Log}
	response, err := sender.Put(app)

	app.Status.FromLatticeController = response.Status

	return err
}

func (r *WasmCloudApplicationReconciler) deleteExternalResources(app *corev1beta1.WasmCloudApplication) error {
	sender := latticeclient.Client{Log: r.Log}
	_, err := sender.Delete(app)

	return err
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
