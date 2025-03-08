// dir structure for proxymanager nginx configs
// NGINX_DIR = /etc/nginx
// $NGINX_DIR/sites-available/                    - normal proxy sites
// $NGINX_DIR/sites-available/k8s_$CLUSTER_NAME/  - k8s loadbalanced sites
package proxy

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"nickneal.dev/go-proxymanager/loadbalancer"
	"nickneal.dev/go-proxymanager/utils/nginx"
	"nickneal.dev/go-proxymanager/utils/settings"
	"nickneal.dev/go-proxymanager/utils/validate"
)

type Site struct {
	Name    string
	Enabled bool
}

func GetNginxDir() string {
	return settings.LoadConfig().Proxy.NginxDir
}

func DirectoryExist(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}

	if file.IsDir() {
		return true
	}

	return false
}

func GetAvailableConfigDir(cluster string) string {
	if cluster != "" {
		return GetNginxDir() + "/sites-available/k8s_" + cluster
	}

	return GetNginxDir() + "/sites-available"

}

func GetEnabledConfigDir() string {
	return GetNginxDir() + "/sites-enabled"
}

func GetEnabledSites() ([]string, error) {
	enabledConfigDir := GetEnabledConfigDir()
	enabledEntries, eErr := os.ReadDir(enabledConfigDir)
	if eErr != nil {
		fmt.Println(eErr)
		return nil, eErr
	}

	var siteArray []string
	for _, e := range enabledEntries {
		// check if not dir, and last 5 chars of name == .conf
		if !e.IsDir() && e.Name()[len(e.Name())-5:] == ".conf" {
			// string .conf from end of string and add to array
			siteArray = append(siteArray, []string{e.Name()[:len(e.Name())-5]}...)
		}
	}

	return siteArray, nil
}

func GetAvailableSites(cluster string) ([]string, error) {
	availableEntries, aErr := os.ReadDir(GetAvailableConfigDir(cluster))
	if aErr != nil {
		fmt.Println(aErr)
		return nil, aErr
	}

	var siteArray []string
	for _, e := range availableEntries {
		// check if not dir, and last 5 chars of name == .conf
		if !e.IsDir() && e.Name()[len(e.Name())-5:] == ".conf" {
			// string .conf from end of string and add to array
			siteArray = append(siteArray, []string{e.Name()[:len(e.Name())-5]}...)
		}
	}

	return siteArray, nil
}

func ClusterExists(cluster string) bool {
	return DirectoryExist(GetAvailableConfigDir(cluster))
}

func SiteExistsInCluster(cluster string, hostname string) bool {
	availableSites, _ := GetAvailableSites(cluster)
	for _, a := range availableSites {
		if a == hostname {
			return true
		}
	}
	return false

}

func SiteEnabled(hostname string) bool {
	enabledSites, _ := GetEnabledSites()
	for _, e := range enabledSites {
		if e == hostname {
			return true
		}
	}

	return false
}

func RestartNginx() bool {
	// restart nginx, if error, restore hosts file.
	if !settings.CheckDevMode() {
		nginxerr := nginx.RestartNginx()
		if nginxerr != nil {
			return false
		}
	}

	return true
}

