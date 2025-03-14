package proxy

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func CreateTestFile(cluster string, site string) error {
	dirString := GetAvailableConfigDir(cluster)
	fileString := dirString + "/" + site + ".conf"

	if DirectoryExist(dirString) && !SiteExistsInCluster(cluster, site) {
		cmd := exec.Command("/usr/bin/touch", fileString)
		return cmd.Run()
	}

	return errors.New("can't create test file '" + fileString + "'")
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

// TODO: add test to test empty dir for cluster ""
var listTests = []listTest{
	listTest{"", "SiteEnabledsingle.localfalse"},
	listTest{"test1", "SiteEnabledtest.localtrue"},
	listTest{"test2", "cluster'test2'doesnotexist."},
	listTest{"empty", "Nositesavailableincluster'empty'"},
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

type enableTest struct {
	Cluster  string
	Hostname string
	Expected string
	Enabled  bool
}

var enableTests = []enableTest{
	enableTest{"", "single.local", "'single.local' enabled.", true},
	enableTest{"test1", "test.local", "Site 'test.local' is already enabled.", false},
	enableTest{"test1", "fail.local", "Site 'fail.local' does not exist in cluster 'test1'.", false},
	enableTest{"", "fail.local", "Site 'fail.local' does not exist.", false},
	enableTest{"test2", "fail.local", "Cluster 'test2' does not exist.", false},
}

func TestEnable(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())
	os.Setenv("PROXYMANAGER_DEV_MODE", "true") // set dev mode to prevent restart

	for _, test := range enableTests {
		// redirect STDOUT to buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run function
		Enable(test.Cluster, test.Hostname)

		// revert STDOUT
		w.Close()
		os.Stdout = oldStdout

		// collect function output to String
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// strip newline chars
		output = strings.ReplaceAll(output, "\n", "")

		// check output
		if output != test.Expected {
			t.Errorf("Expected '%v' for site '%v' in cluster '%v', received '%v'", test.Expected, test.Hostname, test.Cluster, output)
			continue
		}

		// check if site enabled.
		siteEnabled := SiteEnabled(test.Hostname)
		if !siteEnabled && test.Enabled {
			t.Errorf("Site '%v' in cluster '%v' was not enabled", test.Hostname, test.Cluster)
		}

		// disable site that was enabled by test
		if siteEnabled && test.Enabled {
			Disable(test.Cluster, test.Hostname)
		}
	}

	os.Clearenv()
}

type disableTest struct {
	Cluster  string
	Hostname string
	Expected string
	Disabled bool
}

var disableTests = []disableTest{
	disableTest{"test1", "test.local", "'test.local' disabled.", true},
	disableTest{"", "single.local", "Site 'single.local' is already disabled.", false},
	disableTest{"test1", "fail.local", "Site 'fail.local' does not exist in cluster 'test1'.", false},
	disableTest{"", "fail.local", "Site 'fail.local' does not exist.", false},
	disableTest{"test2", "fail.local", "Cluster 'test2' does not exist.", false},
}

func TestDisable(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())
	os.Setenv("PROXYMANAGER_DEV_MODE", "true") // set dev mode to prevent restart

	for _, test := range disableTests {
		// redirect STDout to buffer
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run function
		Disable(test.Cluster, test.Hostname)

		// revert stdout
		w.Close()
		os.Stdout = oldStdout

		// collect function output to string
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// strip newline chars
		output = strings.ReplaceAll(output, "\n", "")

		// check output
		if output != test.Expected {
			t.Errorf("Expected '%v' for site '%v' in cluster '%v', received '%v'", test.Expected, test.Hostname, test.Cluster, output)
			continue
		}

		// check if site disabled
		siteEnabled := SiteEnabled(test.Hostname)
		if siteEnabled && test.Disabled {
			t.Errorf("Site '%v' in cluster '%v' was not disabled", test.Hostname, test.Cluster)
		}

		// enable site that was disabled by test
		if !siteEnabled && test.Disabled {
			Enable(test.Cluster, test.Hostname)
		}
	}

	os.Clearenv()
}

type removeTest struct {
	Cluster  string
	Hostname string
	Create   bool
	Expected string
}

var removeTests = []removeTest{
	removeTest{"empty", "new.local", true, "'new.local' removed."},                                            // in-cluster
	removeTest{"", "new.local", true, "'new.local' removed."},                                                 // non-cluster
	removeTest{"test1", "test.local", false, "Site 'test.local' is enabled. Please disable before removing."}, // test enabled site
	removeTest{"", "new.local", false, "Site 'new.local' does not exist."},                                    // test site that doesn't exist
	removeTest{"empty", "new.local", false, "Site 'new.local' does not exist in cluster 'empty'."},            // test site that doesn't exist cluster
	removeTest{"test2", "new.local", false, "Cluster 'test2' does not exist."},                                // test cluster that doesn't exist
}

func TestRemove(t *testing.T) {
	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range removeTests {
		// check if a test file needs to be created.
		if test.Create {
			err := CreateTestFile(test.Cluster, test.Hostname)
			if err != nil {
				t.Errorf("%v", err)
			}
		}

		// redirect stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// run function
		Remove(test.Cluster, test.Hostname)

		//revert stdout
		w.Close()
		os.Stdout = oldStdout

		// collect output to string
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// strip newline chars
		output = strings.ReplaceAll(output, "\n", "")

		// check output
		if output != test.Expected {
			t.Errorf("Expected '%v' for site '%v' in cluster '%v', received '%v'", test.Expected, test.Hostname, test.Cluster, output)
		}

		// check if test file that was created was removed
		if test.Create && SiteExistsInCluster(test.Cluster, test.Hostname) {
			t.Errorf("Site '%v' in cluster '%v' was not removed as planned", test.Hostname, test.Cluster)
		}
	}

	os.Clearenv()
}

