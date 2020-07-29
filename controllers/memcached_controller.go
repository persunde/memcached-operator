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
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	cachev1alpha1 "github.com/example-inc/memcached-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
)

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch

func (r *MemcachedReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("memcached", req.NamespacedName)

	// your logic here

	ctx := context.TODO()
	// Lookup the Memcached instance for this reconcile request
	memcached := &cachev1alpha1.Memcached{}
	err := r.Get(ctx, req.NamespacedName, memcached)

	// Reconcile failed due to error - requeue
	if err != nil {
		log.Error(err, "Reconcile failed due to error - requeue")
		return ctrl.Result{}, err
	}

	// Requeue for any reason other than an error
	//return ctrl.Result{Requeue: true}, nil

	log.Error("Reconcile success - requeue after 5 seconds")
	// Reconcile for any reason other than an error after 5 seconds
	return ctrl.Result{RequeueAfter: time.Second * 5}, nil

	// Reconcile successful - don't requeue
	//return ctrl.Result{}, nil
}

func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Owns(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
