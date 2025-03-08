package settings

import (
	"os"
	"runtime"
	"testing"

	"github.com/mitchellh/hashstructure/v2"
)

func Getwd() string {
	cwd, _ := os.Getwd()
	return cwd
}

func TestCheckDevMode(t *testing.T) {
	// check true
	os.Setenv("PROXYMANAGER_DEV_MODE", "true")

	get1 := CheckDevMode()
	want1 := true

	if get1 != want1 {
		t.Errorf("Environment variable PROXYMANAGER_DEV_MODE was set to true, but function returned %v", get1)
	}

	// check false
	os.Clearenv()
	get2 := CheckDevMode()
	want2 := false

	if get2 != want2 {
		t.Errorf("Environment variable PROXYMANAGER_DEV_MODE was unset, but function returned %v", get2)
	}

}

type getConfigPathTest struct {
	EnvironmentVariables map[string]string
	Expected             string
}

var getConfigPathTests = []getConfigPathTest{
	getConfigPathTest{map[string]string{"PROXYMANAGER_DEV_MODE": "true"}, Getwd() + "/etc/proxymanager.yml"},
	getConfigPathTest{map[string]string{"PROXYMANAGER_DEV_MODE": "true", "PROXYMANAGER_CONFIG_PATH": "/test/proxymanager.yml"}, "/test/proxymanager.yml"},
	getConfigPathTest{map[string]string{"PROXYMANAGER_CONFIG_PATH": "/test/proxymanager.yml"}, "/test/proxymanager.yml"},
	getConfigPathTest{map[string]string{}, "/etc/proxymanager.yml"},
}

func TestGetConfigPath(t *testing.T) {
	for _, test := range getConfigPathTests {
		for k, v := range test.EnvironmentVariables {
			os.Setenv(k, v)
		}

		if output := GetConfigPath(); output != test.Expected {
			t.Errorf("Expected '%v' but received '%v'", test.Expected, output)
		}

		os.Clearenv()
	}
}

func TestLoadConfig(t *testing.T) {

	var config_path string
	if runtime.GOOS == "windows" {
		config_path = "\\..\\..\\test_configs\\proxymanager.yml"
	} else {
		config_path = "/../../test_configs/proxymanager.yml"
	}

	os.Setenv("PROXYMANAGER_CONFIG_PATH", Getwd()+config_path)
	get, _ := hashstructure.Hash(LoadConfig(), hashstructure.FormatV2, nil)
	var want = uint64(8382684015897467055) //test config hash (test_configs/proxymanager.yml)

	if get != want {
		t.Errorf("test proxymanager.yml config returned hash '%d' instead of '%d'", get, want)
	}
	os.Clearenv()
}

func TestDefaultConfig(t *testing.T) {

	var config_path = "file_not_exist.txt"

	os.Setenv("PROXYMANAGER_CONFIG_PATH", Getwd()+config_path)
	get, _ := hashstructure.Hash(LoadConfig(), hashstructure.FormatV2, nil)
	var want = uint64(3297509149356143636) //default config hash

	if get != want {
		t.Errorf("test proxymanager.yml config returned hash '%d' instead of '%d'", get, want)
	}
	os.Clearenv()
}
