/*
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
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webserverv1alpha1 "github.com/example-inc/memcached-operator/api/v1alpha1"
)

// WebServerReconciler reconciles a WebServer object
type WebserverReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.example.com,resources=webserver,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=webserver/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *WebserverReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background() // this context will NOT trigger a new Reconcile. It is often used to update Status about the result from a Reconcile action.
	log := r.Log.WithValues("webServer", req.NamespacedName)

	// Fetch the Webserver instance
	webserver := &webserverv1alpha1.Webserver{} // TODO: update schema
	err := r.Get(ctx, req.NamespacedName, webserver)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("WebServer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get WebServer")
		return ctrl.Result{}, err
	}

	// Check if deployment exists, if not create it
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: webserver.Name, Namespace: webserver.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForWebserver(webserver)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		log.Info("New Deployment created successfully - return and requeue")
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	latencyJSONNumber := getLatency()
	latencyFloat64, err := latencyJSONNumber.Float64()

	// webserver.Status.Latency = latency
	// err = r.Status().Update(ctx, webserver)
	// if err != nil {
	// 	log.Error(err, "Failed to update Webserver status")
	// 	return ctrl.Result{}, err
	// }

	// Update Status.Latency if needed
	if latencyFloat64 > 2.5 {
		log.Info("Latency is larger than 0.5")
		// TODO: increase number of pods
		newSize := webserver.Spec.Size + 1
		found.Spec.Replicas = &newSize
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil

	}

	// Update Status.Latency if needed
	if latencyFloat64 < 1.5 {
		log.Info("Latency is less than 1.5. latencyFloat64: ", latencyFloat64)
		// Update the Webserver status with the pod names
		// List the pods for this webserver's deployment
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(webserver.Namespace),
			client.MatchingLabels(labelsForWebserver(webserver.Name)),
		}
		if err = r.List(ctx, podList, listOpts...); err != nil {
			log.Error(err, "Failed to list pods", "webserver.Namespace", webserver.Namespace, "webserver.Name", webserver.Name)
			return ctrl.Result{}, err
		}

		//numPods := len(podList.Items)
		numPods := podList.Size()
		log.Info("numPods is: ", numPods)
		if numPods > 1 {
			// TODO: reduce number of pods
			log.Info("Latency is less than 1.5 AND numPods > 1")
			newSize := webserver.Spec.Size - 1
			found.Spec.Replicas = &newSize
			err = r.Update(ctx, found)
			if err != nil {
				log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{RequeueAfter: time.Second * 5}, nil
		}
	}

	webserver.Status.Latency = latencyJSONNumber
	err = r.Status().Update(ctx, webserver)
	if err != nil {
		log.Error(err, "Failed to update Webserver status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// deploymentForWebServer returns a webserver Deployment object
func (r *WebserverReconciler) deploymentForWebserver(ws *webserverv1alpha1.Webserver) *appsv1.Deployment {
	ls := labelsForWebserver(ws.Name)
	replicas := ws.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name,
			Namespace: ws.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "containerstack/cpustress",
						Name:  "ws-stresstest",
						Args:  []string{"--cpu=2", "--timeout=4800s"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "ping",
						}},
					}},
				},
			},
		},
	}
	// Set Webserver instance as the owner and controller
	ctrl.SetControllerReference(ws, dep, r.Scheme)
	return dep
}

// labelsForWebserver returns the labels for selecting the resources
// belonging to the given webserver CR name.
func labelsForWebserver(name string) map[string]string {
	return map[string]string{"app": "webserver", "webserver_cr": name}
}

func getLatency() json.Number {
	/** TODO:
	* 1. Store time before
	* 2. ping webserver pod
	* 3. Get time after
	* 4. Calculate latency
	* 5. Return latency
	 */
	min := 0.0
	max := 5.0
	latencyFloat := min + rand.Float64()*(max-min)
	latencyJSONNumber := json.Number(strconv.FormatFloat(latencyFloat, 'f', 4, 64))
	return latencyJSONNumber
}

func (r *WebserverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webserverv1alpha1.Webserver{}). // these two replaces Watches(...) function that is used in older documentation and guides/blogs. Might be other functions that I can also use!
		Owns(&appsv1.Deployment{}).          // these two replaces Watches(...) function that is used in older documentation and guides/blogs. Might be other functions that I can also use!
		Complete(r)
}
