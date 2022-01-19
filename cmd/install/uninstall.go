package install

import (
	"context"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// NewUninstallCmd creates a command to uninstall ks-releaser
func NewUninstallCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "uninstall",
		Aliases: []string{"delete", "remove", "del", "rm"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			KubernetesConfigFlags := genericclioptions.NewConfigFlags(false)
			var config *rest.Config
			var client dynamic.Interface
			if config, err = KubernetesConfigFlags.ToRESTConfig(); err == nil {
				client, err = dynamic.NewForConfig(config)
			}

			err = client.Resource(getReleaserControllerSchema()).Namespace("default").
				Delete(context.TODO(), "releasercontroller-default", v1.DeleteOptions{})
			return
		},
	}
	return
}