func TestSiteExists(t *testing.T) {
	
	tests := []struct {
		Hostname string
		Expected bool
	}{
		{"test.local", true},
		{"single.local", true},
		{"fail.local", false},
	}

	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())
	for _, test := range tests {
		if output := SiteExists(test.Hostname); output != test.Expected {
			t.Errorf("Expected '%v' for site '%v', received '%v'",test.Expected, test.Hostname, output)
		}
	}

	os.Clearenv()
}

func TestCreateSiteConfig(t *testing.T) {
	tests := []struct {
		FilePath string
		FileLines []string
		FileHash string
	}{
		{BuildFilePath("file1.txt"),[]string{"Hello World!"},"03ba204e50d126e4674c005e04d82e84c21366780af1f43bd54a37816b6ab340"},
		{BuildFilePath("file2.txt"),[]string{"Hello World!","Here's another line."},"85273a58194726014a61e216edc73c796a15799eb2cf4bf216d0cccc1ef789cf"},
	}

	for _, test := range tests {
		// create file
		if fileErr := CreateSiteConfig(test.FilePath, test.FileLines); fileErr != nil {
			t.Errorf("Issue creating file '%v': %v", test.FilePath, fileErr)
			continue
		}

		// check file hash
		if output, _ := GetFileHash(test.FilePath); output != test.FileHash {
			t.Errorf("Expected hash '%v' for file '%v', received '%v'", test.FileHash, test.FilePath, output)
		}

		// remove file
		os.Remove(test.FilePath)
	}
}

func TestGetMD5Hash(t *testing.T) {
	tests := []struct {
		Text string
		Hash string
	}{
		{"","d41d8cd98f00b204e9800998ecf8427e"},
		{"string","b45cffe084dd3d20d928bee85e7b0f21"},
		{"abcdefghijk@@@jlkjaads","74279c25a17e47f6fb22a9a2118dbb9b"},
	}

	for _, test := range tests {
		if output := GetMD5Hash(test.Text); output != test.Hash {
			t.Errorf("Expected hash '%v' for string '%v', received '%v'", test.Hash, test.Text, output)
		}
	}

}

func TestNew(t *testing.T) {
	tests := []struct {
		Cluster string
		Hostname string
		IPAddress string
		Port string
		URI string
		ssl bool
		sslBypassFirewall bool
		proxySsl bool
		proxySslVerifyOff bool
		Expected string
		CheckFileHash bool
		FileHash string
		Cleanup bool // used to tell test to cleanup file after run
	}{
		//{"","","","","",false, false, false, false, "", false, "", false},
		{"","single.local","10.0.0.1","1024","",false, false, false, false, "Site 'single.local' is already in use on this server.", false, "", false},
		{"","fail1.local","10.0.0.256","1024","",false, false, false, false, "IP Address not valid: 10.0.0.256", false, "", false},
		{"","fail2.local","10.0.0.1","1023","",false, false, false, false, "Port '1023' is invalid. please specify a port in the following range: 1024-49151", false, "", false},
		{"","fail3.local","10.0.0.1","49152","",false, false, false, false, "Port '49152' is invalid. please specify a port in the following range: 1024-49151", false, "", false},
		{"","fail4.local","10.0.0.1","abcd","",false, false, false, false, "Port 'abcd' is invalid. please specify a port in the following range: 1024-49151", false, "", false},
		{"","fail5.local","10.0.0.1","1024","/uri?a=b",false, false, false, false, "Uri '/uri?a=b' is invalid.A uri must start with a '/' and only contain the following characters: a-z, A-Z, 0-9, /, -, _, ., and ~", false, "", false},
		{"","fail6.local","10.0.0.1","1024","uri-test",false, false, false, false, "Uri 'uri-test' is invalid.A uri must start with a '/' and only contain the following characters: a-z, A-Z, 0-9, /, -, _, ., and ~", false, "", false},
		{"fail","fail7.local","10.0.0.1","1024","",false, false, false, false, "Cluster 'fail' does not exist.", false, "", false},
		{"empty","fail8.local","10.0.0.1","1024","",false, false, false, false, "Cluster 'empty' has no assigned nodes.", false, "", false},
		{"test1","fail9.local","10.0.0.1","","",false, false, false, false, "no port was specified.", false, "", false},
	}

	os.Setenv("PROXYMANAGER_CONFIG_PATH", GetConfigPath())

	for _, test := range tests {
		// check command ouput
		// redirect stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// run function
		New(test.Cluster,
			test.Hostname,
			test.IPAddress,
			test.Port,
			test.URI,
			test.ssl,
			test.sslBypassFirewall,
			test.proxySsl,
			test.proxySslVerifyOff)

		// revert stdout
		w.Close()
		os.Stdout = oldStdout

		// collect output to string
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// strip newline chars
		output = strings.ReplaceAll(output, "\n", "")

		// check output
		if output != test.Expected {
			t.Errorf("Site '%v': Expected '%v', received '%v'", test.Hostname, test.Expected, output)
		}

		// check file hash
		if test.CheckFileHash {
			filePath := GetAvailableConfigDir(test.Cluster) + "/" + test.Hostname + ".conf"

			if fileHash, _ := GetFileHash(filePath); test.FileHash != fileHash {
				t.Errorf("File '%v': Expected hash '%v', received '%v'", filePath, test.FileHash, fileHash)
			}
		}

		// cleanup any files created
		if test.Cleanup {
			filePath := GetAvailableConfigDir(test.Cluster) + "/" + test.Hostname + ".conf"
			os.Remove(filePath)
		}
		
	}

	os.Clearenv()
}
