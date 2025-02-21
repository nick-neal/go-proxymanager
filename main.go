package main

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"strings"

	"nickneal.dev/go-proxymanager/loadbalancer"
	"nickneal.dev/go-proxymanager/proxy"
	"nickneal.dev/go-proxymanager/utils/settings"
)

type Command struct {
	name     string
	function string
	data     map[string]string
}

func parseArgs() Command {
	// collect arguments from command
	args := os.Args[1:]

	// insitalize command
	var command Command
	command.data = make(map[string]string)

	if len(args) >= 1 {
		switch args[0] {
		case "lb": // load balancer module
			command.name = args[0] // set command name
			// check if args are at least 3 in length to extract subcommand

			if len(args) >= 2 && args[1] == "list" {
				command.function = "list"
				return command
			}

			if len(args) >= 3 {
				if args[1] == "new" || args[1] == "remove" {
					command.function = args[1]
					command.data["cluster"] = args[2]
					return command
				}

				command.function = args[2] // set function name
				switch command.function {
				case "add", "move":
					// error out if all data not present
					if len(args) < 5 {
						fmt.Printf("parser: not enough args supplied for %v %v\n", command.name, command.function)
						printCommandHelp(command.name)
						os.Exit(1)
					}
					command.data["cluster"] = args[1]
					if command.function == "add" {
						command.data["host"] = args[3]
						command.data["ip"] = args[4]
					} else {
						command.data["from"] = args[3]
						command.data["to"] = args[4]
					}
				case "del", "restore":
					// error out if all data not present
					if len(args) < 4 {
						fmt.Printf("parser: not enough args supplied for %v %v\n", command.name, command.function)
						printCommandHelp(command.name)
						os.Exit(1)
					}
					command.data["cluster"] = args[1]
					command.data["host"] = args[3]
				case "status":
					command.data["cluster"] = args[1]
				default:
					printCommandHelp(command.name)
					os.Exit(0)
				}

				return command
			} else {
				printCommandHelp(command.name)
				os.Exit(0)
			}
		case "proxy": // proxy module
			command.name = args[0] // set command name
			if len(args) >= 2 {
				switch args[1] {
				case "list":
					command.function = args[1]
					// check for optional args
					if len(args) > 2 {
						loopArgs := args[2:]
						for index, str := range loopArgs {
							switch str {
							case "--k8s":
								key := strings.Replace(str, "--", "", 1)
								// make sure lookahead isn't out of array bounds
								var value string
								if (index + 1) < len(loopArgs) {
									value = loopArgs[index+1]
								} else {
									continue
								}
								// make sure next value isn't another param
								if !regexp.MustCompile("^--.*").MatchString(value) {
									command.data[key] = value
								}
							}
						}
					}
				case "enable", "disable", "remove":
					command.function = args[1]
					if len(args) < 3 {
						fmt.Printf("parser: not enough args supplied for %v %v\n", command.name, command.function)
						printCommandHelp(command.name)
						os.Exit(1)
					}

					command.data["hostname"] = args[2]

					// check for optional args
					if len(args) > 3 {
						loopArgs := args[3:]
						for index, str := range loopArgs {
							switch str {
							case "--k8s":
								key := strings.Replace(str, "--", "", 1)
								// make sure lookahead isn't out of array bounds
								var value string
								if (index + 1) < len(loopArgs) {
									value = loopArgs[index+1]
								} else {
									continue
								}
								// make sure next value isn't another param
								if !regexp.MustCompile("^--.*").MatchString(value) {
									command.data[key] = value
								}
							}
						}
					}
				case "new":
					command.function = args[1]
					if len(args) < 4 {
						fmt.Printf("parser: not enough args supplied for %v %v\n", command.name, command.function)
						printCommandHelp(command.name)
						os.Exit(1)
					}

					command.data["hostname"] = args[2]

					loopArgs := args[3:]
					for index, str := range loopArgs {
						switch str {
						case "--ip", "--port", "--k8s", "--proxy-uri":
							key := strings.Replace(str, "--", "", 1)
							// make sure lookahead isn't out of array bounds
							var value string
							if (index + 1) < len(loopArgs) {
								value = loopArgs[index+1]
							} else {
								continue
							}
							// make sure next value isn't another param
							if !regexp.MustCompile("^--.*").MatchString(value) {
								command.data[key] = value
							}
						case "--ssl", "--ssl-bypass-firewall", "--proxy-ssl", "--proxy-ssl-verify-off":
							key := strings.Replace(str, "--", "", 1)
							command.data[key] = "true"
						}
					}

					if (command.data["ip"] == "" && command.data["k8s"] == "") || (command.data["ip"] != "" && command.data["k8s"] != "") {
						fmt.Println("parser: must either specify '--ip' or '--k8s'")
						printCommandHelp(command.name)
						os.Exit(1)
					}

					// require port if k8s
					if command.data["k8s"] != "" && command.data["port"] == "" {
						fmt.Println("parser: must specify '--port' if '--k8s' is specified")
						printCommandHelp(command.name)
						os.Exit(1)
					}

				default:
					printCommandHelp(command.name)
					os.Exit(0)
				}

				return command
			} else {
				printCommandHelp(command.name)
				os.Exit(0)
			}

			printCommandHelp("proxy")
		case "fw":
			printCommandHelp("fw")
		default:
			printHelp()
		}

	} else {
		fmt.Println("parser: no arguments supplied.")
		printHelp()
		os.Exit(1)
	}

	return command
}

