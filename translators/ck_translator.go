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
	Client     *http.Client
	RateLimit  time.Duration
	BuyList    []BuyListCard
	MtgSets    *mtgjson.MTGSets
	RejectList []processors.SellListItem
}

func NewCKTranslator(mtgSets *mtgjson.MTGSets) CKTranslator {
	return CKTranslator{
		Client:    &http.Client{},
		MtgSets:   mtgSets,
		RateLimit: time.Second * 3,
	}
}

func (c *CKTranslator) TranslateBuyList(sellList []processors.SellListItem) {
	fmt.Printf("Evaluating %v cards for Card Kingdom Buylist\n", len(sellList))
	for _, v := range sellList {
		// Introduce a delay before each request
		time.Sleep(c.RateLimit)

		url := c.MtgSets.CardKingdomUrl(v.CardNumber, v.SetCode, v.Printing)
		if url == "" {
			c.RejectList = append(c.RejectList, v)
			fmt.Printf("Skipping %s:%s\n", v.Name, v.SetCode)
			continue
		}
		cardName, edition, err := c.cardInfoFromURL(url)
		if err != nil {
			c.RejectList = append(c.RejectList, v)
			fmt.Printf("Error fetching card info for %s: %v\n", v.Name, edition)
			continue
		}
		if cardName == "" {
			c.RejectList = append(c.RejectList, v)
			fmt.Printf("%s is not being accepted for trade in\n", v.Name)
			continue
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

	err := c.writeCSVToDisk()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v cards added to buylist\n", len(c.BuyList))
	fmt.Printf("%v cards added to reject list\n", len(c.RejectList))
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
