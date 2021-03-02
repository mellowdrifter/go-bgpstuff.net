package bgpstuff

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mellowdrifter/bogons"
)

const (
	version = "1.0.0"
	api     = "https://test.bgpstuff.net"
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
	ASPath    []string   `json:"ASPath"` //aspath response
	ASSet     []string   `json:"ASSet"`
	Origin    int        `json:"Origin,string,omitempty"`
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

//NewBGPClient return a pointer to a new client
func NewBGPClient() *Client {
	return &Client{}
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

func getURI(urls []string) string {
	var uri strings.Builder
	uri.WriteString(api)
	for _, v := range urls {
		uri.WriteString("/")
		uri.WriteString(v)
	}

	return uri.String()
}

// decodeJSON will populate a response struct with the body of the reply from the server.
// Returns an error if it cannot unmarshal.
func (res *response) decodeJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(res)
}

// getRequest will take a handler and any arugments and request
// a response from the bgpstuff.net API. Timeouts are set to 5 seconds
// to prevent hanging connections.
func (c Client) getRequest(urls ...string) (*response, error) {
	client := newHTTPClient(time.Second * 5)

	uri := getURI(urls)

	re, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	re.Header.Set("Content-Type", "application/json")
	re.Header.Set("User-Agent", fmt.Sprintf("go-bgpstuff.net/%s", version))

	res, err := client.Do(re)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received status: %s (%d)", http.StatusText(res.StatusCode), res.StatusCode)
	}

	defer res.Body.Close()

	var resp response
	if err := resp.decodeJSON(res.Body); err != nil {
		return &resp, err
	}

	return &resp, nil
}

// GetRoute uses the /route handler
func (c *Client) GetRoute(ip string) (*net.IPNet, bool, error) {
	if !bogons.ValidPublicIP(ip) {
		return nil, false, fmt.Errorf("Not a valid IP")
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("route", p.String())
	if err != nil {
		return nil, false, err
	}

	exists := getExistsFromResponse(resp)
	if !exists {
		return nil, false, err
	}

	_, ipnet, err := net.ParseCIDR(resp.Data.Route)
	if err != nil {
		return nil, exists, err
	}

	return ipnet, exists, nil
}

// GetOrigin uses the /origin handler.
func (c *Client) GetOrigin(ip string) (int, bool, error) {
	if !bogons.ValidPublicIP(ip) {
		return 0, false, fmt.Errorf("Not a valid IP")
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("origin", p.String())
	if err != nil {
		return 0, false, err
	}

	exists := getExistsFromResponse(resp)
	if !exists {
		return 0, false, err
	}

	origin := getOriginFromResponse(resp)
	if origin == 0 {
		return 0, false, fmt.Errorf("Unable to parse origin AS number")
	}

	return origin, exists, nil
}

func getExistsFromResponse(res *response) bool {
	return res.Data.Exists
}

func getOriginFromResponse(res *response) int {
	return res.Data.Origin
	/*o, err := strconv.Atoi(res.Data.Origin)
	if err != nil {
		return 0
	}
	return o*/
}

func getASPathFromResponse(res *response) ([]int, []int) {
	path := make([]int, 0, len(res.Data.ASPath))
	for _, v := range res.Data.ASPath {
		i, _ := strconv.Atoi(v)
		path = append(path, i)
	}
	if len(res.Data.ASSet) > 0 {
		as := make([]int, 0, len(res.Data.ASSet))
		for _, v := range res.Data.ASSet {
			i, _ := strconv.Atoi(v)
			as = append(as, i)
		}
		return path, as
	}

	return path, []int{}
}

// GetASPath uses the /aspath handler.
func (c *Client) GetASPath(ip string) ([]int, []int, bool, error) {
	if !bogons.ValidPublicIP(ip) {
		return nil, nil, false, fmt.Errorf("Not a valid IP")
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("aspath", p.String())
	if err != nil {
		return nil, nil, false, err
	}

	exists := getExistsFromResponse(resp)
	if !exists {
		return nil, nil, false, err
	}

	paths, sets := getASPathFromResponse(resp)
	return paths, sets, exists, nil
}

// GetROA uses the /roa handler.
func (c *Client) GetROA(ip string) (string, bool, error) {
	if !bogons.ValidPublicIP(ip) {
		return "", false, fmt.Errorf("Not a valid IP")
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("roa", p.String())
	if err != nil {
		return "", false, err
	}

	exists := getExistsFromResponse(resp)
	if !exists {
		return "", false, err
	}

	return resp.Data.ROA, exists, nil
}
