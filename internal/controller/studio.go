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
	studioPort = 3000
)

func (r *SupabaseReconciler) CreateStudio(ctx context.Context, supabase *supabasecomv1.Supabase) error {
	dep := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: supabase.Namespace,
		Name:      supabase.Name + "-studio",
	}, dep)

	if err != nil && k8serrors.IsNotFound(err) {
		dep = r.GetStudioDep(supabase)
		r.Log.Info("creating studio deployment")
		err := r.Create(ctx, dep)
		if err != nil {
			return fmt.Errorf("error creating studio deployment: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error getting studio deployment: %w", err)
	}

	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: supabase.Namespace,
		Name:      supabase.Name + "-studio",
	}, svc)

	if err != nil && k8serrors.IsNotFound(err) {
		svc = r.GetStudioSVC(supabase)
		r.Log.Info("creating studio service")
		err := r.Create(ctx, svc)
		if err != nil {
			return fmt.Errorf("error creating studio service: %w", err)
		}

		return err
	} else if err != nil {
		return fmt.Errorf("error getting studio service: %w", err)
	}

	return nil
}

func (r *SupabaseReconciler) GetStudioDep(supabase *supabasecomv1.Supabase) *appsv1.Deployment {
	ls := common.CreateLabels(supabase.Name, "studio")
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            supabase.Name + "-studio",
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
						Image: "supabase/studio:20240101-8e4a094",
						Name:  "studio",
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: studioPort,
							},
						},
						Env: []corev1.EnvVar{
							// {Name: "POSTGRES_PASSWORD", Value: "example"},
							// {Name: "POSTGRES_PASSWORD", Value: "example"},
						},
					}},
				},
			},
		},
	}
}

func (r *SupabaseReconciler) GetStudioSVC(supabase *supabasecomv1.Supabase) *corev1.Service {
	ls := common.CreateLabels(supabase.Name, "studio")
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            supabase.Name + "-studio",
			Namespace:       supabase.Namespace,
			OwnerReferences: common.CreateOwnerReferences(supabase),
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					TargetPort: intstr.IntOrString{IntVal: studioPort},
					Port:       studioPort,
				},
			},
		},
	}
}
