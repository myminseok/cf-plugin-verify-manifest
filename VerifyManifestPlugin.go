package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	"gopkg.in/yaml.v2"
)

// CreateServicePush is the struct implementing the interface defined by the core CLI. It can
// be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type VerifyManifestPlugin struct {
	// Exit ExitInterface
}

type YManifest struct {
	Applications []YApplication `yaml:"applications"`
}

type YApplication struct {
	Name     string                 `yaml:"name"`
	Env      map[string]interface{} `yaml:"env"`
	Services []string               `yaml:"services"`
	Routes   []YRoute               `yaml:"routes"`
}

type YRoute struct {
	Route    string `yaml:"route"`
	Protocol string `yaml:"protocol"`
}

type ManifestService struct {
	appName string
	service string
}
type ManifestRoute struct {
	appName string
	route   string
}

type AppServiceResult struct {
	appName string
	service string
}

type AppRouteResult struct {
	appName string
	route   string
	message string
}

var PRINT_DEBUG bool

func main() {
	plugin.Start(new(VerifyManifestPlugin))
}

func (c *VerifyManifestPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "VerifyManifest",
		Version: plugin.VersionType{Major: 1, Minor: 0, Build: 0},
		Commands: []plugin.Command{
			{
				Name:     "verify-manifest",
				HelpText: "verify manifest.yml for service instances and routes",
				UsageDetails: plugin.Usage{
					Usage: "cf verify-manifest PATH_TO_MANIFEST [-debug true]",
				},
			},
		},
	}
}

func (c *VerifyManifestPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "verify-manifest" {
		manifestPath := ParseArgs(args)
		manifest := loadYAML(manifestPath)
		print_cf_target(cliConnection, args)
		var all_good = true
		all_good = check_manifest_services(cliConnection, manifestPath, manifest)
		all_good = check_routes(cliConnection, manifestPath, manifest)
		if !all_good {
			os.Exit(1)
		}
	}
}

func print_usage() {
	fmt.Println("Usage: cf verify-manifest -f PATH_TO_MANIFEST [-debug true]")
}

func ParseArgs(args []string) string {

	flags := flag.NewFlagSet("verify-manifest", flag.ContinueOnError)
	manifestPath := flags.String("f", "", "path to an application manifest")
	debugModeFlag := flags.String("debug", "", "verbose plugin")
	err := flags.Parse(args[1:])
	if err != nil {
		fmt.Printf("[ERROR] %s\n", err)
		print_usage()
		os.Exit(1)
	}
	if manifestPath == nil || *manifestPath == "" || len(*manifestPath) == 0 {
		fmt.Println("[ERROR] Missing manifest argument")
		print_usage()
		os.Exit(1)
	}
	if len(*debugModeFlag) > 0 {
		fmt.Println("[INFO] found '-debug' flag, enabling print debug mode ")
		PRINT_DEBUG = true
	}
	fmt.Printf("Using manifestPath: '%s'\n", *manifestPath)
	return *manifestPath
}

func print_debug(arg string) {
	if PRINT_DEBUG {
		fmt.Printf("    [DEBUG]%s\n", arg)
	}
}

func loadYAML(manifestPath string) (manifest YManifest) {
	b, err := ioutil.ReadFile(manifestPath)

	if err != nil {
		fmt.Errorf("[ERROR] Unable to read manifest file: %s", manifestPath)
		os.Exit(1)
	}

	var document YManifest
	err = yaml.Unmarshal(b, &document)

	if err != nil {
		fmt.Errorf("[ERROR] Unable to parse manifest file: %s", manifestPath)
		os.Exit(1)
	}
	return document
}

func ParseManifestServices(manifest YManifest) (manifestServices []ManifestService) {
	fmt.Println("  Parsing manifest services ...")
	for _, app := range manifest.Applications {
		if app.Services != nil {
			for _, service := range app.Services {
				manifestServices = append(manifestServices, ManifestService{appName: app.Name, service: service})
			}
		}
	}
	for _, manifestService := range manifestServices {
		print_debug(fmt.Sprintf("  - app: '%s', service: '%s'", manifestService.appName, manifestService.service))
	}
	return manifestServices
}

func ParseManifestRoutes(manifest YManifest) (manifestRoutes []ManifestRoute) {
	fmt.Println("  Parsing manifest routes ...")
	for _, app := range manifest.Applications {
		if app.Routes != nil {
			for _, route := range app.Routes {
				manifestRoutes = append(manifestRoutes, ManifestRoute{appName: app.Name, route: route.Route})
			}
		}
	}

	for _, manifestRoute := range manifestRoutes {
		print_debug(fmt.Sprintf("  - app: '%s', route: '%s'", manifestRoute.appName, manifestRoute.route))
	}

	return manifestRoutes
}

func print_cf_target(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("cf target ...")
	output, err := cliConnection.CliCommandWithoutTerminalOutput(append([]string{"target"})...)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		os.Exit(1)
	}
	for _, item := range output {
		fmt.Println("  ", item)
	}
}

func fetch_cf_services(cliConnection plugin.CliConnection) (cf_services []string) {
	fmt.Println("  Fetching cf services from the target foundation ...")
	cfServices, err := cliConnection.GetServices()
	if err != nil {
		fmt.Errorf("[ERROR] ", err)
		os.Exit(1)
	}
	for _, cfService := range cfServices {
		cf_services = append(cf_services, cfService.Name)
	}
	fmt.Println("  [INFO] Existing service instances in current space:")
	for _, cf_service := range cf_services {
		fmt.Printf("  - '%s'\n", cf_service)
	}
	return cf_services
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {

		if a == b {
			// print_debug(fmt.Sprintf(" stringInSlice: %s %s %s", a, b, "true"))
			return true
		}
	}
	// print_debug(fmt.Sprintf(" stringInSlice: %s %s ", a, "false"))
	return false
}

