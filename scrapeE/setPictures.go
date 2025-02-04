package scrapeE

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
)

type setInfo struct {
	SetName     string `json:"SetName"`
	SetAmmount  int    `json:"SetAmmount"`
	ProductInfo []productInfo
}
type productInfo struct {
	ProductName string `json:"ProductName"`
	ProductID   int    `json:"ProductID"`
}

var SetRequestURL = "https://mp-search-api.tcgplayer.com/v1/search/request?q=&isList=false"
var imageurl1 = "https://tcgplayer-cdn.tcgplayer.com/product/"
var imageurl2 = "_in_1000x1000.jpg"

func DownloadSetsInfo(category string, sets []string) {

	ctx, cancel := context.WithCancel(context.Background())
	go TerminalLoad(ctx)
	proxyfile, err := os.Open("Proxy.txt")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			proxyfile = nil
		} else {
			log.Panicln("Failed opening Proxy.txt, error unrelated to file not found \n", err)
		}
	}
	var proxylist []string
	if proxyfile != nil {
		scanner := bufio.NewScanner(proxyfile)
		for scanner.Scan() {
			proxylist = append(proxylist, scanner.Text())
		}
	}
	log.Println("loaded proxies (If applicable")
	//I tried to do an annonymous function within the client creation call for some reason, so then i copy pasted the
	//code out of it... and made it work? this looks interesting so im leaving this in.
	apxy := func() string {
		if proxyfile == nil {
			return ""
		} else {
			return proxylist[0]
		}
	}
	aclient, err := NewClientWithProxy(apxy())
	ErrorCheck(err, "Failed to create the initial client to pull entire list of set")
	areq, err := NewScrapeHTTPRequest("POST")
	ErrorCheck(err, "Unable to create new http request for initial set pull")
	err = SetScrapeURL(areq, SetRequestURL)
	ErrorCheck(err, "Failed to set URL")

	var workinglist []setInfo

	for _, value := range sets {
		err = SetScrapeBody(areq, 0, category, value)
		ErrorCheck(err, "Unable to set scrape body when iterating through sets")
		resp, err := aclient.Do(areq)
		ErrorCheck(err, "Client failed to produce request when scraping sets contents")
		srVar := QueryResponse{}
		err = json.NewDecoder(resp.Body).Decode(&srVar)
		ErrorCheck(err, "Unable to decode set query response")

		workinglist = append(workinglist, struct {
			SetName     string `json:"SetName"`
			SetAmmount  int    `json:"SetAmmount"`
			ProductInfo []productInfo
		}{SetName: value, SetAmmount: srVar.Results[0].TotalResults})
		log.Println("Got set amount ", value, " ", srVar)

	}
	wg := sync.WaitGroup{}
	lenwg := len(workinglist)
	wg.Add(lenwg)
	idchannel := make(chan setInfo, lenwg)
	for _, x := range workinglist {
		go downloadSetsImageIDs(
			proxylist[rand.Intn(len(proxylist)-1)],
			category,
			x,
			idchannel,
			&wg)

	}
	wg.Wait()
	var idlists []setInfo
	for range workinglist {
		x := <-idchannel
		idlists = append(idlists, x)
	}
	if err = createDir(category); err != nil {
		log.Panic(err)
	}

	wg.Add(len(idlists))
	for _, v := range idlists {
		go imageScraper(proxylist, category, v, &wg)
	}
	wg.Wait()
	cancel()
	ClearTerminal()

}

func downloadSetsImageIDs(proxies string, category string, set setInfo, idchannel chan setInfo, wg *sync.WaitGroup) {
	var idLists []productInfo
	for workers := 0; workers < set.SetAmmount; workers += 24 {
		client, err := NewClientWithProxy(proxies)
		ErrorCheck(err, "Unable to create client for downloadsetsImageIds method")
		req, err := NewScrapeHTTPRequest("POST")
		ErrorCheck(err, "failed creating http request for downloadSetsImageIDs")
		err = SetScrapeURL(req, SetRequestURL)
		ErrorCheck(err, "Failed to set the RequestURL")
		err = SetScrapeBody(req, workers, category, set.SetName)
		ErrorCheck(err, "Failed to set the scrapebody")
		resp, err := client.Do(req)
		ErrorCheck(err, "Failed to do request at client.do")
		resvar := QueryResponse{}
		err = json.NewDecoder(resp.Body).Decode(&resvar)
		ErrorCheck(err, "Unable to decode set query response")
		for _, v := range resvar.Results[0].Results {
			prodID := int(v.ProductID)
			var product = productInfo{
				ProductName: v.ProductName,
				ProductID:   prodID,
			}
			idLists = append(idLists, product)
		}

	}
	set.ProductInfo = idLists
	idchannel <- set
	wg.Done()

}

func imageScraper(proxies []string, category string, set setInfo, wg *sync.WaitGroup) {
	dir := category + "/" + set.SetName
	if err := createDir(dir); err != nil {
		wg.Done()
		log.Panicln(err, "\n Unable to create directory for set: ", set.SetName)
	}
	sync := sync.WaitGroup{}
	sync.Add(set.SetAmmount)
	for _, v := range set.ProductInfo {
		go workers(proxies[rand.Intn(len(proxies)-1)], dir, v, &sync)
	}
	sync.Wait()
	wg.Done()

}

func workers(proxy string, dir string, product productInfo, wg *sync.WaitGroup) {

	client, err := NewClientWithProxy(proxy)
	ErrorCheck(err, "failed to create client with proxy for worker")
	ul := imageurl1 + strconv.Itoa(product.ProductID) + imageurl2
	ulc, err := url.Parse(ul)
	ErrorCheck(err, "Unable to parse url for product id on worker")
	req := http.Request{
		Method: "GET",
		URL:    ulc,
	}
	ErrorCheck(err, "unable to create get request for worker")
	resp, err := client.Do(&req)
	if resp.StatusCode != http.StatusOK {
		log.Println("Worker rate limtied on request to download id ", product.ProductID)
		wg.Done()
		return
	}
	filen := dir + "/" + product.ProductName + "-" + strconv.Itoa(product.ProductID) + ".jpg"
	file, err := os.Create(filen)
	if err != nil {
		log.Println("Worker unable to create file to write to with product id", product.ProductID)
		wg.Done()
		return
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		log.Println("Worker unable to create file to write to with product id", product.ProductID)
		wg.Done()
		return
	}
	if err := file.Close(); err != nil {
		log.Println(err, "unable to close file? wtf")
		wg.Done()
		return
	}
	wg.Done()
	return
}

func createDir(dirPath string) error {
	_, err := os.Stat(dirPath)

	if os.IsNotExist(err) {
		log.Printf("Creating directory: %s\n", dirPath)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		// Other error occurred while checking directory
		return fmt.Errorf("error checking directory: %w", err)
	} else {
		log.Printf("Directory already exists: %s\n", dirPath)
	}

	return nil
}
