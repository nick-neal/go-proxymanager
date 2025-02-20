package loadbalancer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	nginx "nickneal.dev/go-proxymanager/utils/nginx"
	settings "nickneal.dev/go-proxymanager/utils/settings"
	validate "nickneal.dev/go-proxymanager/utils/validate"
)

func ReadHostsFileLines() ([]string, error) {
	// get location of hosts file
	hostsFilePath := settings.LoadConfig().LoadBalancer.HostsFile

	// open file
	file, err := os.Open(hostsFilePath)
	if err != nil {
		//fmt.Println("Error opening file:", err)
		return []string{}, err
	}
	defer file.Close()

	// read file lines to string array
	var fileLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		//fmt.Println("Error reading file:", err)
		return []string{}, err
	}
	// close file
	file.Close()
	return fileLines, nil
}

func WriteHostsFileLines(fileLines []string) error {
	// get location of hosts file
	hostsFilePath := settings.LoadConfig().LoadBalancer.HostsFile

	file, err := os.OpenFile(hostsFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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

func CreateClusterConfigDir(cluster string) error {
	return nil
}

func RemoveClusterConfigDir(cluster string, backup bool) error {
	return nil
}

func RestartNginx(fileLinesBackup []string) bool {
	// restart nginx, if error, restore hosts file.
	if !settings.CheckDevMode() {
		nginxerr := nginx.RestartNginx()
		if nginxerr != nil {
			fmt.Println("Nginx issue: ", nginxerr)
			fmt.Println("Restoring hosts file...")

			// write changes
			err := WriteHostsFileLines(fileLinesBackup)
			if err != nil {
				fmt.Println("There was an issue restoring file:", err)
				return false
			}

			fmt.Println("Hosts file successfully restored!")
			return false
		}
	}

	return true
}

func List() {
	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\("
	var clusters []string
	for _, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			cluster := strings.Split(strings.Split(line, "(")[1], ")")[0]
			clusters = append(clusters, []string{cluster}...)
		}
	}

	if len(clusters) > 0 {
		//fmt.Println("Clusters")
		for _, cluster := range clusters {
			fmt.Println(cluster)
		}
	} else {
		fmt.Println("No clusters exist.")
	}

}

func New(cluster string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)
	if !validate.ValidateClusterName(cluster) {
		fmt.Println("Invalid cluster name:", cluster)
		return
	}

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\("
	pattern2 := "^### LB_K8S_END"
	var lastOccurence int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			if cluster == strings.Split(strings.Split(line, "(")[1], ")")[0] {
				formattedString := fmt.Sprintf("Cluster '%v' already exists.", cluster)
				fmt.Println(formattedString)
				return
			}
		}

		// grab last index of cluster definition
		if regexp.MustCompile(pattern2).MatchString(line) {
			lastOccurence = index
		}
	}

	// add new cluster
	newLines := []string{"", "# DO NOT EDIT, USE proxymanager", "### LB_K8S(" + cluster + ")", "### LB_K8S_END"}
	var newFileContent []string
	if lastOccurence != 0 && lastOccurence != (len(fileLines)-1) {
		newFileContent = append(fileLines[:lastOccurence+1], append(newLines, fileLines[lastOccurence+1:]...)...)
	} else {
		newFileContent = append(fileLines, newLines...)
	}

	// rewrite hosts file
	writeErr := WriteHostsFileLines(newFileContent)
	if writeErr != nil {
		fmt.Println("There was an issue writing to file:", writeErr)
	}

	formattedString := fmt.Sprintf("Cluster '%v' created.", cluster)
	fmt.Println(formattedString)
}

