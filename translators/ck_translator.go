package translators

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time" // Import the time package

	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
	"github.com/sbrito85/buylistconverter/mtgjson"
	"github.com/sbrito85/buylistconverter/processors"
)

type BuyListCard struct {
	Title    string `csv: "title"`
	Edition  string `csv: "edition"`
	Foil     string `csv: "foil"`
	Quantity int    `csv: "Quantity"`
}

type CKTranslator struct {
	Client     *http.Client
	RateLimit  time.Duration
	BuyList    []BuyListCard
	MtgSets    *mtgjson.MTGSets
	RejectList []processors.SellListItem
}

type CKPriceListResponse struct {
	Data []Item `json:"data"`
}

type Item struct {
	Scryfall_id string `json:"scryfall_id"`
	IsFoil      string `json:"is_foil"`
	QtyBuying   int    `json:"qty_buying"`
	PriceBuy    string `json:"price_buy"`
	Name        string `json:"name"`
	Edition     string `json:"edition"`
	Variation string `json:"variation"`
}

func NewCKTranslator(mtgSets *mtgjson.MTGSets) CKTranslator {
	return CKTranslator{
		Client:    &http.Client{},
		MtgSets:   mtgSets,
		RateLimit: time.Second * 3,
	}
}

const ckAPI = "https://api.cardkingdom.com/api/v2/pricelist"

func (c *CKTranslator) TranslateBuyList(sellList []processors.SellListItem) {
	fmt.Printf("Evaluating %v cards for Card Kingdom Buylist\n", len(sellList))
	var priceList []Item
	priceList, err := c.ckPriceList()
	if err != nil {
		fmt.Printf("Error fetching Card Kingdom price list: %v\n", err)
		return
	}
	for _, v := range sellList {
		// Introduce a delay before each request
		for _, card := range priceList {
			scryfallId := c.MtgSets.GetScryfallid(v.CardNumber, v.SetCode, v.Printing)
			if scryfallId == "" {
				continue
			}
			if card.Scryfall_id == scryfallId && checkPrinting(v.Printing) == card.IsFoil && card.QtyBuying > 0 {
				fmt.Printf("Accepting %s(%s) qty:%d at %s\n", card.Name, v.SetCode, card.QtyBuying, card.PriceBuy)
				foil := "0"
				if v.Printing == "Foil" {
					foil = "1"
				}

				name := card.Name
				if card.Variation != "" {
					name = fmt.Sprintf("%s (%s)", card.Name, card.Variation )
				}
				c.BuyList = append(c.BuyList, BuyListCard{
					Title:    name,
					Edition:  card.Edition,
					Foil:     foil,
					Quantity: v.Quantity,
				})
				continue
			}
		}
		c.RejectList = append(c.RejectList, v)
	}

	err = c.writeCSVToDisk()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v cards added to buylist\n", len(c.BuyList))
	fmt.Printf("%v cards added to reject list\n", len(c.RejectList))
}

func checkPrinting(printing string) string {
	if printing == "Foil" {
		return "true"
	}
	return "false"
}

func (c *CKTranslator) ckPriceList() ([]Item, error) {
	fmt.Printf("Fetching Card Kingdom price list\n")
	req, err := http.NewRequest("GET", ckAPI, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var ckprices CKPriceListResponse
	err = json.NewDecoder(res.Body).Decode(&ckprices)
	if err != nil {
		return nil, err
	}
	return ckprices.Data, nil
}

func (c *CKTranslator) cardInfoFromURL(url string) (string, string, error) {
	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print("Error creating request: ", err)
		return "", "", err
	}

	// Set a User-Agent header to mimic a web browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Perform the request
	res, err := c.Client.Do(req)
	if err != nil {
		fmt.Print("Error fetching URL: ", err)
		return "", "", err
	}
	if res.StatusCode == 429 {
		fmt.Println("Rate limit exceeded, waiting for 5 minutes and trying again...")
		time.Sleep(c.RateLimit * 360)
		res, err = c.Client.Do(req)
		if err != nil {
			return "", "", err
		}
	}

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	cardName := doc.Find(".sellCardName").Text()
	edition := doc.Find(".editionLink").Text()
	if edition == "" {
		doc.Find("ul.breadcrumb li.breadcrumb-item a").Each(func(i int, s *goquery.Selection) {
			linkHref, exists := s.Attr("href")
			if exists && linkHref != "/" && linkHref != "/mtg" {
				if len(s.Text()) > 0 {
					edition = s.Text()
					return
				}
			}
		})
	}

	return cardName, edition, nil
}

func (c *CKTranslator) writeCSVToDisk() error {
	if len(c.BuyList) != 0 {
		file, err := os.OpenFile("CardKingdomBuylist.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		err = gocsv.MarshalFile(&c.BuyList, file) // Pass a pointer to the slice
		if err != nil {
			return err
		}
	}
	if len(c.RejectList) != 0 {
		file, err := os.OpenFile("CardKingdomRejectList.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		err = gocsv.MarshalFile(&c.RejectList, file) // Pass a pointer to the slice
		if err != nil {
			return err
		}
	}
	return nil
}
