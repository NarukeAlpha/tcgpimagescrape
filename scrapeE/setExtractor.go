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

type RequestBody struct {
	Algorithm     string          `json:"algorithm,omitempty"`
	From          int             `json:"from,omitempty"`
	Size          int             `json:"size,omitempty"`
	Filters       Filters         `json:"filters,omitempty"`
	ListingSearch ListingSearch   `json:"listingSearch,omitempty"`
	Context       Context         `json:"context,omitempty"`
	Settings      Settings        `json:"settings,omitempty"`
	Sort          json.RawMessage `json:"sort,omitempty"` // Assuming sort is dynamic or empty
}

type Filters struct {
	Term struct {
		ProductLineName []string `json:"productLineName"`
		SetName         []string `json:"setName"`
	} `json:"term"`
	Range json.RawMessage `json:"range,omitempty"`
	Match json.RawMessage `json:"match,omitempty"`
}

type ListingSearch struct {
	Context ListingContext `json:"context,omitempty"`
	Filters struct {
		Term struct {
			SellerStatus string `json:"sellerStatus"`
			ChannelId    int    `json:"channelId"`
		} `json:"term"`
		Range   json.RawMessage `json:"range"`
		Exclude struct {
			ChannelExclusion int `json:"channelExclusion"`
		} `json:"exclude"`
	} `json:"filters"`
}

type ListingContext struct {
	Cart json.RawMessage `json:"cart,omitempty"` // Placeholder for potential nested objects
}

type Context struct {
	Cart            json.RawMessage `json:"cart,omitempty"` // Placeholder for potential nested objects
	ShippingCountry string          `json:"shippingCountry,omitempty"`
	UserProfile     json.RawMessage `json:"userProfile,omitempty"` // Placeholder for future fields
}

type Settings struct {
	UseFuzzySearch bool            `json:"useFuzzySearch"`
	DidYouMean     json.RawMessage `json:"didYouMean,omitempty"` // Placeholder for potential nested objects
}

type QueryResponse struct {
	Results []QueryResult `json:"results"`
}

type QueryResult struct {
	Aggregations Aggregations `json:"aggregations"`
	Results      []Product    `json:"results"`
	Algorithm    string       `json:"algorithm"`
	SearchType   string       `json:"searchType"`
	TotalResults int          `json:"totalResults"`
	ResultID     string       `json:"resultId"`
}

type Aggregations struct {
	CardType        []Aggregation `json:"cardType"`
	RarityName      []Aggregation `json:"rarityName"`
	SetName         []Aggregation `json:"setName"`
	ProductTypeName []Aggregation `json:"productTypeName"`
	ProductLineName []Aggregation `json:"productLineName"`
	Condition       []Aggregation `json:"condition"`
	Language        []Aggregation `json:"language"`
	Printing        []Aggregation `json:"printing"`
}

type Aggregation struct {
	UrlValue string  `json:"urlValue"`
	IsActive bool    `json:"isActive"`
	Value    string  `json:"value"`
	Count    float64 `json:"count"`
}

type Product struct {
	Listings                []interface{}    `json:"listings"` // Assuming listings is dynamic
	ShippingCategoryID      float64          `json:"shippingCategoryId"`
	Duplicate               bool             `json:"duplicate"`
	ProductLineUrlName      string           `json:"productLineUrlName"`
	ProductUrlName          string           `json:"productUrlName"`
	ProductTypeID           float64          `json:"productTypeId"`
	RarityName              string           `json:"rarityName"`
	Sealed                  bool             `json:"sealed"`
	MarketPrice             float64          `json:"marketPrice"`
	CustomAttributes        CustomAttributes `json:"customAttributes"`
	LowestPriceWithShipping float64          `json:"lowestPriceWithShipping"`
	ProductName             string           `json:"productName"`
	SetID                   float64          `json:"setId"`
	ProductID               float64          `json:"productId"`
	MedianPrice             float64          `json:"medianPrice"`
	Score                   float64          `json:"score"`
	SetName                 string           `json:"setName"`
	FoilOnly                bool             `json:"foilOnly"`
	SetUrlName              string           `json:"setUrlName"`
	SellerListable          bool             `json:"sellerListable"`
	TotalListings           float64          `json:"totalListings"`
	ProductLineID           float64          `json:"productLineId"`
	ProductStatusID         float64          `json:"productStatusId"`
	ProductLineName         string           `json:"productLineName"`
	MaxFulfillableQuantity  float64          `json:"maxFulfillableQuantity"`
	LowestPrice             float64          `json:"lowestPrice"`
}

type CustomAttributes struct {
	Description  string      `json:"description"`
	Attribute    []string    `json:"attribute"`
	ReleaseDate  interface{} `json:"releaseDate"` // null or a date string
	Number       string      `json:"number"`
	CardType     []string    `json:"cardType"`
	MonsterType  []string    `json:"monsterType"`
	CardTypeB    string      `json:"cardTypeB"`
	RarityDbName string      `json:"rarityDbName"`
	Level        string      `json:"level"`
	Defense      string      `json:"defense"`
	LinkArrows   interface{} `json:"linkArrows"` // null or another type
	FlavorText   interface{} `json:"flavorText"` // null or string
	Attack       string      `json:"attack"`
}

var ContentType = "application/json; charset=utf-8"
var Userag = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
var Apiver = "1.0"

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
		Context: Context{
			ShippingCountry: "US", // Default to "US"
		},
		Settings: Settings{
			UseFuzzySearch: true,
		},
	}
	return body
}

// NewScrapeHTTPRequest returns a new http.request type filled with the default configuration and a error
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
func SetScrapeBody(r *http.Request, s ...string) error {
	b := NewRequestBody()
	b.Filters.Term.ProductLineName = []string{s[0]}
	if len(s) > 1 {
		for i := 1; i < len(s); i++ {
			b.Filters.Term.SetName = []string{s[i]}
		}
	}
	jd, err := json.Marshal(b)
	if err != nil {
		return errors.New(fmt.Sprintln("Cannot marshal body: \n", b, "\nInto Json format"))
	}

	r.Body = io.NopCloser(strings.NewReader(string(jd)))
	r.ContentLength = int64(len(jd))
	return nil
}

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
