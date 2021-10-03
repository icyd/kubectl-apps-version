package plugin

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

func skipNamespace(namespace string, excluded []string) (isExcluded bool) {
	isExcluded = false

	for _, excludedNs := range excluded {
		if match, err := regexp.Match(excludedNs, []byte(namespace)); match && err == nil {
			isExcluded = true
			return
		}
	}

	return
}

func processMetadata(metadata metav1.ObjectMeta) (data Metadata) {
	if val, ok := metadata.Labels["app.kubernetes.io/managed-by"]; ok {
		data.ManagedBy = val
	}

	if val, ok := metadata.Labels["helm.sh/chart"]; ok && data.ManagedBy == "Helm" {
		data.HelmChart = val
	}

	if val, ok := metadata.Labels["argocd.argoproj.io/instance"]; ok && data.ManagedBy != "" {
		data.ArgoCDApp = val
	}

	return
}

func processContainers(containers []corev1.Container) (result []ContainerImage) {
	for _, c := range containers {
		if c.Name == "" || c.Image == "" {
			continue
		}

		result = append(result, ContainerImage{
			Name:  c.Name,
			Image: c.Image,
		})
	}

	return
}
