package bgpstuff

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mellowdrifter/bogons"
	"golang.org/x/time/rate"
)

const (
	version = "1.0.0"
	liveapi = "https://bgpstuff.net"
	testapi = "https://test.bgpstuff.net"
)

var (
	errInvalidIP  = errors.New("invalid IP")
	errInvalidASN = errors.New("invalid AS Number")
	rpm           = 30 // requests per minute
)

// Client is a client to the bgpstuff.net REST API
type Client struct {
	Loc      string
	limiter  *rate.Limiter
	api      string
	ASNames  map[int]string
	Invalids map[int][]*net.IPNet
}

// NewBGPClient return a pointer to a new client
// TODO: Hate setting testing here...
func NewBGPClient(testing bool) *Client {
	r := rate.Every(time.Minute / time.Duration(rpm))
	limit := rate.NewLimiter(r, rpm)

	var api string
	if testing {
		api = testapi
	} else {
		api = liveapi
	}

	return &Client{
		limiter: limit,
		api:     api,
	}
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

func (c *Client) getURI(urls []string) string {
	var uri strings.Builder
	uri.WriteString(c.api)
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
	if err := c.limiter.Wait(context.Background()); err != nil {
		return nil, err
	}
	client := newHTTPClient(time.Second * 8)

	uri := c.getURI(urls)

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
		return nil, fmt.Errorf("received status: %s (%d)", http.StatusText(res.StatusCode), res.StatusCode)
	}

	defer res.Body.Close()

	var resp response
	if err := resp.decodeJSON(res.Body); err != nil {
		return &resp, err
	}

	return &resp, nil
}

// GetRoute uses the /route handler
func (c *Client) GetRoute(ip string) (*net.IPNet, error) {
	if !bogons.ValidPublicIP(ip) {
		return nil, errInvalidIP
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("route", p.String())
	if err != nil {
		return nil, err
	}

	// Response could be no route.
	if resp.Data.Route == "" {
		return nil, nil
	}

	// TODO: stop returning "/0"
	if resp.Data.Route == "/0" {
		return nil, err
	}

	_, ipnet, err := net.ParseCIDR(resp.Data.Route)
	if err != nil {
		return nil, err
	}

	return ipnet, nil
}

// GetOrigin uses the /origin handler.
func (c *Client) GetOrigin(ip string) (int, error) {
	if !bogons.ValidPublicIP(ip) {
		return 0, errInvalidIP
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("origin", p.String())
	if err != nil {
		return 0, err
	}

	return resp.Data.Origin, nil
}

func getASPathFromResponse(res *response) ([]int, []int) {
	if len(res.Data.ASPath) == 0 {
		return nil, nil
	}
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

	return path, nil
}

// GetASPath uses the /aspath handler.
func (c *Client) GetASPath(ip string) ([]int, []int, error) {
	if !bogons.ValidPublicIP(ip) {
		return nil, nil, errInvalidIP
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("aspath", p.String())
	if err != nil {
		return nil, nil, err
	}

	paths, sets := getASPathFromResponse(resp)
	return paths, sets, nil
}

// GetROA uses the /roa handler.
func (c *Client) GetROA(ip string) (string, error) {
	if !bogons.ValidPublicIP(ip) {
		return "", errInvalidIP
	}

	p := net.ParseIP(ip)
	resp, err := c.getRequest("roa", p.String())
	if err != nil {
		return "", err
	}

	// If there is no origin, there is no prefix ROA to check.
	if resp.Data.Origin == 0 {
		return "", nil
	}

	return resp.Data.ROA, nil
}

// GetASName uses the /asname handler
func (c *Client) GetASName(asn int) (string, error) {
	if !bogons.ValidPublicASN(uint32(asn)) {
		return "", errInvalidASN
	}

	// Check asnames if it has the entry
	if len(c.ASNames) > 1 {
		if name, ok := c.ASNames[asn]; ok {
			return name, nil
		}
		return "", nil
	}

	resp, err := c.getRequest("asname", fmt.Sprint(asn))
	if err != nil {
		return "", err
	}

	return resp.Data.ASName, nil
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
func (c *Client) GetInvalid(asn int) ([]*net.IPNet, error) {
	if !bogons.ValidPublicASN(uint32(asn)) {
		return nil, errInvalidASN
	}

	if c.Invalids == nil {
		return nil, fmt.Errorf("invalids is empty, run GetInvalids() first")
	}

	return c.Invalids[asn], nil
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
