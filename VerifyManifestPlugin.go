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
		manifestPath, _ := ParseArgs(args)
		manifest := LoadYAML(manifestPath)
		if !CheckManifestAppName(manifestPath, manifest) {
			os.Exit(1)
		}

		print_cf_target(cliConnection, args)
		var all_good = true
		all_good = Check_services(cliConnection, manifestPath, manifest)
		all_good = Check_routes(cliConnection, manifestPath, manifest)
		if !all_good {
			os.Exit(1)
		}
	}
}

func print_usage() {
	fmt.Println("Usage: cf verify-manifest -f PATH_TO_MANIFEST [-debug true]")
}

func ParseArgs(args []string) (string, error) {

	flags := flag.NewFlagSet("verify-manifest", flag.ContinueOnError)
	manifestPath := flags.String("f", "", "path to an application manifest")
	debugModeFlag := flags.String("debug", "", "verbose plugin")
	err := flags.Parse(args[1:])
	// if err != nil {
	// 	// fmt.Printf("[ERROR] %s\n", err)
	// 	// print_usage()
	// 	// os.Exit(1)
	// }

	if err != nil {
		return "", err
	}

	if *manifestPath == "" {
		print_usage()
		return "", fmt.Errorf("Missing manifest argument")
	}

	if len(*debugModeFlag) > 0 {
		fmt.Println("[INFO] found '-debug' flag, enabling print debug mode")
		PRINT_DEBUG = true
	}
	fmt.Printf("Using manifestPath: '%s'\n", *manifestPath)
	return *manifestPath, err
}

func print_debug(arg string) {
	if PRINT_DEBUG {
		fmt.Printf("    [DEBUG]%s\n", arg)
	}
}

func LoadYAML(manifestPath string) (manifest YamlManifest) {
	b, err := ioutil.ReadFile(manifestPath)

	if err != nil {
		fmt.Errorf("[ERROR] Unable to read manifest file: %s", manifestPath)
		os.Exit(1)
	}

	var document YamlManifest
	err = yaml.Unmarshal(b, &document)

	if err != nil {
		fmt.Errorf("[ERROR] Unable to parse manifest file: %s", manifestPath)
		os.Exit(1)
	}
	return document
}

func CheckManifestAppName(manifestPath string, manifest YamlManifest) (status bool) {
	status = true
	fmt.Println("  Checking manifest app name ...")
	for i, app := range manifest.Applications {
		print_debug(fmt.Sprintf("  index: %v appname:%s", i, app.Name))
		if len(app.Name) == 0 {
			fmt.Printf("[ERROR] Invalid App name for index '%v' in %s", i, manifestPath)
			status = false
			break
		}
	}
	return status
}

func ParseManifestServices(manifestPath string, manifest YamlManifest) (manifestServices []ManifestService) {
	if !CheckManifestAppName(manifestPath, manifest) {
		return manifestServices
	}
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

func ParseManifestRoutes(manifestPath string, manifest YamlManifest) (manifestRoutes []ManifestRoute) {
	if !CheckManifestAppName(manifestPath, manifest) {
		return manifestRoutes
	}
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
	fmt.Print("cf target ...")
	output, err := cliConnection.CliCommandWithoutTerminalOutput(append([]string{"target"})...)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		os.Exit(1)
	}
	for _, item := range output {
		fmt.Println("  ", item)
	}
}

func Fetch_cf_services(cliConnection plugin.CliConnection) (cf_services []string) {
	fmt.Println("  Fetching cf services from the target foundation ...")
	cfServices, err := cliConnection.GetServices()
	if err != nil {
		fmt.Errorf("[ERROR] %s", err)
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
			return true
		}
	}
	return false
}