func isRootUser() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Username == "root"
}

func printHelp() {
	fmt.Println("This is the main help menu.")
}

func printCommandHelp(commandName string) {
	fmt.Printf("This is a help menu for %v.\n", commandName)
}

func main() {
	// check if root user except if devmode is enabled
	if !isRootUser() && !settings.CheckDevMode() {
		formattedString := fmt.Sprintf("error: '%v' must be run as root user.", os.Args[0])
		fmt.Println(formattedString)
		os.Exit(1)
	}

	// get args
	command := parseArgs()

	// route command
	switch command.name {
	case "lb":
		switch command.function {
		case "list":
			loadbalancer.List()
		case "new":
			loadbalancer.New(command.data["cluster"])
		case "remove":
			loadbalancer.Remove(command.data["cluster"])
		case "status":
			loadbalancer.Status(command.data["cluster"])
		case "add":
			loadbalancer.Add(command.data["cluster"], command.data["ip"], command.data["host"])
		case "del":
			loadbalancer.Del(command.data["cluster"], command.data["host"])
		case "move":
			loadbalancer.Move(command.data["cluster"], command.data["from"], command.data["to"])
		case "restore":
			loadbalancer.Restore(command.data["cluster"], command.data["host"])
		}
	case "proxy":
		switch command.function {
		case "list":
			proxy.List(command.data["k8s"])
		case "enable":
			proxy.Enable(command.data["k8s"], command.data["hostname"])
		case "disable":
			proxy.Disable(command.data["k8s"], command.data["hostname"])
		case "remove":
			proxy.Remove(command.data["k8s"], command.data["hostname"])
		case "new":
			var ssl bool
			var sslBypassFirewall bool
			var proxySsl bool
			var proxySslVerifyOff bool

			if command.data["ssl"] == "true" {
				ssl = true
			} else {
				ssl = false
			}

			if command.data["ssl-bypass-firewall"] == "true" {
				sslBypassFirewall = true
			} else {
				sslBypassFirewall = false
			}

			if command.data["proxy-ssl"] == "true" {
				proxySsl = true
			} else {
				proxySsl = false
			}

			if command.data["proxy-ssl-verify-off"] == "true" {
				proxySslVerifyOff = true
			} else {
				proxySslVerifyOff = false
			}

			proxy.New(command.data["k8s"], command.data["hostname"], command.data["ip"], command.data["port"], command.data["proxy-uri"], ssl, sslBypassFirewall, proxySsl, proxySslVerifyOff)
		}

	}

}
