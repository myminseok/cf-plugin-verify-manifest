package main

type Domains struct {
	Resources []struct {
		Guid string `json:"guid"`
		Name string `json:"name"`
	} `json:"resources"`
}
type RouteReservationOutput struct {
	MatchingRoute bool `json:"matching_route"`
}
