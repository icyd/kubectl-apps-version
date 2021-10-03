package plugin

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type dsConditions struct {
	conditions []appsv1.DaemonSetCondition
}

func dsInNamespace(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, namespace string) ([]AppVersion, error) {
	apps := []AppVersion{}
	daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list deploys: %v", err)
	}
	for _, ds := range daemonsets.Items {
		app := AppVersion{
			Kind:                 "DaemonSets",
			Name:                 ds.Name,
			Namespace:            ds.Namespace,
			InitContainersImages: processContainers(ds.Spec.Template.Spec.InitContainers),
			ContainersImages:     processContainers(ds.Spec.Template.Spec.Containers),
			LastStatus:           AppStatus{},
			Metadata:             processMetadata(ds.ObjectMeta),
		}
		apps = append(apps, app)
	}

	outputCh <- fmt.Sprintf("Namespace %s", namespace)
	return apps, nil
}

func dsAllNamespaces(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, excludedNs []string) ([]AppVersion, error) {
	apps := []AppVersion{}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, namespace := range namespaces.Items {
		if skipNamespace(namespace.Name, excludedNs) {
			continue
		}

		app, err := dsInNamespace(clientset, outputCh, ctx, namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("Error: %v", err)
		}

		apps = append(apps, app...)
	}

	return apps, nil
}
