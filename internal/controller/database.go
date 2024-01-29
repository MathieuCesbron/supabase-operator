package controller

import (
	"context"
	"fmt"

	supabasecomv1 "github.com/MathieuCesbron/supabase-operator/api/v1"
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
		Name:      supabase.Name,
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
	ls := labelsForDatabase(supabase.Name)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      supabase.Name,
			Namespace: supabase.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: supabase.APIVersion,
					Kind:       supabase.Kind,
					Name:       supabase.Name,
					UID:        supabase.UID,
				},
			},
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
						Image: "supabase/postgres:14.1.0.105",
						Name:  "postgresql",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "realtime",
								MountPath: "/docker-entrypoint-initdb.d/migrations/99-realtime.sql",
							},
							{
								Name:      "webhooks",
								MountPath: "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql",
							},
							{
								Name:      "roles",
								MountPath: "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql",
							},
							{
								Name:      "jwt",
								MountPath: "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql",
							},
							{
								Name:      "data",
								MountPath: "/var/lib/postgresql/data",
							},
							{
								Name:      "logs",
								MountPath: "/docker-entrypoint-initdb.d/migrations/99-logs.sql",
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
					Volumes: []corev1.Volume{
						{
							Name: "realtime",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/realtime.sql"},
							},
						},
						{
							Name: "webhooks",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/webhooks.sql"},
							},
						},
						{
							Name: "roles",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/roles.sql"},
							},
						},
						{
							Name: "jwt",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/jwt.sql"},
							},
						},
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/data"},
							},
						},
						{
							Name: "logs",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/internal/controller/volumes/db/logs.sql"},
							},
						},
					},
				},
			},
		},
	}
}

func labelsForDatabase(name string) map[string]string {
	return map[string]string{
		"app": "supabase",
		"cr":  name,
	}
}
