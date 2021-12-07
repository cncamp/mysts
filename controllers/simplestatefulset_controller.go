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
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1alpha1 "github.com/cncamp/mysts/api/v1alpha1"
)

// SimpleStatefulsetReconciler reconciles a SimpleStatefulset object
type SimpleStatefulsetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.cncamp.io,resources=simplestatefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.cncamp.io,resources=simplestatefulsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.cncamp.io,resources=simplestatefulsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SimpleStatefulset object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *SimpleStatefulsetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	mylogger := log.FromContext(ctx)
	mylogger.Info("Reconcile SimpleStatefulset", "NamespacedName", req.NamespacedName)
	myss := appsv1alpha1.SimpleStatefulset{}
	if err := r.Client.Get(ctx, req.NamespacedName, &myss); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get %s", req.NamespacedName)
	}
	replicas := myss.Spec.Replicas
	if replicas == 0 {
		replicas = 2
	}
	image := myss.Spec.Image
	if image == "" {
		image = "nginx"
	}
	for i := 0; i < replicas; i++ {
		name := fmt.Sprintf("%s-%d", myss.Name, i)
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: myss.Namespace, Name: name}, &v1.Pod{}); err != nil {
			if errors.IsNotFound(err) {
				p := v1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: myss.Namespace,
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps.crane.io/v1alpha1",
								Kind:       "SimpleStatefulset",
								Name:       myss.Name,
								UID:        myss.UID,
							},
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: image,
								Name:  name,
							},
						},
					},
				}
				if err := r.Client.Create(ctx, &p); err != nil {
					mylogger.Info("Failed to create pod", "err", err)
					return ctrl.Result{}, fmt.Errorf("failed to create pod")
				} else {
					toBeUpdate := myss.DeepCopy()
					toBeUpdate.Status.AvailableReplicas += 1
					if err := r.Client.Status().Update(ctx, toBeUpdate); err != nil {
						mylogger.Info("Failed to update myss status", "err", err)
						return ctrl.Result{}, fmt.Errorf("failed to update myss status")
					}
				}
			} else {
				mylogger.Info("Failed to get pod", "err", err)
				return ctrl.Result{}, fmt.Errorf("failed to get pod")
			}
		} else {
			mylogger.Info("Pod exists", "name", name)
		}
	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SimpleStatefulsetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SimpleStatefulset{}).
		Complete(r)
}
