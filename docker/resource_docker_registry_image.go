package docker

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	if strings.Index(targetName, ":") == -1 {
		targetName = targetName + ":latest"
	}

	digest := data.DockerImages[targetName].ID
	d.SetId(digest)
	d.Set("sha256_digest", digest)
	return nil
}

func resourceDockerRegistryImageRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ProviderConfig).DockerClient

	// TODO - only look for the image we need
	// we don't have to list all images...
	iss, err := client.ImageList(context.TODO(), types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	targetName := d.Get("target_name").(string)
	if strings.Index(targetName, ":") == -1 {
		targetName = targetName + ":latest"
	}

	var theDigestIWannaRelease string
	for _, is := range iss {
		for _, tag := range is.RepoTags {
			if tag == targetName {
				theDigestIWannaRelease = is.ID
			}
		}
	}

	if theDigestIWannaRelease == "" {
		return fmt.Errorf("Can't find docker image: %s", targetName)
	}

	log.Printf("[DEBUG] found image: %s %s", targetName, theDigestIWannaRelease)
	d.SetId(theDigestIWannaRelease)
	d.Set("sha256_digest", theDigestIWannaRelease)
	return nil
}

func resourceDockerRegistryImageUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceDockerRegistryImageCreate(d, meta)
}

func resourceDockerRegistryImageDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
