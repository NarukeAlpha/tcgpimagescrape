package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	sc "tcgpImageScrape/scrapeE"
)

var (
	CatalogURL    = "https://mpapi.tcgplayer.com/v2/Catalog/CategoryFilters?mpfev=3118"
	SetRequestURL = "https://mp-search-api.tcgplayer.com/v1/search/request?q=&isList=false"
)

func init() {
	logfile, err := os.OpenFile("log"+strconv.Itoa(time.Now().Hour())+"-"+strconv.Itoa(time.Now().YearDay())+"-"+time.Now().Month().String()+"-"+strconv.Itoa(time.Now().Year())+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Panic(err)
	}
	//mw := io.MultiWriter(os.Stdout, logfile)
	log.SetOutput(logfile)

}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to the TCG Player set image scraper \n press Crt + C to close at any time")

	client, err := sc.NewClientWithProxy("")
	sc.ErrorCheck(err, "Unable to create client with proxy")

	catrq, err := sc.NewScrapeHTTPRequest("GET")
	sc.ErrorCheck(err, "Unable to create request for category")

	err = sc.SetScrapeURL(catrq, CatalogURL)
	sc.ErrorCheck(err, "Unable to set category URL")

	response := sc.Response{}
	res, err := client.Do(catrq)
	err = json.NewDecoder(res.Body).Decode(&response)
	sc.ErrorCheck(err, "Unable to decode categories")

	for {
		fmt.Println("Loading Categories")
		time.Sleep(1000)

		sc.ClearTerminal()
		sort.SliceStable(response.Results, func(i, j int) bool {
			return response.Results[i].CategoryId < response.Results[j].CategoryId
		})
		midpoint := len(response.Results) / 2
		for i, result := range response.Results {
			if i == midpoint {
				fmt.Println(i+1, ") ", result.DisplayName)
				break
			} else {
				fmt.Printf("%-3v) %-30v  %-2v) %-15v \n", strconv.Itoa(i+1), result.DisplayName, strconv.Itoa(i+1+midpoint), response.Results[i+midpoint].DisplayName)
			}
		}
		scanner.Scan()
		input := scanner.Text()
		categorySelection, err := strconv.Atoi(input)
		if input == "exit" {
			os.Exit(0)
		}
		if err != nil || categorySelection > len(response.Results) {
			fmt.Println("Please enter a valid number: ")
			time.Sleep(3000)
			sc.ClearTerminal()
			continue
		}
		setRequest, err := sc.NewScrapeHTTPRequest("POST")
		sc.ErrorCheck(err, "Unable to create request for set")
		err = sc.SetScrapeBody(setRequest, 0, response.Results[categorySelection-1].UrlName)
		sc.ErrorCheck(err, "Unable to set body for set request")

		err = sc.SetScrapeURL(setRequest, SetRequestURL)
		sc.ErrorCheck(err, "Unable to set URL for set scrape")

		setresp, err := client.Do(setRequest)
		sc.ErrorCheck(err, "Unable to do set request")
		srVar := sc.QueryResponse{}

		err = json.NewDecoder(setresp.Body).Decode(&srVar)
		sc.ErrorCheck(err, "Unable to decode set query response")
		setlist := srVar.Results[0].Aggregations.SetName
		divider := len(setlist) / 2
		var setliststring []string
		for i, result := range setlist {
			setliststring = append(setliststring, result.Value)
			if i == 0 || i%2 == 0 {
				if i == divider && len(setlist)%i == 1 {
					fmt.Printf("%-3v) %-30v \n", strconv.Itoa(i+1), result.Value)
					break
				}
				if i == divider {
					break
				}
				var sv, fv string
				if i == 0 {
					fv = "1"
					sv = "2"
				} else {
					fv = strconv.Itoa(i + 1)
					sv = strconv.Itoa(i + 2)
				}
				fmt.Printf("%-3v) %-80v %-3v) %-15v\n", fv, result.Value, sv, setlist[i+1].Value)
			}

		}
		scanner.Scan()
		input = scanner.Text()
		finput := strings.Fields(input)
		sc.ClearTerminal()
		var inputlist []string
		for _, v := range finput {
			value, err := strconv.Atoi(v)
			if err != nil {
				log.Panicln(err)
			}
			inputlist = append(inputlist, setliststring[value])
		}

		sc.DownloadSetsInfo(response.Results[categorySelection-1].UrlName, inputlist)

	}
}
