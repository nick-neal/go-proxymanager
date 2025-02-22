package validate

import "testing"

// test ValidateClusterName(cluster)
type clusterNameTest struct {
	Name     string
	Expected bool
}

var clusterNameTests = []clusterNameTest{
	clusterNameTest{"hello", true},      // good
	clusterNameTest{"hell-", false},     // has illegal char -
	clusterNameTest{"asdf1234", true},   // good
	clusterNameTest{"L337C0D3", false},  // uses uppercase
	clusterNameTest{";rm -rf /", false}, // uses space and illegal chars
}

func TestValidateClusterName(t *testing.T) {
	for _, test := range clusterNameTests {
		if output := ValidateClusterName(test.Name); output != test.Expected {
			t.Errorf("'%v' returned '%v' when it should have returned '%v'", test.Name, output, test.Expected)
		}
	}
}

// test ValidateHostName(host)
type hostNameTest struct {
	HostName string
	Expected bool
}

var hostNameTests = []hostNameTest{
	hostNameTest{"example.com", true},
	hostNameTest{";rm -rf /", false},
}

func TestValidateHostName(t *testing.T) {
	for _, test := range hostNameTests {
		if output := ValidateHostName(test.HostName); output != test.Expected {
			t.Errorf("'%v' returned '%v' when it should have returned '%v'", test.HostName, output, test.Expected)
		}
	}
}

// test ValidateIPAddress(ipAddress)
type ipAddressTest struct {
	ipAddress string
	Expected  bool
}

var ipAddressTests = []ipAddressTest{
	ipAddressTest{"192.168.0.1", true},
	ipAddressTest{"192.168.0.0", false},
	ipAddressTest{"256.256.256.256", false},
	ipAddressTest{"hello", false},
	ipAddressTest{";rm -rf /", false},
}

func TestValidateIPAddress(t *testing.T) {
	for _, test := range ipAddressTests {
		if output := ValidateIPAddress(test.ipAddress); output != test.Expected {
			t.Errorf("'%v' returned '%v' when it should have returned '%v'", test.ipAddress, output, test.Expected)
		}
	}
}

// test ValidatePort(port)
type portTest struct {
	Port     string
	Expected bool
}

var portTests = []portTest{
	portTest{"80", true},
	portTest{"443", true},
	portTest{"1024", true},
	portTest{"49151", true},
	portTest{"1023", false},
	portTest{"49152", false},
	portTest{"hello", false},
	portTest{";rm -rf /", false},
}

func TestValidatePort(t *testing.T) {
	for _, test := range portTests {
		if output := ValidatePort(test.Port); output != test.Expected {
			t.Errorf("'%v' returned '%v' when it should have returned '%v'", test.Port, output, test.Expected)
		}
	}
}

// test ValidateUri(uri)
type uriTest struct {
	uri      string
	Expected bool
}

var uriTests = []uriTest{
	uriTest{"/backups", true},
	uriTest{"/backup/s", true},
	uriTest{"/BACKUPS/s/09/", true},
	uriTest{"/", true},
	uriTest{"/.back~-_", true},
	uriTest{"/hello?a=b", false},
	uriTest{"hello", false},
	uriTest{";rm -rf /", false},
}

func TestValidateUri(t *testing.T) {
	for _, test := range uriTests {
		if output := ValidateUri(test.uri); output != test.Expected {
			t.Errorf("'%v' returned '%v' when it should have returned '%v'", test.uri, output, test.Expected)
		}
	}
}
