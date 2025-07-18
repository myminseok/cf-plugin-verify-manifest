package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
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
	manifest_invalid_path := "./fixtures/manifest-invalid.yml"
	manifest_good_path := "./fixtures/manifest-good.yml"
	manifest_bad_path := "./fixtures/manifest-bad.yml"
	manifest_invalid := LoadYAML(manifest_invalid_path)
	manifest_good := LoadYAML(manifest_good_path)
	manifest_bad := LoadYAML(manifest_bad_path)

	It("parse invalid manifest for service", func() {
		manifestServices := ParseManifestServices(manifest_invalid_path, manifest_invalid)
		Expect(manifestServices).To(HaveLen(0))
	})

	It("parse good manifest for service", func() {
		manifestServices := ParseManifestServices(manifest_good_path, manifest_good)
		Expect(manifestServices).To(HaveLen(3))
	})

	It("parse bad manifest for service", func() {
		manifestServices := ParseManifestServices(manifest_bad_path, manifest_bad)
		Expect(manifestServices).To(HaveLen(7))
	})

	It("parse invalid manifest for route", func() {
		manifestRoutes := ParseManifestRoutes(manifest_invalid_path, manifest_invalid)
		Expect(manifestRoutes).To(HaveLen(0))
	})
	It("parse good manifest for route", func() {
		manifestRoutes := ParseManifestRoutes(manifest_good_path, manifest_good)
		Expect(manifestRoutes).To(HaveLen(2))
	})

	It("parse bad manifest for route", func() {
		manifestRoutes := ParseManifestRoutes(manifest_bad_path, manifest_bad)
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

var _ = Describe("Check_services", func() {
	manifest_good_path := "./fixtures/manifest-good.yml"
	manifest_good := LoadYAML(manifest_good_path)
	manifest_bad_path := "./fixtures/manifest-bad.yml"
	manifest_bad := LoadYAML(manifest_bad_path)
	var cliConnection *pluginfakes.FakeCliConnection
	var faceServices []plugin_models.GetServices_Model

	BeforeEach(func() {
		cliConnection = &pluginfakes.FakeCliConnection{}
		fakeServices1 := plugin_models.GetServices_Model{
			Name: "existing-service-1",
		}
		fakeServices2 := plugin_models.GetServices_Model{
			Name: "existing-service-2",
		}

		faceServices = []plugin_models.GetServices_Model{fakeServices1, fakeServices2}
		cliConnection.GetServicesStub = func() ([]plugin_models.GetServices_Model, error) {
			return faceServices, nil
		}

	})

	It("Check_services manifest_good", func() {
		status := Check_services(cliConnection, manifest_good_path, manifest_good)
		statusStr := fmt.Sprintf("%v", status) // TODO: no method for bool value
		Expect(statusStr).To(Equal("true"))
	})
	It("Check_services manifest_bad", func() {
		status := Check_services(cliConnection, manifest_bad_path, manifest_bad)
		statusStr := fmt.Sprintf("%v", status) // TODO: no method for bool value
		Expect(statusStr).To(Equal("false"))
	})

})