func List(cluster string) {
	// make cluster lowercase
	cluster = strings.ToLower(cluster)

	if !ClusterExists(cluster) {
		fmt.Printf("cluster '%v' does not exist.\n", cluster)
		return
	}

	availableSites, asErr := GetAvailableSites(cluster)
	if asErr != nil {
		fmt.Println(asErr)
		return
	}

	if availableSites == nil {
		output := "No sites available"
		if cluster != "" {
			output = output + " in cluster '" + cluster + "'"
		}
		fmt.Println(output)
		return
	}

	var output []Site
	for _, a := range availableSites {
		var site Site
		site.Name = a
		site.Enabled = SiteEnabled(a)
		output = append(output, []Site{site}...)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	fmt.Fprintln(w, "Site\tEnabled")
	for _, s := range output {
		formattedstring := fmt.Sprintf("%v\t%v", s.Name, s.Enabled)
		fmt.Fprintln(w, formattedstring)
	}
	w.Flush()
}

func Enable(cluster string, hostname string) {
	// make sure args are lowercase
	cluster = strings.ToLower(cluster)
	hostname = strings.ToLower(hostname)

	if !ClusterExists(cluster) {
		fmt.Printf("Cluster '%v' does not exist.\n", cluster)
		return
	}

	if !SiteExistsInCluster(cluster, hostname) {
		if cluster == "" {
			fmt.Printf("Site '%v' does not exist.\n", hostname)
		} else {
			fmt.Printf("Site '%v' does not exist in cluster '%v'.\n", hostname, cluster)
		}
		return
	}

	if SiteEnabled(hostname) {
		fmt.Printf("Site '%v' is already enabled.\n", hostname)
		return
	}

	// create symlink
	sourcePath := GetAvailableConfigDir(cluster) + "/" + hostname + ".conf"
	destinationPath := GetEnabledConfigDir() + "/" + hostname + ".conf"
	err := os.Symlink(sourcePath, destinationPath)
	if err != nil {
		fmt.Printf("There was an error enabling '%v'.\n", hostname)
	}

	// restart nginx
	if !RestartNginx() {
		rmErr := os.Remove(destinationPath)
		if rmErr != nil {
			fmt.Println("There was an error reverting changes:", rmErr)
			fmt.Println("be sure to delete symlink before restarting nginx:", destinationPath)
		}
		fmt.Println("There was an error in nginx config.")
		return
	}

	// finished
	fmt.Printf("'%v' enabled.\n", hostname)

}

func Disable(cluster string, hostname string) {
	// make sure args are lowercase
	cluster = strings.ToLower(cluster)
	hostname = strings.ToLower(hostname)

	if !ClusterExists(cluster) {
		fmt.Printf("Cluster '%v' does not exist.\n", cluster)
		return
	}

	if !SiteExistsInCluster(cluster, hostname) {
		if cluster == "" {
			fmt.Printf("Site '%v' does not exist.\n", hostname)
		} else {
			fmt.Printf("Site '%v' does not exist in cluster '%v'.\n", hostname, cluster)
		}
		return
	}

	if !SiteEnabled(hostname) {
		fmt.Printf("Site '%v' is already disabled.\n", hostname)
		return
	}

	// remove symlink
	destinationPath := GetEnabledConfigDir() + "/" + hostname + ".conf"

	err := os.Remove(destinationPath)
	if err != nil {
		fmt.Printf("There was an error disabling '%v'.\n", hostname)
	}

	// restart nginx
	if !RestartNginx() {
		fmt.Println("There was an issue restarting nginx.")
		fmt.Println("Your site will be disabled after the next nginx restart.")
		return
	}

	// finished
	fmt.Printf("'%v' disabled.\n", hostname)
}

func Remove(cluster string, hostname string) {
	// make sure args are lowercase
	cluster = strings.ToLower(cluster)
	hostname = strings.ToLower(hostname)

	if !ClusterExists(cluster) {
		fmt.Printf("Cluster '%v' does not exist.\n", cluster)
		return
	}

	if !SiteExistsInCluster(cluster, hostname) {
		if cluster == "" {
			fmt.Printf("Site '%v' does not exist.\n", hostname)
		} else {
			fmt.Printf("Site '%v' does not exist in cluster '%v'.\n", hostname, cluster)
		}
		return
	}

	if SiteEnabled(hostname) {
		fmt.Printf("Site '%v' is enabled. Please disable before removing.\n", hostname)
		return
	}

	// remove symlink
	sourcePath := GetAvailableConfigDir(cluster) + "/" + hostname + ".conf"

	err := os.Remove(sourcePath)
	if err != nil {
		fmt.Printf("There was an error removing '%v'.\n", hostname)
	}

	// finished
	fmt.Printf("'%v' removed.\n", hostname)

}

// for new site only
func SiteExists(hostname string) bool {
	baseDir := GetAvailableConfigDir("")
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return true
	}

	for _, e := range entries {
		if e.IsDir() && regexp.MustCompile("^k8s_.*").MatchString(e.Name()) {
			baseClusterDir := baseDir + "/" + e.Name()
			clusterEntries, cErr := os.ReadDir(baseClusterDir)
			if cErr != nil {
				return true
			}

			for _, c := range clusterEntries {
				if !c.IsDir() && c.Name()[len(c.Name())-5:] == ".conf" && c.Name()[:len(c.Name())-5] == hostname {
					return true
				}
			}
		}

		if !e.IsDir() && e.Name()[len(e.Name())-5:] == ".conf" && e.Name()[:len(e.Name())-5] == hostname {
			return true
		}
	}

	return false
}

