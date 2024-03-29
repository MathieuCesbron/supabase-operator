/*
Copyright 2024.

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
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	supabasecomv1 "github.com/MathieuCesbron/supabase-operator/api/v1"
	"github.com/go-logr/logr"
)

// SupabaseReconciler reconciles a Supabase object
type SupabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=supabase.com,resources=supabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=supabase.com,resources=supabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=supabase.com,resources=supabases/finalizers,verbs=update
func (r *SupabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = log.FromContext(ctx).WithName("SupabaseReconciler")

	supabase := &supabasecomv1.Supabase{}
	err := r.Get(ctx, req.NamespacedName, supabase)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			r.Log.Info("supabase cr has been deleted")
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting supabase cr: %w", err)
	}

	err = r.CreateDatabase(ctx, supabase)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.CreateStudio(ctx, supabase)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SupabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&supabasecomv1.Supabase{}).
		Complete(r)
}
