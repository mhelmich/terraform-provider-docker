package docker

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDockerRegistryImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceDockerRegistryImageCreate,
		Read:   resourceDockerRegistryImageRead,
		Update: resourceDockerRegistryImageUpdate,
		Delete: resourceDockerRegistryImageDelete,

		Schema: map[string]*schema.Schema{
			"source_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"target_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"sha256_digest": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDockerRegistryImageCreate(d *schema.ResourceData, meta interface{}) error {
	sourceName := d.Get("source_name").(string)
	targetName := d.Get("target_name").(string)
	client := meta.(*ProviderConfig).DockerClient
	authConfig := meta.(*ProviderConfig).AuthConfigs

	var data Data
	err := tagAndPushImage(&data, client, authConfig, sourceName, targetName)
	if err != nil {
		return err
	}

	digest := data.DockerImages[targetName+":latest"].ID
	d.SetId(digest)
	d.Set("sha256_digest", digest)
	return nil
}

func resourceDockerRegistryImageRead(d *schema.ResourceData, meta interface{}) error {
	return dataSourceDockerRegistryImageRead(d, meta)
}

func resourceDockerRegistryImageUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceDockerRegistryImageCreate(d, meta)
}

func resourceDockerRegistryImageDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
