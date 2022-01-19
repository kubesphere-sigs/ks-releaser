package main

import (
	"github.com/kubesphere-sigs/ks-releaser/cli/install"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "ks-releaser-cli",
		Short: "CLI for install/uninstall ks-releaser",
	}
	root.AddCommand(install.NewCommand(), install.NewUninstallCmd())

	err := root.Execute()
	if err != nil {
		panic(err)
	}
}
