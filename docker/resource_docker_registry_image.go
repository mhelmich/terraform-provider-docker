package docker

import (
	"fmt"
	"log"
	"strings"

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
	pullOpts := parseImageOptions(d.Get("target_name").(string))
	authConfig := meta.(*ProviderConfig).AuthConfigs

	// Use the official Docker Hub if a registry isn't specified
	if pullOpts.Registry == "" {
		pullOpts.Registry = "registry.hub.docker.com"
	} else {
		// Otherwise, filter the registry name out of the repo name
		pullOpts.Repository = strings.Replace(pullOpts.Repository, pullOpts.Registry+"/", "", 1)
	}

	if pullOpts.Registry == "registry.hub.docker.com" {
		// Docker prefixes 'library' to official images in the path; 'consul' becomes 'library/consul'
		if !strings.Contains(pullOpts.Repository, "/") {
			pullOpts.Repository = "library/" + pullOpts.Repository
		}
	}

	if pullOpts.Tag == "" {
		pullOpts.Tag = "latest"
	}

	var username string
	var password string

	if auth, ok := authConfig.Configs[normalizeRegistryAddress(pullOpts.Registry)]; ok {
		username = auth.Username
		password = auth.Password
	}

	digest, err := getImageDigest(pullOpts.Registry, pullOpts.Repository, pullOpts.Tag, username, password, false)
	if err != nil {
		digest, err = getImageDigest(pullOpts.Registry, pullOpts.Repository, pullOpts.Tag, username, password, true)
		if err != nil {
			return fmt.Errorf("Got error when attempting to fetch image version from registry: %s", err)
		}
	}

	log.Printf("[DEBUG] found image %v : digest %s", d.Get("target_name").(string), digest)

	d.SetId(digest)
	d.Set("sha256_digest", digest)
	return nil
}

func resourceDockerRegistryImageUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceDockerRegistryImageCreate(d, meta)
}

func resourceDockerRegistryImageDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
