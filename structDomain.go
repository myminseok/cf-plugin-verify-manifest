package main

// the struct implementing the interface defined by the core CLI.
// It can be found at  "code.cloudfoundry.org/cli/plugin/plugin.go"
type VerifyManifestPlugin struct {
}

type YamlManifest struct {
	Applications []YamlApplication `yaml:"applications"`
}

type YamlApplication struct {
	Name     string                 `yaml:"name"`
	Env      map[string]interface{} `yaml:"env"`
	Services []string               `yaml:"services"`
	Routes   []YamlRoute            `yaml:"routes"`
}

type YamlRoute struct {
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

type DomainsResources struct {
	Resources []struct {
		Guid string `json:"guid"`
		Name string `json:"name"`
	} `json:"resources"`
}

type DomainsPagination struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
		TotalPages   int `json:"total_pages"`
	} `json:"pagination"`
}
type RouteReservationOutput struct {
	MatchingRoute bool `json:"matching_route"`
}
