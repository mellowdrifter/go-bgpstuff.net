package bgpstuff

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/mellowdrifter/bogons"
)

const (
	version = "1.0.0"
	api     = "https://bgpstuff.net"
)

// Client is a client to the bgpstuff.net REST API
type Client struct {
	Loc string
}

type response struct {
	Data data `json:"Response"`
}

// data is the struct received on each successul query.
// Copied directly from https://github.com/mellowdrifter/bgpstuff.netv2/blob/master/pkg/models/models.go
type data struct {
	Action    string     // What action is being performed?
	Route     string     // /route response
	ASPath    []string   // /aspath response
	ASSet     []string   // /aspath response
	Origin    string     // /origin response
	ROA       string     // /roa response
	ASName    string     // /asname response
	ASLocale  string     // /asname locale
	Invalids  []Invalids // /invalids response
	Sourced   Sourced    // /sourced response
	Location  Location   // /whereami response
	Totals    Totals     // /totals response
	IP        string     // IP address being queried
	Exists    bool       // Specifies if there was an actual reply
	CacheTime time.Time  // If set, this is how old the entry is in the cache
}

// Sourced contains the amount of IPv4 and IPv6 prefixes.
// As well as the prefixes.
type Sourced struct {
	Ipv4, Ipv6 uint32
	Prefixes   []string
}

// Totals contains the amount of IPv4 and IPv6 prefixes in the RIB.
// Also the unix timestamp
type Totals struct {
	Ipv4, Ipv6 uint32
	Time       uint64
}

// Location contains the coordinates and map of the ingress location.
type Location struct {
	Lat, Long     string
	City, Country string
	Map           string // a base64 encoded png
}

// Invalids contains all the ROA invalids prefixes originated by an ASN.
type Invalids struct {
	ASN      string
	Prefixes []string
}

// getRequest will take a handler and any arugments and request
// a response from the bgpstuff.net API. Timeouts are set to 5 seconds
// to prevent hanging connections.
func (c *Client) getRequest(urls ...string) (response, error) {
	var client = &http.Client{
		Timeout: time.Second * 5,
	}

	u, err := url.Parse(api)
	if err != nil {
		log.Fatal(err)
	}

	for _, url := range urls {
		u.Path = path.Join(u.Path, url)
	}

	var resp response
	re, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return response{}, err
	}
	re.Header.Set("Content-Type", "application/json")

	res, err := client.Do(re)
	if err != nil {
		return response{}, err
	}

	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return response{}, err
	}

	return resp, nil
}

func (c Client) GetRoute(ip string) (*net.IPNet, error) {
	if !bogons.ValidPublicIP(ip) {
		return nil, fmt.Errorf("Not a valid IP")
	}
	p := net.ParseIP(ip)
	resp, err := c.getRequest("route", p.String())
	if err != nil {
		return nil, err
	}
	_, ipnet, err := net.ParseCIDR(getRouteFromResponse(resp))
	if err != nil {
		return nil, err
	}

	return ipnet, nil
}

func getRouteFromResponse(res response) string {
	return res.Data.Route
}