// for new site only
func CreateSiteConfig(filePath string, fileLines []string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, str := range fileLines {
		_, err = writer.WriteString(str + "\n")
		if err != nil {
			//fmt.Println("Error writing to file:", err)
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		//fmt.Println("Error flushing writer:", err)
		return err
	}

	return nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func New(cluster string,
	hostname string,
	ipAddress string,
	port string,
	uri string,
	ssl bool,
	sslBypassFirewall bool,
	proxySsl bool,
	proxySslVerifyOff bool) {

	// make sure args are lowercase
	cluster = strings.ToLower(cluster)
	hostname = strings.ToLower(hostname)

	// check if hostname config exists on web server
	if SiteExists(hostname) {
		fmt.Printf("Site '%v' is already in use on this server.\n", hostname)
		return
	}

	// if cluster isn't specified, verify IP address
	if cluster == "" && (ipAddress == "" || !validate.ValidateIPAddress(ipAddress)) {
		fmt.Println("IP Address not valid:", ipAddress)
		return
	}

	if port != "" && !validate.ValidatePort(port) {
		fmt.Printf("Port '%v' is invalid. please specify a port in the following range: 1024-49151\n", port)
		return
	}

	if uri != "" && !validate.ValidateUri(uri) {
		fmt.Printf("Uri '%v' is invalid.\nA uri must start with a '/' and only contain the following characters: a-z, A-Z, 0-9, /, -, _, ., and ~\n", uri)
		return
	}

	// define config
	var config string
	var backend string

	// run check only of cluster is defined
	if cluster != "" {
		//check if cluster exists
		if !ClusterExists(cluster) {
			fmt.Printf("Cluster '%v' does not exist.\n", cluster)
			return
		}

		// check if there are any nodes in the cluster
		if loadbalancer.GetClusterNodeCount(cluster) == 0 {
			fmt.Printf("Cluster '%v' has no assigned nodes.\n", cluster)
			return
		}

		if port == "" {
			fmt.Printf("no port was specified.")
			return
		}

		ipAddress = GetMD5Hash(hostname)
		config = settings.LoadConfig().Proxy.K8sProxyConfig

		var upstreamNodes string
		for _, str := range loadbalancer.GetClusterNodes(cluster) {
			upstreamNodes = upstreamNodes + "\tserver " + str + ":" + port + ";\n"
		}

		config = strings.Replace(config, "%UPSTREAM_NAME%", ipAddress, 1)
		config = strings.Replace(config, "%UPSTREAM_NODES%", upstreamNodes, 1)

	} else {
		// load config for non-k8s resource
		config = settings.LoadConfig().Proxy.ProxyConfig
	}

	// config params
	backend = "http://" + ipAddress
	if proxySsl {
		backend = "https://" + ipAddress
	}

	if cluster == "" && port != "" {
		backend = backend + ":" + port
	}

	if uri != "" {
		backend = backend + uri
	}

	verifyBackendSsl := ""
	if proxySsl && proxySslVerifyOff {
		verifyBackendSsl = "proxy_ssl_verify off;"
	}

	// perpare config
	config = strings.Replace(config, "%HOSTNAME%", hostname, 1)
	config = strings.Replace(config, "%BACKEND%", backend, 1)
	config = strings.Replace(config, "%PROXY_SSL_VERIFY_OFF%", verifyBackendSsl, 1)

	// prepare for writing file.
	configDir := GetAvailableConfigDir(cluster)
	configLines := strings.Split(config, "\n")
	filePath := configDir + "/" + hostname + ".conf"

	err := CreateSiteConfig(filePath, configLines)
	if err != nil {
		fmt.Println("There was an issue creating the site config.", err)
		return
	}

	if cluster == "" {
		fmt.Printf("Site '%v' created.\n", hostname)
	} else {
		fmt.Printf("Site '%v' created in cluster '%v'.\n", hostname, cluster)
	}

}
