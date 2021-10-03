package plugin

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func stsInNamespace(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, namespace string) ([]AppVersion, error) {
	apps := []AppVersion{}
	sts, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list statefulsets: %v", err)
	}
	for _, s := range sts.Items {
		app := AppVersion{
			Kind:                 "StatefulSet",
			Name:                 s.Name,
			Namespace:            s.Namespace,
			InitContainersImages: processContainers(s.Spec.Template.Spec.InitContainers),
			ContainersImages:     processContainers(s.Spec.Template.Spec.Containers),
			LastStatus:           AppStatus{},
			Metadata:             processMetadata(s.ObjectMeta),
		}
		apps = append(apps, app)
	}

	outputCh <- fmt.Sprintf("Namespace %s", namespace)
	return apps, nil
}

func stsAllNamespaces(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, excludedNs []string) ([]AppVersion, error) {
	apps := []AppVersion{}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, namespace := range namespaces.Items {
		if skipNamespace(namespace.Name, excludedNs) {
			continue
		}

		app, err := stsInNamespace(clientset, outputCh, ctx, namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("Error: %v", err)
		}

		apps = append(apps, app...)
	}

	return apps, nil
}
