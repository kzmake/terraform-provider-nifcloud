package nifcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The access key for API operations.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The secret key for API operations.",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The endpoint for API operations.",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The region where Nifcloud operations will take place.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			// "nifcloud_instance": dataSourceInstance(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"nifcloud_instance": resourceInstance(),
			"nifcloud_network":  resourceNetwork(),
			"nifcloud_keypair":  resourceKeyPair(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		AccessKey: d.Get("access_key").(string),
		SecretKey: d.Get("secret_key").(string),
		Endpoint:  d.Get("endpoint").(string),
		Region:    d.Get("region").(string),
	}

	return config.Client()
}