func Check_routes(cliConnection plugin.CliConnection, manifestPath string, manifest YamlManifest) (status bool) {
	status = true
	fmt.Println("\nChecking Routes availability specified in the manifest from the target ... ", manifestPath)
	manifestRoutes := ParseManifestRoutes(manifestPath, manifest)
	domainsGuidMap, domainList := Fetch_cf_domains_guid(cliConnection)
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
			if Check_route_reserved(cliConnection, host, domain, domainsGuidMap[domain]) {
				badList = append(badList, AppRouteResult{appName: app, route: route, message: fmt.Sprintf("Reserved is reserved")})
			} else {
				goodList = append(goodList, AppRouteResult{appName: app, route: route, message: ""})
			}
		}
	}
	if len(goodList) > 0 {
		fmt.Println("  [GOOD] Available routes specified in the manifest in the target foundation:")
		for _, item := range goodList {
			fmt.Printf("  - good app: '%s', route: '%s'\n", item.appName, item.route)
		}
	}
	if len(badList) > 0 {
		status = false
		fmt.Println("  [ERROR] Not Available routes specified in the manifest in the target foundation:")
		for _, item := range badList {
			fmt.Printf("  - error app: '%s', route: '%s' -> %s\n", item.appName, item.route, item.message)
		}
	}
	return status
}

func Check_services(cliConnection plugin.CliConnection, manifestPath string, manifest YamlManifest) (status bool) {
	status = true

	fmt.Println("\nChecking Service instance from the manifest ...", manifestPath)
	manifestServices := ParseManifestServices(manifestPath, manifest)
	cf_services := Fetch_cf_services(cliConnection)

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
		fmt.Println("  [GOOD] Service instance specified in manifest that exists in current space:")
		for _, item := range goodList {
			fmt.Printf("  - good app: '%s', service: '%s'\n", item.appName, item.service)
		}
	}

	if len(badList) > 0 {
		status = false
		fmt.Println("  [ERROR] Missing Service instance specified in manifest but not exist in current space:")
		for _, item := range badList {
			fmt.Printf("  - error app: '%s', service: '%s'\n", item.appName, item.service)
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
func Fetch_cf_domains_guid(cliConnection plugin.CliConnection) (domainsGuidMap map[string]string, domainList []string) {
	fmt.Println("  Fetching cf domains, guid from the target foundation ...")

	var paginationObj DomainsPagination
	var resourcesObj DomainsResources
	currentPage := 1
	for true {
		url := fmt.Sprintf("/v3/domains?page=%v&per_page=50", currentPage)
		print_debug(fmt.Sprintf("  url:%s", url))
		output, err := cliConnection.CliCommandWithoutTerminalOutput(append([]string{"curl", url})...)
		print_debug(fmt.Sprintf("  %s", output))
		if err != nil {
			fmt.Println("[ERROR] ", err)
			os.Exit(1)
		}

		jsonStr := strings.Join(output, "") // Combine output lines
		err = json.Unmarshal([]byte(jsonStr), &paginationObj)
		if err != nil {
			fmt.Println("[ERROR] Unmarshal DomainsPagination: ", err)
			os.Exit(1)
		}

		err = json.Unmarshal([]byte(jsonStr), &resourcesObj)
		if err != nil {
			fmt.Println("[ERROR] Unmarshal DomainsResources: ", err)
			os.Exit(1)
		}

		domainsGuidMap = make(map[string]string)
		for _, resource := range resourcesObj.Resources {
			print_debug(fmt.Sprintf("  - %s -> %s", resource.Guid, resource.Name))
			domainList = append(domainList, resource.Name)
			domainsGuidMap[resource.Name] = resource.Guid
		}

		// process the next pages if exists.
		totalResults := paginationObj.Pagination.TotalResults
		totalPages := paginationObj.Pagination.TotalPages
		print_debug(fmt.Sprintf("  TotalResults:%v totalPages:%v currentPage:%v", totalResults, totalPages, currentPage))
		if currentPage >= totalPages {
			break
		}
		currentPage++
	}

	return domainsGuidMap, domainList

}

// if route reserved -> wrong.
// https://v3-apidocs.cloudfoundry.org/version/3.197.0/index.html#check-reserved-routes-for-a-domain
func Check_route_reserved(cliConnection plugin.CliConnection, host string, domain string, domainGuid string) (status bool) {
	print_debug(fmt.Sprintf("  Checking route reservation via cf api for '%s.%s'", host, domain))
	url := fmt.Sprintf("/v3/domains/%s/route_reservations?host=%s", domainGuid, host)
	print_debug(fmt.Sprintf(" check_route_reserved url: %s", url))
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
		// os.Exit(1) // continue on error
	}

	return jsonObj.MatchingRoute
}
