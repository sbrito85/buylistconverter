package translators

import (
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
	Client    *http.Client
	RateLimit time.Duration
	BuyList   []BuyListCard
	MtgSets   *mtgjson.MTGSets
}

func NewCKTranslator(mtgSets *mtgjson.MTGSets) CKTranslator {
	return CKTranslator{
		Client:    &http.Client{},
		MtgSets:   mtgSets,
		RateLimit: time.Second,
	}
}

func (c *CKTranslator) TranslateBuyList(sellList []processors.SellListItem) {
	for _, v := range sellList {
		// Introduce a delay before each request
		time.Sleep(c.RateLimit)

		url := c.MtgSets.CardKingdomUrl(v.CardNumber, v.SetCode, v.Printing)
		if url == "" {
			fmt.Printf("Skipping %s:%s\n", v.Name, v.SetCode)
			continue
		}
		// Create a new GET request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Set a User-Agent header to mimic a web browser
		req.Header.Set("User-Agent", "MTG Buylist Converter/1.0")

		// Perform the request
		res, err := c.Client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		//body, err := io.ReadAll(res.Body)
		if res.StatusCode != 200 {
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}

		// Parse the HTML using goquery
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		// Find the edition
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

		if cardName == "" {
			fmt.Printf("%s is not being accepted for trade in\n", v.Name)
		}
		foil := "0"
		if v.Printing == "Foil" {
			foil = "1"
		}
		c.BuyList = append(c.BuyList, BuyListCard{
			Title:    cardName,
			Edition:  edition,
			Foil:     foil,
			Quantity: v.Quantity,
		})
	}
	c.writeCSVToDisk()
}

func (c *CKTranslator) writeCSVToDisk() error {
	file, err := os.OpenFile("CardKingdomBuylist.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	defer file.Close()
	err = gocsv.MarshalFile(&c.BuyList, file) // Pass a pointer to the slice
	if err != nil {
		fmt.Errorf("%v", err)
	}

	return nil
}
