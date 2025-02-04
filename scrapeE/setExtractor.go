package scrapeE

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewClientWithProxy creates a new http client with a proxy. To not use a Proxy simply pass "" as the proxy string
func NewClientWithProxy(proxy string) (*http.Client, error) {
	if proxy == "" {
		return &http.Client{
			Timeout: 10 * time.Minute,
		}, nil
	} else {
		proxyURL, err := parseProxy(proxy)
		if err != nil {
			return nil, errors.New(fmt.Sprintln("Error parsing proxy: ", err))
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		return &http.Client{
			Transport: transport,
			Timeout:   10 * time.Minute,
		}, nil
	}

}
func parseProxy(proxyStr string) (*url.URL, error) {
	parts := strings.Split(proxyStr, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid proxy format. Expected ip:port:user:password")
	}
	ip := parts[0]
	port := parts[1]
	user := parts[2]
	password := parts[3]
	// Construct proxy URL with authentication
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", ip, port),
		User:   url.UserPassword(user, password),
	}
	return proxyURL, nil
}

// NewScrapeHTTPRequest Invoke with http method and returns a new http.request type filled with the request  and a error
func NewScrapeHTTPRequest(method string) (*http.Request, error) {
	req, err := http.NewRequest(method, "www.google.com", nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintln("Error creating request: ", err))
	}
	req.Header.Set("Content-Type", ContentType)
	req.Header.Set("User-Agent", Userag)
	req.Header.Set("Api-Supported-Version", Apiver)
	return req, nil
}

// NewRequestBody is the body to be able to query the API
func NewRequestBody() RequestBody {
	body := RequestBody{
		Algorithm: "sales_dismax",
		From:      0,
		Size:      24,
		Filters: Filters{
			Term: struct {
				ProductLineName []string `json:"productLineName"`
				SetName         []string `json:"setName"`
			}{
				ProductLineName: []string{""},
				SetName:         []string{},
			},
		},
		ListingSearch: ListingSearch{
			Filters: struct {
				Term struct {
					SellerStatus string `json:"sellerStatus"`
					ChannelId    int    `json:"channelId"`
				} `json:"term"`
				Range   json.RawMessage `json:"range"`
				Exclude struct {
					ChannelExclusion int `json:"channelExclusion"`
				} `json:"exclude"`
			}{
				Term: struct {
					SellerStatus string `json:"sellerStatus"`
					ChannelId    int    `json:"channelId"`
				}{
					SellerStatus: "Live",
					ChannelId:    0,
				},
				Range: json.RawMessage(`{"quantity": {"gte": 1}}`),
				Exclude: struct {
					ChannelExclusion int `json:"channelExclusion"`
				}{
					ChannelExclusion: 0,
				},
			},
		},
		Context: ContextBody{
			ShippingCountry: "US", // Default to "US"
		},
		Settings: Settings{
			UseFuzzySearch: true,
		},
	}
	return body
}

// SetScrapeURL sets the URL of the request to the given string
func SetScrapeURL(r *http.Request, ur string) error {
	u, err := url.Parse(ur)
	if err != nil {
		log.Panicln("Unable to parse URL\n", err)
	}
	r.URL = u
	if r.URL.Host != u.Host {
		return errors.New(fmt.Sprintln("Hosts do not match \nSetScrapeIDURL failed with URL:  ", ur))
	}
	return nil
}

// SetScrapeBody sets the body of the request entirely, takes either just the product name or various sets to dynamically be re-used
func SetScrapeBody(r *http.Request, x int, s ...string) error {
	b := NewRequestBody()
	b.Filters.Term.ProductLineName = []string{s[0]}
	if len(s) > 1 {
		for i := 1; i < len(s); i++ {
			b.Filters.Term.SetName = []string{s[i]}
		}
	}
	if x > 0 {
		b.From = x
	}
	jd, err := json.Marshal(b)
	if err != nil {
		return errors.New(fmt.Sprintln("Cannot marshal body: \n", b, "\nInto Json format"))
	}

	r.Body = io.NopCloser(strings.NewReader(string(jd)))
	r.ContentLength = int64(len(jd))
	return nil
}

// NewCRB returns a client and a request, with the url and body set
func NewCRB(proxy string, method string, from int, s ...string) {

}
