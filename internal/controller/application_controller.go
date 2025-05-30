/*
Copyright 2025.

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

package controller

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "tip.io/api/v1alpha1"
)

// ApplicationReconciler reconciles an Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.tip.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.tip.io,resources=applications/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps.tip.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	app := &appsv1alpha1.Application{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		if errors.IsNotFound(err) {
			//resource not found
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	prefix := app.Spec.Foo + app.Spec.Bar
	existedPod := &v1.Pod{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: v1.NamespaceDefault,
		Name:      prefix + "-pod",
	}, existedPod); errors.IsNotFound(err) {
		//if not exist, then create
		existedPod = constructPod(prefix)
		//set ctrl reference
		if err = ctrl.SetControllerReference(app, existedPod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		err = r.Create(ctx, existedPod)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Application{}).
		Named("application").
		Complete(r)
}

func constructPod(namePrefix string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namePrefix + "-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{Containers: []v1.Container{
			{
				Name:            namePrefix + "-container",
				Image:           "busybox:latest",
				Command:         []string{"sleep", "3600"},
				ImagePullPolicy: v1.PullIfNotPresent,
			},
		}},
	}
}