func check_routes(cliConnection plugin.CliConnection, manifestPath string, manifest YManifest) (status bool) {
	status = true
	fmt.Println("\nChecking Routes availability specified in the manifest from the target ... ", manifestPath)
	manifestRoutes := ParseManifestRoutes(manifest)
	domainsGuidMap, domainList := fetch_cf_domains_guid(cliConnection)
	var goodList []AppRouteResult
	var badList []AppRouteResult
	for _, manifestRoute := range manifestRoutes {
		app := manifestRoute.appName
		route := manifestRoute.route
		host, domain := split_route(route)
		if len(domain) == 0 {
			badList = append(badList, AppRouteResult{appName: app, route: route, message: "Invalid route. too short"})
		} else if !stringInSlice(domain, domainList) {
			badList = append(badList, AppRouteResult{appName: app, route: route, message: fmt.Sprintf("No such domain '%s' in cf domains", domain)})
		} else {
			if check_route_reserved(cliConnection, host, domain, domainsGuidMap[domain]) {
				badList = append(badList, AppRouteResult{appName: app, route: route, message: fmt.Sprintf("Reserved is reserved")})
			} else {
				goodList = append(goodList, AppRouteResult{appName: app, route: route, message: ""})
			}
		}
	}
	if len(goodList) > 0 {
		fmt.Println("  [GOOD] Available routes specified in the manifest in the target foundation:")
		for _, item := range goodList {
			fmt.Printf("  - app: '%s', service: '%s'\n", item.appName, item.route)
		}
	}
	if len(badList) > 0 {
		status = false
		fmt.Println("  [ERROR] Not Available routes specified in the manifest in the target foundation:")
		for _, item := range badList {
			fmt.Printf("  - app: '%s', service: '%s' -> %s\n", item.appName, item.route, item.message)
		}
	}
	return status
}

func check_manifest_services(cliConnection plugin.CliConnection, manifestPath string, manifest YManifest) (status bool) {
	status = true
	fmt.Println("\nChecking Service instance from the manifest ...", manifestPath)
	manifestServices := ParseManifestServices(manifest)
	cf_services := fetch_cf_services(cliConnection)

	var goodList []AppServiceResult
	var badList []AppServiceResult
	for _, manifestService := range manifestServices {
		if stringInSlice(manifestService.service, cf_services) {
			goodList = append(goodList, AppServiceResult{appName: manifestService.appName, service: manifestService.service})
		} else {
			badList = append(badList, AppServiceResult{appName: manifestService.appName, service: manifestService.service})
		}
	}

	if len(goodList) > 0 {
		fmt.Println("  [GOOD] Service instance specified in manifest exists in current space:")
		for _, item := range goodList {
			fmt.Printf("  - app: '%s', service: '%s'\n", item.appName, item.service)
		}
	}

	if len(badList) > 0 {
		status = false
		fmt.Println("  [ERROR] Missing Service instance specified in manifest exists in current space:")
		for _, item := range badList {
			fmt.Printf("  - app: '%s', service: '%s'\n", item.appName, item.service)
		}
	}
	return status
}

func split_route(manifestRoute string) (host string, domain string) {
	manifestRouteSplit := strings.SplitN(manifestRoute, ".", 2)

	host = manifestRouteSplit[0]

	if len(manifestRouteSplit) > 1 {
		domain = manifestRouteSplit[1]
	}
	// print_debug(fmt.Sprintf("  split_route: %s ->  host:'%s', domain:'%s'", manifestRoute, host, domain))
	return host, domain
}

// fetch domain list/guid
func fetch_cf_domains_guid(cliConnection plugin.CliConnection) (domainsGuidMap map[string]string, domainList []string) {
	fmt.Println("  Fetching cf domains, guid from the target foundation ...")

	// url := fmt.Sprintf("curl /v3/domains | jq \'.resources[]| \"\\(.name) \\(.guid)\"\\")
	output, err := cliConnection.CliCommandWithoutTerminalOutput(append([]string{"curl", "/v3/domains"})...)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		os.Exit(1)
	}

	jsonStr := strings.Join(output, "") // Combine output lines
	var jsonObj Domains
	err = json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		os.Exit(1)
	}

	domainsGuidMap = make(map[string]string)
	for _, resource := range jsonObj.Resources {
		print_debug(fmt.Sprintf("  - %s -> %s", resource.Guid, resource.Name))
		domainList = append(domainList, resource.Name)
		domainsGuidMap[resource.Name] = resource.Guid
	}

	return domainsGuidMap, domainList

}

// if route reserved -> wrong.
// https://v3-apidocs.cloudfoundry.org/version/3.197.0/index.html#check-reserved-routes-for-a-domain
func check_route_reserved(cliConnection plugin.CliConnection, host string, domain string, domainGuid string) (status bool) {
	print_debug(fmt.Sprintf("  Checking route reservation via cf api for '%s.%s'", host, domain))
	// url := fmt.Sprintf("curl -H \"Authorization: $(cf oauth-token)\" %s/v3/domains/%s/route_reservations\\?host\\=%s", cfApiEndpoint, domainGuid, host)
	url := fmt.Sprintf("/v3/domains/%s/route_reservations?host=%s", domainGuid, host)
	// fmt.Println(url)
	output, err := cliConnection.CliCommandWithoutTerminalOutput(append([]string{"curl", url})...)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		os.Exit(1)
	}

	jsonStr := strings.Join(output, "") // Combine output lines
	var jsonObj RouteReservationOutput
	err = json.Unmarshal([]byte(jsonStr), &jsonObj)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		// os.Exit(1)
	}

	return jsonObj.MatchingRoute
}
