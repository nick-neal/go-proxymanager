package settings

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LoadBalancer struct {
		HostsFile string `yaml:"hostsFile"`
	} `yaml:"loadBalancer"`

	Proxy struct {
		NginxDir       string `yaml:"nginxDir"`
		ProxyConfig    string `yaml:"proxyConfig"`
		K8sProxyConfig string `yaml:"k8sProxyConfig"`
	} `yaml:"proxy"`
}

func DefaultConfig() *Config {
	config := &Config{}
	config.LoadBalancer.HostsFile = "/etc/hosts"
	config.Proxy.NginxDir = "/etc/nginx"
	config.Proxy.ProxyConfig = `
	test
	test2
	test3
	`
	config.Proxy.K8sProxyConfig = `
	test k8s
	test cause
	why not
	`

	return config

}

func LoadConfig() *Config {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(GetConfigPath())
	if err != nil {
		return DefaultConfig()
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return DefaultConfig()
	}

	return config
}

func CheckDevMode() bool {
	return os.Getenv("PROXYMANAGER_DEV_MODE") == "true"
}

func GetConfigPath() string {
	if os.Getenv("PROXYMANAGER_CONFIG_PATH") != "" {
		return os.Getenv("PROXYMANAGER_CONFIG_PATH")
	}

	if CheckDevMode() {
		cwd, _ := os.Getwd()
		return cwd + "/etc/proxymanager.yml"
	}

	return "/etc/proxymanager.yml"
}
