package proxy

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

func Getwd() string {
	cwd, _ := os.Getwd()
	return cwd
}

func GetConfigPath() string {
	if runtime.GOOS == "windows" {
		return Getwd() + "\\..\\test_configs\\proxymanager.yml"
	} else {
		return Getwd() + "/../test_configs/proxymanager.yml"
	}
}

func BuildFilePath(filePath string) string {
	if runtime.GOOS == "windows" {
		filePath = strings.ReplaceAll(filePath, "/", "\\")
		return Getwd() + "\\..\\" + filePath
	} else {
		return Getwd() + "/../" + filePath
	}
}

func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	return fmt.Sprintf("%x", hashBytes), nil

}

func TestGetNginxDir(t *testing.T) {

	// set test config path
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	get := GetNginxDir()
	want := "../test_configs/nginx"

	if get != want {
		t.Errorf("Expected '%v', got '%v'", want, get)
	}

	os.Clearenv()
}

type directoryExistTest struct {
	Path     string
	Expected bool
}

var directoryExistTests = []directoryExistTest{
	directoryExistTest{BuildFilePath("test_configs/nginx/sites-available/k8s_test1"), true},
	directoryExistTest{"test_configs/nginx/k8s_test0", false},
}

func TestDirectoryExist(t *testing.T) {
	for _, test := range directoryExistTests {
		if output := DirectoryExist(test.Path); output != test.Expected {
			t.Errorf("Expected '%v' for '%v' but received '%v'", test.Expected, test.Path, output)
		}
	}

}

func TestGetAvailableConfigDir(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	// test cluster
	clusterGet := GetAvailableConfigDir("cluster")
	clusterWant := GetNginxDir() + "/sites-available/k8s_cluster"

	if clusterGet != clusterWant {
		t.Errorf("Expected '%v' but received '%v'", clusterWant, clusterGet)
	}

	// test non-cluster
	get := GetAvailableConfigDir("")
	want := GetNginxDir() + "/sites-available"

	if get != want {
		t.Errorf("Expected '%v' but received '%v'", want, get)
	}

	os.Clearenv()
}

func TestGetEnabledConfigDir(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	get := GetEnabledConfigDir()
	want := GetNginxDir() + "/sites-enabled"

	if get != want {
		t.Errorf("Expected '%v' but received '%v'", want, get)
	}

	os.Clearenv()
}

func TestGetEnabledSites(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	get, err := GetEnabledSites()
	if err != nil {
		t.Errorf("Error occured %v", err)
		return
	}
	want := []string{"test.local"}

	for i, _ := range get {
		if get[i] != want[i] {
			t.Errorf("Expected '%v' at index %d but received '%v'", want[i], i, get[i])
		}
	}

	os.Clearenv()
}

type getAvailableSitesTest struct {
	Cluster  string
	Expected []string
}

var getAvailableSitesTests = []getAvailableSitesTest{
	getAvailableSitesTest{"", nil},
	getAvailableSitesTest{"test1", []string{"test.local"}},
}

func TestGetAvailableSites(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range getAvailableSitesTests {
		output, err := GetAvailableSites(test.Cluster)

		if err != nil {
			t.Errorf("Error occured %v", err)
			continue
		}

		if output == nil {
			if test.Expected != nil {
				t.Errorf("Expected %d items, received nil", len(test.Expected))
			}
			continue
		}

		for i, _ := range output {
			if output[i] != test.Expected[i] {
				t.Errorf("Expected '%v' at index %d but received '%v'", test.Expected[i], i, output[i])
			}
		}
	}

	os.Clearenv()
}

func TestClusterExists(t *testing.T) {}

func TestSiteExistsInCluster(t *testing.T) {}

func TestSiteEnabled(t *testing.T) {}

func TestList(t *testing.T) {}

func TestEnable(t *testing.T) {}

func TestDisable(t *testing.T) {}

func TestRemove(t *testing.T) {}

func TestSiteExists(t *testing.T) {}

func TestCreateSiteConfig(t *testing.T) {}

func TestGetMD5Hash(t *testing.T) {}

func TestNew(t *testing.T) {}
