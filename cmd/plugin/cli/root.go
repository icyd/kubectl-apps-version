package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/icyd/kubectl-apps-version/pkg/logger"
	"github.com/icyd/kubectl-apps-version/pkg/plugin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-spin"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl-apps-version",
		Short:         "short description",
		Long:          `Long description.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()

			s := spin.New()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			namespaceName := make(chan string, 1)
			go func() {
				lastNamespaceName := ""
				for {
					select {
					case <-ctx.Done():
						fmt.Printf("\r")
						return
					case n := <-namespaceName:
						lastNamespaceName = n
					case <-time.After(time.Millisecond * 100):
						if lastNamespaceName == "" {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s", s.Next())
						} else {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s (%s)", s.Next(), lastNamespaceName)
						}
					}
				}
			}()

			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				log.Error(err)
			}

			allNamespaces, err := cmd.Flags().GetBool("all-namespaces")
			if err != nil {
				log.Error(err)
			}

			sts, err := cmd.Flags().GetBool("statefulsets")
			if err != nil {
				log.Error(err)
			}

			ds, err := cmd.Flags().GetBool("daemonsets")
			if err != nil {
				log.Error(err)
			}

			excludedNs, err := cmd.Flags().GetStringSlice("exclude")
			if err != nil {
				log.Error(err)
			}

			pluginConfig := plugin.PluginConfig{
				Namespace:           namespace,
				AllNamespaces:       allNamespaces,
				ExcludedNs:          excludedNs,
				IncludeStatefulSets: sts,
				IncludeDaemonSets:   ds,
			}

			apps, err := plugin.RunPlugin(KubernetesConfigFlags, namespaceName, ctx, &pluginConfig)
			if err != nil {
				return errors.Unwrap(err)
			}

			apps.Print(os.Stdout)

			log.Info("")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	cmd.PersistentFlags().BoolP("all-namespaces", "A", false, "All namespaces")
	cmd.Flags().StringSlice("exclude", []string{"kube-system"}, "List of namespaces regexp to exclude. If namespace is given this option is ignored")
	cmd.Flags().Bool("statefulsets", false, "Include StatefulSets in search")
	cmd.Flags().Bool("daemonsets", false, "Include DaemonSets in search")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