func Remove(cluster string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	// backup filelines
	fileLinesBackup := make([]string, len(fileLines))
	_ = copy(fileLinesBackup, fileLines)

	// remove cluster
	if startLine != 0 && endLine != 0 {
		if GetClusterNodeCount(cluster) > 0 {
			fmt.Printf("Node count is higher than 0 on cluster '%v'. Remove canceled.\n", cluster)
			return
		}
		fileLines = append(fileLines[:startLine-2], fileLines[endLine+1:]...)
	} else {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	// write changes
	err := WriteHostsFileLines(fileLines)
	if err != nil {
		fmt.Println("There was an issue writing file:", err)
		return
	}

	// restart nginx, if error, restore hosts file.
	if !RestartNginx(fileLinesBackup) {
		return
	}

	formattedString := fmt.Sprintf("Cluster '%v' removed.", cluster)
	fmt.Println(formattedString)
}

func Add(cluster string, ipAddress string, host string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)

	// validate ip address
	if !validate.ValidateIPAddress(ipAddress) {
		fmt.Println("Invalid IP Address format. must be value between 0.0.0.0 - 255.255.255.255.")
		return
	}

	// make hostname lowercase and validate
	host = strings.ToLower(host)
	if !validate.ValidateHostName(host) {
		fmt.Println("Invalid Hostname format. Can only contain lowercase letters, numbers, hypens, and periods.")
		return
	}

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	// check whole hosts file for hostname
	for _, str := range fileLines {
		if regexp.MustCompile(host + ".?( |$)").MatchString(str) {
			formattedString := fmt.Sprintf("Host '%v' exists in hosts file.", host)
			fmt.Println(formattedString)
			return
		}
	}

	hosts := NewHostConfig(fileLines[startLine:endLine])
	// check IP Address
	if hosts.IPExists(ipAddress) {
		formattedString := fmt.Sprintf("IP Address '%v' already exists in cluster '%v'.", ipAddress, cluster)
		fmt.Println(formattedString)
		return
	}

	hosts.AddHost(host, ipAddress)
	fileLinesBackup := make([]string, len(fileLines))
	_ = copy(fileLinesBackup, fileLines)
	fileLines = append(fileLines[:startLine], append(hosts.ToArray(), fileLines[endLine:]...)...)

	//fmt.Println(fileLines[:startLine])
	//fmt.Println(fileLines[endLine:])

	// write changes
	err := WriteHostsFileLines(fileLines)
	if err != nil {
		fmt.Println("There was an issue writing file:", err)
		return
	}

	// restart nginx, if error, restore hosts file.
	if !RestartNginx(fileLinesBackup) {
		return
	}

	formattedString := fmt.Sprintf("Host '%v' added to cluster '%v'.", host, cluster)
	fmt.Println(formattedString)
}

func Del(cluster string, host string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)
	host = strings.ToLower(host)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	hosts := NewHostConfig(fileLines[startLine:endLine])
	if !hosts.HostExists(host) {
		formattedString := fmt.Sprintf("Host '%v' does not exist in cluster '%v'.", host, cluster)
		fmt.Println(formattedString)
		return
	}

	hosts.DelHost(host)
	fileLinesBackup := make([]string, len(fileLines))
	_ = copy(fileLinesBackup, fileLines)

	if len(hosts) == 0 {
		fileLines = append(fileLines[:startLine], fileLines[endLine:]...)
	} else {
		fileLines = append(fileLines[:startLine], append(hosts.ToArray(), fileLines[endLine:]...)...)
	}

	// write changes
	err := WriteHostsFileLines(fileLines)
	if err != nil {
		fmt.Println("There was an issue writing file:", err)
		return
	}

	// restart nginx, if error, restore hosts file.
	if !RestartNginx(fileLinesBackup) {
		return
	}

	formattedString := fmt.Sprintf("Host '%v' removed from cluster '%v'.", host, cluster)
	fmt.Println(formattedString)
}

func Status(cluster string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	if startLine != endLine {
		hosts := NewHostConfig(fileLines[startLine:endLine])
		hosts.PrintHosts()

	} else {
		formattedString := fmt.Sprintf("No hosts defined in cluster '%v'.", cluster)
		fmt.Println(formattedString)
	}
}

