package bgpstuff

import (
	"encoding/json"
	"io"
	"time"
)

type response struct {
	Data data `json:"Response"`
}

// data is the struct received on each successul query.
// Copied directly from https://github.com/mellowdrifter/bgpstuff.netv2/blob/master/pkg/models/models.go
type data struct {
	Action    string      // What action is being performed?
	Route     string      // /route response
	ASPath    []string    `json:"ASPath"` //aspath response
	ASSet     []string    `json:"ASSet"`
	Origin    int         `json:"Origin,string"`
	ROA       string      `json:"ROA"`    // /roa response
	ASName    string      `json:"ASName"` // /asname response
	ASLocale  string      // /asname locale
	ASNames   []ASNumName `json:"ASNames"` // /asnames response
	Invalids  []Invalids  // /invalids response
	Sourced   Sourced     // /sourced response
	Location  Location    // /whereami response
	Totals    Totals      // /totals response
	IP        string      // IP address being queried
	Exists    bool        // Specifies if there was an actual reply
	CacheTime time.Time   // If set, this is how old the entry is in the cache
}

// Sourced contains the amount of IPv4 and IPv6 prefixes.
// As well as the prefixes.
type Sourced struct {
	Ipv4, Ipv6 int
	Prefixes   []string
}

// Totals contains the amount of IPv4 and IPv6 prefixes in the RIB.
// Also the unix timestamp
type Totals struct {
	Ipv4, Ipv6 int
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
	ASN      int `json:"ASN,string"`
	Prefixes []string
}

//ASNumName contains an AS number, name, and locale.
type ASNumName struct {
	ASN      uint32
	ASName   string
	ASLocale string
}

// decodeJSON will populate a response struct with the body of the reply from the server.
// Returns an error if it cannot unmarshal.
func (res *response) decodeJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(res)
}
