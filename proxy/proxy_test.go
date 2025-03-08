package proxy

import (
	"bytes"
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
	getAvailableSitesTest{"fail", nil},
	getAvailableSitesTest{"test1", []string{"test.local"}},
	getAvailableSitesTest{"", []string{"single.local"}},
}

func TestGetAvailableSites(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range getAvailableSitesTests {
		output, _ := GetAvailableSites(test.Cluster)

		// prevent range panic
		if output == nil || test.Expected == nil {
			if output == nil && test.Expected != nil {
				t.Errorf("Expected %d items, received nil", len(test.Expected))
			}

			if output != nil && test.Expected == nil {
				t.Errorf("Expected nil items, received %d items", len(output))
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

type clusterExistsTest struct {
	Cluster  string
	Expected bool
}

var clusterExistsTests = []clusterExistsTest{
	clusterExistsTest{"test1", true},
	clusterExistsTest{"404", false},
	clusterExistsTest{"", true}, // empty test
}

func TestClusterExists(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range clusterExistsTests {
		if output := ClusterExists(test.Cluster); output != test.Expected {
			t.Errorf("Expected '%v' for cluster '%v', but received '%v'", test.Expected, test.Cluster, output)
		}
	}

	os.Clearenv()
}

type siteExistsInClusterTest struct {
	Cluster  string
	Hostname string
	Expected bool
}

var siteExistsInClusterTests = []siteExistsInClusterTest{
	siteExistsInClusterTest{"test1", "test.local", true},
	siteExistsInClusterTest{"test1", "test.com", false},
	siteExistsInClusterTest{"failure", "test.local", false},
	siteExistsInClusterTest{"", "single.local", true},
	siteExistsInClusterTest{"", "", false}, // empty test
}

func TestSiteExistsInCluster(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range siteExistsInClusterTests {
		if output := SiteExistsInCluster(test.Cluster, test.Hostname); output != test.Expected {
			t.Errorf("Expected '%v' for site '%v' in cluster '%v', but received '%v'", test.Expected, test.Hostname, test.Cluster, output)
		}
	}

	os.Clearenv()
}

type siteEnabledTest struct {
	Hostname string
	Expected bool
}

var siteEnabledTests = []siteEnabledTest{
	siteEnabledTest{"test.local", true},
	siteEnabledTest{"fail.net", false},
	siteEnabledTest{"", false}, // empty test
}

func TestSiteEnabled(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range siteEnabledTests {
		if output := SiteEnabled(test.Hostname); output != test.Expected {
			t.Errorf("Expected '%v' for site '%v'. but received '%v'", test.Expected, test.Hostname, output)
		}
	}

	os.Clearenv()
}

type listTest struct {
	Cluster  string
	Expected string
}

var listTests = []listTest{
	listTest{"", "SiteEnabledsingle.localfalse"},
	listTest{"test1", "SiteEnabledtest.localtrue"},
	listTest{"test2", "cluster'test2'doesnotexist."},
	listTest{"empty", "Nositesavailable"},
}

func TestList(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range listTests {
		// redirect STDOUT to buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run function
		List(test.Cluster)

		// revert STDOUT
		w.Close()
		os.Stdout = oldStdout

		// collect function output to String
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// strip new lines and spaces (used since table formatting is unpredictable)
		output = strings.ReplaceAll(output, " ", "")
		output = strings.ReplaceAll(output, "\n", "")

		if output != test.Expected {
			t.Errorf("Expected '%v', received '%v'", test.Expected, output)
		}

	}

	os.Clearenv()
}

func TestEnable(t *testing.T) {}

func TestDisable(t *testing.T) {}

func TestRemove(t *testing.T) {}

func TestSiteExists(t *testing.T) {}

func TestCreateSiteConfig(t *testing.T) {}

func TestGetMD5Hash(t *testing.T) {}

func TestNew(t *testing.T) {}
