package common

import (
	supabasecomv1 "github.com/MathieuCesbron/supabase-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateLabels(name, role string) map[string]string {
	return map[string]string{
		"app":  "supabase",
		"cr":   name,
		"role": role,
	}
}

func CreateOwnerReferences(supabase *supabasecomv1.Supabase) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion: supabase.APIVersion,
			Kind:       supabase.Kind,
			Name:       supabase.Name,
			UID:        supabase.UID,
		},
	}
}
