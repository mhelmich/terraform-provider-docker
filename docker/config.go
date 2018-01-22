package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	dc "github.com/mavogel/go-dockerclient"
)

// DockerConfig is the structure that stores the configuration to talk to a
// Docker API compatible host.
type DockerConfig struct {
	Host          string
	Ca            string
	Cert          string
	Key           string
	CertPath      string
	ForwardConfig []interface{}
}

// NewClient() returns a new Docker client.
func (c *DockerConfig) NewClient() (*dc.Client, error) {
	if c.Ca != "" || c.Cert != "" || c.Key != "" {
		if c.Ca == "" || c.Cert == "" || c.Key == "" {
			return nil, fmt.Errorf("ca_material, cert_material, and key_material must be specified")
		}

		if c.CertPath != "" {
			return nil, fmt.Errorf("cert_path must not be specified")
		}

		if c.ForwardConfig != nil {
			forwardConfig, err := parseForwardConfig(c.ForwardConfig)
			if err != nil {
				return nil, fmt.Errorf("Invalid forward config: %s", err)
			}
			return dc.NewTLSClientFromBytesWithForward(c.Host, []byte(c.Cert), []byte(c.Key), []byte(c.Ca), forwardConfig)
		}

		return dc.NewTLSClientFromBytes(c.Host, []byte(c.Cert), []byte(c.Key), []byte(c.Ca))
	}

	if c.CertPath != "" {
		// If there is cert information, load it and use it.
		ca := filepath.Join(c.CertPath, "ca.pem")
		cert := filepath.Join(c.CertPath, "cert.pem")
		key := filepath.Join(c.CertPath, "key.pem")

		if c.ForwardConfig != nil {
			forwardConfig, err := parseForwardConfig(c.ForwardConfig)
			if err != nil {
				return nil, fmt.Errorf("Invalid forward config: %s", err)
			}
			return dc.NewTLSClientWithForward(c.Host, cert, key, ca, forwardConfig)
		}

		return dc.NewTLSClient(c.Host, cert, key, ca)
	}

	if c.ForwardConfig != nil {
		forwardConfig, err := parseForwardConfig(c.ForwardConfig)
		if err != nil {
			return nil, fmt.Errorf("Invalid forward config: %s", err)
		}
		return dc.NewClientWithForward(c.Host, forwardConfig)
	}

	// If there is no cert information, then just return the direct client
	return dc.NewClient(c.Host)
}

// Data structure for holding data that we fetch from Docker.
type Data struct {
	DockerImages map[string]*dc.APIImages
}

// ProviderConfig for the custom registry provider
type ProviderConfig struct {
	DockerClient *dc.Client
	AuthConfigs  *dc.AuthConfigurations
}

// The registry address can be referenced in various places (registry auth, docker config file, image name)
// with or without the http(s):// prefix; this function is used to standardize the inputs
func normalizeRegistryAddress(address string) string {
	if !strings.HasPrefix(address, "https://") && !strings.HasPrefix(address, "http://") {
		return "https://" + address
	}
	return address
}

// Takes the given forward config and parses it into the dc.ForwardConfig
func parseForwardConfig(forwardConfigList []interface{}) (*dc.ForwardConfig, error) {
	forwardConfig := &dc.ForwardConfig{}

	if len(forwardConfigList) > 0 {
		fc := forwardConfigList[0].(map[string]interface{})
		jumpHostConfigs := make([]*dc.ForwardSSHConfig, 0)
		forwardConfig.JumpHostConfigs = jumpHostConfigs

		if v, ok := fc["bastion_host_config"]; ok {
			bastionHostConfig := &dc.ForwardSSHConfig{}
			bastionHostConfigRaw := v.(map[string]interface{})
			if v, ok := bastionHostConfigRaw["host"]; ok {
				bastionHostConfig.Address = v.(string)
			}
			if v, ok := bastionHostConfigRaw["user"]; ok {
				bastionHostConfig.User = v.(string)
			}
			if v, ok := bastionHostConfigRaw["password"]; ok {
				bastionHostConfig.Password = v.(string)
			}
			if v, ok := bastionHostConfigRaw["private_key_file"]; ok {
				bastionHostConfig.PrivateKeyFile = v.(string)
			}
			jumpHostConfigs = append(jumpHostConfigs, bastionHostConfig)
		}

		if v, ok := fc["end_host_config"]; ok {
			forwardConfig.EndHostConfig = &dc.ForwardSSHConfig{}
			endHostConfigRaw := v.(map[string]interface{})
			if v, ok := endHostConfigRaw["host"]; ok {
				forwardConfig.EndHostConfig.Address = v.(string)
			}
			if v, ok := endHostConfigRaw["user"]; ok {
				forwardConfig.EndHostConfig.User = v.(string)
			}
			if v, ok := endHostConfigRaw["password"]; ok {
				forwardConfig.EndHostConfig.Password = v.(string)
			}
			if v, ok := endHostConfigRaw["private_key_file"]; ok {
				forwardConfig.EndHostConfig.PrivateKeyFile = v.(string)
			}
		}

		if v, ok := fc["local_address"]; ok {
			forwardConfig.LocalAddress = v.(string)
		}

		if v, ok := fc["remote_address"]; ok {
			forwardConfig.RemoteAddress = v.(string)
		}

	}

	return forwardConfig, nil
}
