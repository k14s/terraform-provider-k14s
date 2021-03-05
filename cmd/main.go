package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/k14s/terraform-provider-k14s/pkg/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: provider.Provider})
}
