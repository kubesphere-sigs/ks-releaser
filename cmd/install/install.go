package install

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type installCmdOpt struct {
	version string
}

// NewCommand creates a command to install ks-releaser
func NewCommand() (cmd *cobra.Command) {
	opt := &installCmdOpt{}

	cmd = &cobra.Command{
		Use:   "install",
		Short: "Install ks-releaser via https://github.com/kubesphere-sigs/ks-releaser-operator",
		RunE:  opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.version, "version", "v", "v0.0.13",
		"The version of target ks-releaser. Get the full version list from https://github.com/kubesphere-sigs/ks-releaser/pkgs/container/ks-releaser")
	return
}

func (o *installCmdOpt) runE(cmd *cobra.Command, args []string) (err error) {
	KubernetesConfigFlags := genericclioptions.NewConfigFlags(false)
	var config *rest.Config
	var client dynamic.Interface
	if config, err = KubernetesConfigFlags.ToRESTConfig(); err == nil {
		client, err = dynamic.NewForConfig(config)
	}

	var controllerObj *unstructured.Unstructured
	if controllerObj, err = getObjectFromYaml(fmt.Sprintf(`apiVersion: devops.kubesphere.io/v1alpha1
kind: ReleaserController
metadata:
  name: releasercontroller-default
spec:
  image: "ghcr.io/kubesphere-sigs/ks-releaser"
  version: "%s"
  webhook: false`, o.version)); err != nil {
		return
	}
	_, err = client.Resource(getReleaserControllerSchema()).Namespace("default").Create(context.TODO(), controllerObj, v1.CreateOptions{})
	return
}

func getReleaserControllerSchema() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "devops.kubesphere.io",
		Version:  "v1alpha1",
		Resource: "releasercontrollers",
	}
}

func getObjectFromYaml(yamlText string) (obj *unstructured.Unstructured, err error) {
	obj = &unstructured.Unstructured{}
	err = yaml.Unmarshal([]byte(yamlText), obj)
	return
}
