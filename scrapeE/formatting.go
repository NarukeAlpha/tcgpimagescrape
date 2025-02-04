package scrapeE

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Response struct {
	Errors  []string `json:"errors"`
	Results []Result `json:"results"`
}

type Result struct {
	CategoryId          int    `json:"categoryId"`
	CatalogGroupId      int    `json:"catalogGroupId"`
	CategoryName        string `json:"categoryName"`
	DisplayName         string `json:"displayName"`
	UrlName             string `json:"urlName"`
	CategoryDescription string `json:"categoryDescription"`
	CategoryPageTitle   string `json:"categoryPageTitle"`
	MpCanSearch         bool   `json:"mpCanSearch"`
}

type RequestBody struct {
	Algorithm     string          `json:"algorithm,omitempty"`
	From          int             `json:"from,omitempty"`
	Size          int             `json:"size,omitempty"`
	Filters       Filters         `json:"filters,omitempty"`
	ListingSearch ListingSearch   `json:"listingSearch,omitempty"`
	Context       ContextBody     `json:"context,omitempty"`
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

type ContextBody struct {
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

func ClearTerminal() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Println("Unable to clear terminal", err)
	}
}

func ErrorCheck(err error, errstrng string) {
	if err != nil {
		log.Panicln(errstrng, err)
	}
}

func TerminalLoad(ctx context.Context) {
	fmt.Print("Scraping and Downloading......")
	for ctx.Err() != nil {
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading.     ")
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading..    ")
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading...   ")
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading....  ")
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading..... ")
		time.Sleep(250)
		fmt.Print("\rScraping and Downloading......")
	}
}
