package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/kzmake/terraform-provider-nifcloud/nifcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: nifcloud.Provider})
}
