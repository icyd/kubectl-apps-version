package plugin

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type deployConditions struct {
	conditions []appsv1.DeploymentCondition
}

func deploysInNamespace(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, namespace string) ([]AppVersion, error) {
	apps := []AppVersion{}

	deploys, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list deploys: %v", err)
	}
	for _, deploy := range deploys.Items {
		status := deployConditions{
			conditions: deploy.Status.Conditions,
		}
		app := AppVersion{
			Kind:                 "Deployment",
			Name:                 deploy.Name,
			Namespace:            deploy.Namespace,
			InitContainersImages: processContainers(deploy.Spec.Template.Spec.InitContainers),
			ContainersImages:     processContainers(deploy.Spec.Template.Spec.Containers),
			LastStatus:           status.processConditions(),
			Metadata:             processMetadata(deploy.ObjectMeta),
		}
		apps = append(apps, app)
	}

	outputCh <- fmt.Sprintf("Namespace %s", namespace)
	return apps, nil
}

func deploysAllNamespaces(clientset *kubernetes.Clientset, outputCh chan string, ctx context.Context, excludedNs []string) ([]AppVersion, error) {
	apps := []AppVersion{}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, namespace := range namespaces.Items {
		if skipNamespace(namespace.Name, excludedNs) {
			continue
		}

		app, err := deploysInNamespace(clientset, outputCh, ctx, namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("Error: %v", err)
		}

		apps = append(apps, app...)
	}

	return apps, nil
}

func (deploys deployConditions) processConditions() AppStatus {
	lastStatusIdx := 0

	for i, status := range deploys.conditions {
		if status.LastUpdateTime.Unix() > deploys.conditions[lastStatusIdx].LastUpdateTime.Unix() {
			lastStatusIdx = i
		}
	}
	return AppStatus{
		LastUpdateTime: deploys.conditions[lastStatusIdx].LastUpdateTime.String(),
		Reason:         deploys.conditions[lastStatusIdx].Reason,
		Type:           string(deploys.conditions[lastStatusIdx].Type),
		Status:         string(deploys.conditions[lastStatusIdx].Status),
	}
}
