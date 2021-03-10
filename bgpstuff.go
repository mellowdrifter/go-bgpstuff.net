package bgpstuff

import (
	"errors"
	"fmt"
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

var (
	errInvalidIP  = errors.New("Invalid IP")
	errInvalidASN = errors.New("Invalid AS Number")
)

// Client is a client to the bgpstuff.net REST API
type Client struct {
	Loc      string
	ASNames  map[int]string
	Invalids map[int][]*net.IPNet
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

// getRequest will take a handler and any arugments and request
// a response from the bgpstuff.net API. Timeouts are set to 5 seconds
// to prevent hanging connections.
func (c Client) getRequest(urls ...string) (*response, error) {
	client := newHTTPClient(time.Second * 8)

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
		return nil, false, errInvalidIP
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
		return 0, false, errInvalidIP
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
		return nil, nil, false, errInvalidIP
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
		return "", false, errInvalidIP
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

// GetASName uses the /asname handler
func (c *Client) GetASName(asn int) (string, bool, error) {
	if !bogons.ValidPublicASN(uint32(asn)) {
		return "", false, errInvalidASN
	}

	// Check asnames if it has the entry
	if len(c.ASNames) > 1 {
		if name, ok := c.ASNames[asn]; ok {
			return name, ok, nil
		}
		return "", false, nil
	}

	resp, err := c.getRequest("asname", fmt.Sprint(asn))
	if err != nil {
		return "", false, err
	}

	exists := getExistsFromResponse(resp)
	if !exists {
		return "", false, err
	}

	return resp.Data.ASName, exists, nil
}

// GetASNames uses the /asnames handler
func (c *Client) GetASNames() error {
	c.ASNames = make(map[int]string)

	resp, err := c.getRequest("asnames")
	if err != nil {
		return err
	}

	for _, v := range resp.Data.ASNames {
		c.ASNames[int(v.ASN)] = v.ASName
	}

	return nil
}

// GetInvalids grabs all current invalids and populates c.Invalids
func (c *Client) GetInvalids() error {
	c.Invalids = make(map[int][]*net.IPNet)

	resp, err := c.getRequest("invalids")
	if err != nil {
		return err
	}

	for _, v := range resp.Data.Invalids {
		prefixes := make([]*net.IPNet, 0, len(v.Prefixes))
		for _, prefix := range v.Prefixes {
			_, ipnet, err := net.ParseCIDR(prefix)
			if err != nil {
				return err
			}
			prefixes = append(prefixes, ipnet)
		}
		c.Invalids[int(v.ASN)] = prefixes
	}
	return nil
}

// GetInvalid implements the /invalid handler
func (c *Client) GetInvalid(asn int) ([]*net.IPNet, bool, error) {
	if !bogons.ValidPublicASN(uint32(asn)) {
		return nil, false, errInvalidASN
	}

	if c.Invalids == nil {
		return nil, false, fmt.Errorf("invalids is empty, run GetInvalids() first")
	}

	val, ok := c.Invalids[asn]
	return val, ok, nil
}

// GetSourced implements the /sourced handler
func (c *Client) GetSourced(asn int) ([]*net.IPNet, int, int, error) {
	if !bogons.ValidPublicASN(uint32(asn)) {
		return nil, 0, 0, errInvalidASN
	}

	resp, err := c.getRequest("sourced", fmt.Sprint(asn))
	if err != nil {
		return nil, 0, 0, err
	}

	prefixes := make([]*net.IPNet, 0, len(resp.Data.Sourced.Prefixes))
	for _, v := range resp.Data.Sourced.Prefixes {
		_, prefix, err := net.ParseCIDR(v)
		if err != nil {
			return nil, 0, 0, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, resp.Data.Sourced.Ipv4, resp.Data.Sourced.Ipv6, nil
}

// GetTotals implements the /totals handler
func (c *Client) GetTotals() (int, int, error) {
	resp, err := c.getRequest("totals")
	if err != nil {
		return 0, 0, err
	}

	return resp.Data.Totals.Ipv4, resp.Data.Totals.Ipv6, nil
}
