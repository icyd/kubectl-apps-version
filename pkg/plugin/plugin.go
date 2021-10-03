package plugin

import (
	"context"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var header = []metav1.TableColumnDefinition{
	{Name: "Kind", Type: "String"},
	{Name: "Namespace", Type: "String"},
	{Name: "Name", Type: "String"},
	{Name: "Image", Type: "String"},
	{Name: "Managed_By", Type: "String"},
	{Name: "Last_Updated", Type: "String"},
}

type PluginConfig struct {
	Namespace           string
	AllNamespaces       bool
	ExcludedNs          []string
	IncludeStatefulSets bool
	IncludeDaemonSets   bool
}

type ContainerImage struct {
	Name  string
	Image string
}

type AppStatus struct {
	LastUpdateTime string
	Reason         string
	Type           string
	Status         string
}

type Metadata struct {
	ManagedBy string
	HelmChart string
	ArgoCDApp string
}

type AppVersion struct {
	Namespace            string
	Kind                 string
	Name                 string
	InitContainersImages []ContainerImage
	ContainersImages     []ContainerImage
	LastStatus           AppStatus
	Metadata             Metadata
}

type AppVersions []AppVersion

func RunPlugin(configFlags *genericclioptions.ConfigFlags, outputCh chan string, ctx context.Context, pluginConfig *PluginConfig) (AppVersions, error) {
	var result AppVersions
	var namespace string

	if !pluginConfig.AllNamespaces && pluginConfig.Namespace == "" {
		clientCfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			return nil, fmt.Errorf("failed to read default configuration: %w", err)
		}

		namespace = clientCfg.Contexts[clientCfg.CurrentContext].Namespace
		if namespace == "" {
			namespace = "default"
		}

	} else if pluginConfig.AllNamespaces {
		namespace = ""
	} else {
		namespace = pluginConfig.Namespace
	}

	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	result, err = deploysInNamespace(clientset, outputCh, ctx, namespace)
	if err != nil {
		return nil, err
	}

	if pluginConfig.IncludeStatefulSets {
		apps, err := stsInNamespace(clientset, outputCh, ctx, namespace)
		if err != nil {
			return nil, err
		}
		result = append(result, apps...)
	}

	if pluginConfig.IncludeDaemonSets {
		apps, err := dsInNamespace(clientset, outputCh, ctx, namespace)
		if err != nil {
			return nil, err
		}
		result = append(result, apps...)
	}

	return result, nil
}

func (apps AppVersions) Print(w io.Writer) {
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	rows := []metav1.TableRow{}

	for _, app := range apps {
		for _, c := range app.InitContainersImages {
			row := metav1.TableRow{
				Cells: []interface{}{
					app.Kind,
					app.Namespace,
					app.Name,
					c.Image,
					app.Metadata.ManagedBy,
					app.LastStatus.LastUpdateTime,
				},
			}
			rows = append(rows, row)
		}
	}

	for _, app := range apps {
		for _, c := range app.ContainersImages {
			row := metav1.TableRow{
				Cells: []interface{}{
					app.Kind,
					app.Namespace,
					app.Name,
					c.Image,
					app.Metadata.ManagedBy,
					app.LastStatus.LastUpdateTime,
				},
			}
			rows = append(rows, row)
		}
	}

	table := &metav1.Table{
		ColumnDefinitions: header,
		Rows:              rows,
	}
	printer.PrintObj(table, w)
}
