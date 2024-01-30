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
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	DBPort = 5432
)

func (r *SupabaseReconciler) CreateDatabase(ctx context.Context, supabase *supabasecomv1.Supabase) error {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: supabase.Namespace,
		Name:      supabase.Name + "-database",
	}, found)

	if err != nil && k8serrors.IsNotFound(err) {
		dep := r.GetDBDep(supabase)
		r.Log.Info("creating database deployment")
		err := r.Create(ctx, dep)
		if err != nil {
			return fmt.Errorf("error creating database deployment: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error getting database deployment: %w", err)
	}

	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: supabase.Namespace,
		Name:      supabase.Name + "-database",
	}, svc)

	if err != nil && k8serrors.IsNotFound(err) {
		svc = r.GetDBSVC(supabase)
		r.Log.Info("creating database service")
		err := r.Create(ctx, svc)
		if err != nil {
			return fmt.Errorf("error creating database service: %w", err)
		}

		return err
	} else if err != nil {
		return fmt.Errorf("error getting database service: %w", err)
	}

	return nil
}

func (r *SupabaseReconciler) GetDBDep(supabase *supabasecomv1.Supabase) *appsv1.Deployment {
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
						Ports: []corev1.ContainerPort{
							{
								Name:          "postgres",
								ContainerPort: 5432,
							},
						},
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

func (r *SupabaseReconciler) GetDBSVC(supabase *supabasecomv1.Supabase) *corev1.Service {
	ls := common.CreateLabels(supabase.Name, "database")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            supabase.Name + "-database",
			Namespace:       supabase.Namespace,
			OwnerReferences: common.CreateOwnerReferences(supabase),
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:       "postgres",
					TargetPort: intstr.IntOrString{IntVal: DBPort},
					Port:       DBPort,
				},
			},
		},
	}
}