func Move(cluster string, fromHost string, toHost string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)
	fromHost = strings.ToLower(fromHost)
	toHost = strings.ToLower(toHost)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	hosts := NewHostConfig(fileLines[startLine:endLine])
	// check fromHost
	if !hosts.HostExists(fromHost) {
		formattedString := fmt.Sprintf("Host '%v' does not exist in cluster '%v'.", fromHost, cluster)
		fmt.Println(formattedString)
		return
	}

	// check toHost
	if !hosts.HostExists(toHost) {
		formattedString := fmt.Sprintf("Host '%v' does not exist in cluster '%v'.", toHost, cluster)
		fmt.Println(formattedString)
		return
	}

	hosterr := hosts.MoveTraffic(fromHost, toHost)
	if hosterr != nil {
		fmt.Println("There was an issue moving traffic:", hosterr)
		return
	}
	fileLinesBackup := make([]string, len(fileLines))
	_ = copy(fileLinesBackup, fileLines)
	fileLines = append(fileLines[:startLine], append(hosts.ToArray(), fileLines[endLine:]...)...)

	//fmt.Println(fileLines[:startLine])
	//fmt.Println(fileLines[endLine:])

	// write changes
	err := WriteHostsFileLines(fileLines)
	if err != nil {
		fmt.Println("There was an issue writing file:", err)
		return
	}

	// restart nginx, if error, restore hosts file.
	if !RestartNginx(fileLinesBackup) {
		return
	}

	formattedString := fmt.Sprintf("Traffic moved from '%v' to '%v' in cluster '%v'.", fromHost, toHost, cluster)
	fmt.Println(formattedString)
}

func Restore(cluster string, host string) {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)
	host = strings.ToLower(host)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return
	}

	hosts := NewHostConfig(fileLines[startLine:endLine])
	if !hosts.HostExists(host) {
		formattedString := fmt.Sprintf("Host '%v' does not exist in cluster '%v'.", host, cluster)
		fmt.Println(formattedString)
		return
	}

	restoreErr := hosts.RestoreTraffic(host)
	if restoreErr != nil {
		fmt.Println("There was an issue restoring traffic:", restoreErr)
		return
	}

	fileLinesBackup := make([]string, len(fileLines))
	_ = copy(fileLinesBackup, fileLines)
	fileLines = append(fileLines[:startLine], append(hosts.ToArray(), fileLines[endLine:]...)...)

	// write changes
	err := WriteHostsFileLines(fileLines)
	if err != nil {
		fmt.Println("There was an issue writing file:", err)
		return
	}

	// restart nginx, if error, restore hosts file.
	if !RestartNginx(fileLinesBackup) {
		return
	}

	formattedString := fmt.Sprintf("Traffic restored for '%v' in cluster '%v'.", host, cluster)
	fmt.Println(formattedString)
}

func GetClusterNodeCount(cluster string) int {
	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return 0
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int

	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			// add cluster to map
			return index - startLine
		}
	}

	return 0
}

func GetClusterNodes(cluster string) []string {
	// make cluster name lowercase
	cluster = strings.ToLower(cluster)

	// read hosts file
	fileLines, readErr := ReadHostsFileLines()
	if readErr != nil {
		fmt.Println("There was an issue opening/reading file:", readErr)
		return nil
	}

	pattern1 := "^### LB_K8S\\(" + cluster + "\\)"
	pattern2 := "^### LB_K8S_END"
	var startLine int
	var endLine int
	for index, line := range fileLines {
		// check if cluster exists
		if regexp.MustCompile(pattern1).MatchString(line) {
			startLine = index + 1
		}

		// grab cluster block end
		if startLine != 0 && regexp.MustCompile(pattern2).MatchString(line) {
			endLine = index
			break
		}
	}

	if startLine == 0 || startLine == endLine {
		// exit since cluster doesn't exist
		formattedString := fmt.Sprintf("Cluster '%v' does not exist.", cluster)
		fmt.Println(formattedString)
		return nil
	}

	hosts := NewHostConfig(fileLines[startLine:endLine])
	var hostList []string
	for k, _ := range hosts {
		hostList = append(hostList, []string{k}...)
	}

	return hostList
}
