package loadbalancer

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
)

type Host struct {
	IpAddress       string
	Enabled         bool
	AdditionalHosts []string
}

type Hosts map[string]Host

func NewHostConfig(hostInfo []string) Hosts {
	var hosts Hosts
	hosts = make(Hosts)

	for _, str := range hostInfo {
		var host Host

		// check if host is disabled and remove '#' if so
		disabled_pattern := "^#.*"
		if regexp.MustCompile(disabled_pattern).MatchString(str) {
			host.Enabled = false
			str = strings.TrimSpace(strings.Replace(str, "#", "", 1))
		} else {
			host.Enabled = true
		}

		items := strings.Split(str, " ")
		hostname := items[1]
		host.IpAddress = items[0]

		// if additional items
		if len(items) > 2 {
			host.AdditionalHosts = append(host.AdditionalHosts, items[2:]...)
		}

		// append host
		hosts[hostname] = host
	}

	return hosts
}

func (h *Hosts) PrintHosts() {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
	fmt.Fprintln(w, "Host\tIP Address\tEnabled\tAdditional Hosts")
	for k, v := range *h {
		formattedString := fmt.Sprintf("%v\t%v\t%v\t%v", k, v.IpAddress, v.Enabled, strings.Join(v.AdditionalHosts, ","))
		fmt.Fprintln(w, formattedString)
	}
	w.Flush()
}

func (h *Hosts) HostExists(host string) bool {
	for k, _ := range *h {
		if k == host {
			return true
		}
	}

	return false
}

func (h *Hosts) IPExists(ipAddress string) bool {
	for _, v := range *h {
		if v.IpAddress == ipAddress {
			return true
		}
	}

	return false
}

func (h *Hosts) AddHost(host string, ipAddress string) error {
	var newHost Host
	newHost.Enabled = true
	newHost.IpAddress = ipAddress

	(*h)[host] = newHost
	return nil
}

func (h *Hosts) DelHost(host string) error {
	delete((*h), host)
	return nil
}

func (h *Hosts) MoveTraffic(fromHost string, toHost string) error {
	// check if fromHost == toHost
	if fromHost == toHost {
		return errors.New("loadbalancer(MoveTraffic): can't move traffic to self")
	}

	// check if fromHost already disabled.
	if !(*h)[fromHost].Enabled {
		return errors.New("loadbalancer(MoveTraffic): fromHost already disabled")
	}

	// check if from host is currently holding traffic
	if len((*h)[fromHost].AdditionalHosts) > 0 {
		return errors.New("loadbalancer(MoveTraffic): fromHost is already handling additional host's traffic")
	}

	// check if toHost is disabled.
	if !(*h)[toHost].Enabled {
		return errors.New("loadbalancer(MoveTraffic): toHost is disabled")
	}

	var from Host
	from.IpAddress = (*h)[fromHost].IpAddress
	from.Enabled = false

	var to Host
	to.IpAddress = (*h)[toHost].IpAddress
	to.Enabled = true
	to.AdditionalHosts = append((*h)[toHost].AdditionalHosts, []string{fromHost}...)

	delete((*h), fromHost)
	delete((*h), fromHost)

	(*h)[fromHost] = from
	(*h)[toHost] = to

	return nil
}

func (h *Hosts) RestoreTraffic(host string) error {
	if (*h)[host].Enabled {
		return errors.New("loadbalancer(MoveTraffic): Host is already enabled")
	}

	var restore Host
	restore.Enabled = true
	restore.IpAddress = (*h)[host].IpAddress

	var update Host
	var updateHost string
	// find where host traffic has been routed to
outerLoop:
	for k, v := range *h {
		for _, str := range v.AdditionalHosts {
			if str == host {
				updateHost = k
				break outerLoop
			}
		}
	}

	update.Enabled = true
	update.IpAddress = (*h)[updateHost].IpAddress
	for _, str := range (*h)[updateHost].AdditionalHosts {
		// skip if restore host
		if str == host {
			continue
		}

		update.AdditionalHosts = append(update.AdditionalHosts, []string{str}...)
	}

	delete((*h), host)
	delete((*h), updateHost)

	(*h)[host] = restore
	(*h)[updateHost] = update

	return nil
}

func (h *Hosts) ToArray() []string {
	var hosts []string
	for k, v := range *h {
		hostStr := v.IpAddress + " " + k + " " + strings.Join(v.AdditionalHosts, " ")
		if !v.Enabled {
			hostStr = "#" + hostStr
		}
		hosts = append(hosts, []string{strings.TrimSpace(hostStr)}...)
	}

	return hosts
}
