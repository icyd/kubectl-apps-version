package plugin

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_skipNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		exclude   []string
		expected  bool
	}{
		{
			name:      "shouldn't skip",
			namespace: "default",
			exclude:   []string{"kube.*"},
			expected:  false,
		},
		{
			name:      "shouldn't skip in list",
			namespace: "default",
			exclude:   []string{"kube.*", "other.*", "defe.*"},
			expected:  false,
		},
		{
			name:      "excluded",
			namespace: "kube-system",
			exclude:   []string{"kube.*"},
			expected:  true,
		},
		{
			name:      "exact match",
			namespace: "kube-system",
			exclude:   []string{"kube-system"},
			expected:  true,
		},
		{
			name:      "multiple matches",
			namespace: "kube-system",
			exclude:   []string{"ku.*", "kube.*", "kube-system"},
			expected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := skipNamespace(test.namespace, test.exclude)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_processMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata metav1.ObjectMeta
		expected Metadata
	}{
		{
			name:     "empty metadata",
			metadata: metav1.ObjectMeta{},
			expected: Metadata{},
		},
		{
			name: "complete metadata",
			metadata: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "Helm",
					"helm.sh/chart":                "nginx-1.19",
					"argocd.argoproj.io/instance":  "test-app",
				},
			},
			expected: Metadata{
				ManagedBy: "Helm",
				HelmChart: "nginx-1.19",
				ArgoCDApp: "test-app",
			},
		},
		{
			name: "missing managed-by",
			metadata: metav1.ObjectMeta{
				Labels: map[string]string{
					"helm.sh/chart":               "nginx-1.19",
					"argocd.argoproj.io/instance": "test-app",
				},
			},
			expected: Metadata{},
		},
		{
			name: "with managed-by",
			metadata: metav1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "Helm",
				},
			},
			expected: Metadata{
				ManagedBy: "Helm",
				HelmChart: "",
				ArgoCDApp: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := processMetadata(test.metadata)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_processContainers(t *testing.T) {
	tests := []struct {
		name       string
		containers []corev1.Container
		expected   []ContainerImage
	}{
		{
			name:       "empty containers",
			containers: []corev1.Container{},
			expected:   nil,
		},
		{
			name: "one",
			containers: []corev1.Container{
				{
					Name:    "ImageOne",
					Image:   "nginx:1.18",
					Command: []string{"sleep", "1d"},
				},
			},
			expected: []ContainerImage{
				{
					Name:  "ImageOne",
					Image: "nginx:1.18",
				},
			},
		},
		{
			name: "one and empty",
			containers: []corev1.Container{
				{
					Name:    "ImageOne",
					Image:   "nginx:1.18",
					Command: []string{"sleep", "1d"},
				},
				{},
			},
			expected: []ContainerImage{
				{
					Name:  "ImageOne",
					Image: "nginx:1.18",
				},
			},
		},
		{
			name: "one with empty name",
			containers: []corev1.Container{
				{
					Image:   "nginx:1.18",
					Command: []string{"sleep", "1d"},
				},
				{},
			},
			expected: nil,
		},
		{
			name: "one with empty image",
			containers: []corev1.Container{
				{
					Name:    "nginx",
					Command: []string{"sleep", "1d"},
				},
				{},
			},
			expected: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := processContainers(test.containers)
			assert.Equal(t, test.expected, actual)
		})
	}
}
