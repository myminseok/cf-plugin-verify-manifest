package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/myminseok/verify-manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVerifyManifestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TestVerifyManifestPlugin Suite")
}

var _ = Describe("Parsing Arguments", func() {
	It("parses args", func() {
		manifestPath, err := ParseArgs(
			[]string{
				"verify-manifest",
				"-f", "manifest-path",
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(manifestPath).To(Equal("manifest-path"))
		Expect(PRINT_DEBUG).To((Equal(false)))
	})

	It("requires a manifest", func() {
		_, err := ParseArgs(
			[]string{
				"validate-manifest-ok",
			},
		)
		Expect(err).To(MatchError("Missing manifest argument"))
	})
	It("print debug mode", func() {
		manifestPath, err := ParseArgs(
			[]string{
				"validate-manifest-ok",
				"-f", "manifest-path",
				"-debug", "true",
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(manifestPath).To(Equal("manifest-path"))
		Expect(PRINT_DEBUG).To((Equal(true)))
	})
})

var _ = Describe("Parse manifest yaml", func() {

	manifest_good := LoadYAML("./fixtures/manifest-good.yml")
	manifest_bad := LoadYAML("./fixtures/manifest-bad.yml")

	It("parse good manifest for service", func() {
		manifestServices := ParseManifestServices(manifest_good)
		Expect(manifestServices).To(HaveLen(3))
	})

	It("parse bad manifest for service", func() {
		manifestServices := ParseManifestServices(manifest_bad)
		Expect(manifestServices).To(HaveLen(7))
	})

	It("parse good manifest for route", func() {
		manifestRoutes := ParseManifestRoutes(manifest_good)
		Expect(manifestRoutes).To(HaveLen(2))
	})

	It("parse bad manifest for route", func() {
		manifestRoutes := ParseManifestRoutes(manifest_bad)
		Expect(manifestRoutes).To(HaveLen(5))
	})
})

var _ = Describe("Fetch_cf_domains_guid", func() {
	var cliConnection *pluginfakes.FakeCliConnection
	var mapData []interface{}
	var stringSlice []string
	BeforeEach(func() {
		cliConnection = &pluginfakes.FakeCliConnection{}
		byteArray, _ := ioutil.ReadFile("./fixtures/domains-program-output-simple.json")
		err := json.Unmarshal(byteArray, &mapData)
		if err != nil {
			fmt.Println("[ERROR-test] ", err) //json: cannot unmarshal object into Go value of type []string
		}
		// fmt.Println(mapData)

		// Example: convert map values to string slice if all values are strings
		fmt.Println("convert map values")
		for a, v := range mapData {
			fmt.Printf("  A: %v, V:%s\n", a, v)
			v1b, _ := json.Marshal(v)
			fmt.Printf("  V1:%s\n", v1b)
			stringSlice = append(stringSlice, string(v1b))
		}

		// fmt.Println(stringSlice)
		cliConnection.CliCommandWithoutTerminalOutputStub = func(args ...string) ([]string, error) {
			return stringSlice, nil
		}
	})

	It("Fetch_cf_domains_guid", func() {
		domainsGuidMap, domainList := Fetch_cf_domains_guid(cliConnection)
		Expect(domainsGuidMap).To(HaveLen(3))
		Expect(domainList).To(HaveLen(3))
	})

})
