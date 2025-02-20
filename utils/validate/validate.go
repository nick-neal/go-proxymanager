package validate

import (
	"regexp"
	"strconv"
)

func ValidateClusterName(cluster string) bool {
	pattern := "^[a-z0-9]*$" //only lowercase and numbers
	if regexp.MustCompile(pattern).MatchString(cluster) {
		return true
	}

	return false
}

func ValidateHostName(host string) bool {
	pattern := "^[a-z0-9]+([-\\.][a-z0-9]+)*$" //only lowercase and numbers
	if regexp.MustCompile(pattern).MatchString(host) {
		return true
	}

	return false

}

func ValidateIPAddress(ipAddress string) bool {
	pattern := "^(([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])$" //only lowercase and numbers
	if regexp.MustCompile(pattern).MatchString(ipAddress) {
		return true
	}

	return false
}

func ValidatePort(port string) bool {
	// can 1024-49151
	pattern := "^.[0-9]*$"
	if regexp.MustCompile(pattern).MatchString(port) && len(port) <= 5 {
		num, _ := strconv.ParseInt(port, 10, 64)
		if num >= 1024 && num <= 49151 {
			return true
		}
	}

	return false
}

func ValidateUri(uri string) bool {
	pattern := "^\\/[a-zA-Z0-9\\/\\-_.~]*$"
	if regexp.MustCompile(pattern).MatchString(uri) {
		return true
	}

	return false
}
