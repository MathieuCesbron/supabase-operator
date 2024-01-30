package controller

import (
	"context"
	"fmt"

	supabasecomv1 "github.com/MathieuCesbron/supabase-operator/api/v1"
	"github.com/MathieuCesbron/supabase-operator/internal/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *SupabaseReconciler) CreateDatabase(ctx context.Context, supabase *supabasecomv1.Supabase) error {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: supabase.Namespace,
		Name:      supabase.Name + "-database",
	}, found)

	if err != nil && k8serrors.IsNotFound(err) {
		dep := r.GetDBDManifest(supabase)
		r.Log.Info("creating database deployment")
		err := r.Create(ctx, dep)
		if err != nil {
			return fmt.Errorf("error creating database deployment: %w", err)
		}

		return err
	} else if err != nil {
		return fmt.Errorf("error getting database deployment: %w", err)
	}

	return nil
}

func (r *SupabaseReconciler) GetDBDManifest(supabase *supabasecomv1.Supabase) *appsv1.Deployment {
	ls := common.CreateLabels(supabase.Name, "database")
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            supabase.Name + "-database",
			Namespace:       supabase.Namespace,
			OwnerReferences: common.CreateOwnerReferences(supabase),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "supabase/postgres:latest",
						Name:  "postgres",
						Env: []corev1.EnvVar{
							{Name: "POSTGRES_HOST", Value: "/var/run/postgresql"},
							{Name: "POSTGRES_PORT", Value: "5432"},
							{Name: "POSTGRES_USER", Value: "postgres"},
							{Name: "POSTGRES_PASSWORD", Value: "example"},
							{Name: "POSTGRES_DB", Value: "postgres"},
							{Name: "JWT_SECRET", Value: "example"},
							{Name: "JWT_EXP", Value: "example"},
						},
					}},
				},
			},
		},
	}
}
