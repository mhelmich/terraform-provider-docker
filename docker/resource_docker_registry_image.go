package docker

import (
	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDockerRegistryImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceDockerRegistryImageCreate,
		Read:   resourceDockerRegistryImageRead,
		Update: resourceDockerRegistryImageUpdate,
		Delete: resourceDockerRegistryImageDelete,

		Schema: map[string]*schema.Schema{
			"local_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"remote_name": &schema.Schema{
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
	localName := d.Get("local_name").(string)
	remoteName := d.Get("remote_name").(string)
	client := meta.(*ProviderConfig).DockerClient
	authConfig := meta.(*ProviderConfig).AuthConfigs
	return tagAndPushImage(nil, client, authConfig, localName, remoteName)
}

func resourceDockerRegistryImageRead(d *schema.ResourceData, meta interface{}) error {
	return dataSourceDockerRegistryImageRead(d, meta)
}

func resourceDockerRegistryImageUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDockerRegistryImageDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func getAuthToken(username string, password string) string {
	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON)
}
