package translators

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time" // Import the time package

	"github.com/gocarina/gocsv"
	"github.com/sbrito85/buylistconverter/mtgjson"
	"github.com/sbrito85/buylistconverter/processors"
)

type SCGTranslator struct {
	Client     *http.Client
	RateLimit  time.Duration
	SCGBuyList []SCGBuyListCard
	MtgSets    *mtgjson.MTGSets
	RejectList []processors.SellListItem
}

type SCGBuyListCard struct {
	Quantity  int    `csv: "quantity"`
	Productid string `csv: "productid"`
	Language  string `csv: "language"`
	Finish    string `csv: "finish"`
}

func NewSCGTranslator(mtgSets *mtgjson.MTGSets) SCGTranslator {
	return SCGTranslator{
		Client:    &http.Client{},
		MtgSets:   mtgSets,
		RateLimit: time.Second * 2,
	}
}

var scgCardUrl = "https://www.starcitygames.com/"

func (s *SCGTranslator) TranslateBuyList(sellList []processors.SellListItem) {
	fmt.Printf("Evaluating %v cards for Star City Games Buylist\n", len(sellList))
	for _, v := range sellList {
		//octomancer-sgl-mtg-blc-037-enn
		replacer := strings.NewReplacer(" ", "-", "'", "", ":", "", ",", "", "/", "")
		SCGCardName := replacer.Replace(v.Name)
		SCGCardNumber := v.CardNumber
		if len(SCGCardNumber) == 2 {
			SCGCardNumber = fmt.Sprintf("0%v", SCGCardNumber)
		}
		if len(SCGCardNumber) == 1 {
			SCGCardNumber = fmt.Sprintf("00%v", SCGCardNumber)
		}
		SCGFoil := "n"
		if v.Printing == "Foil" {
			SCGFoil = "f"
		}
		productID := strings.ToUpper(fmt.Sprintf("sgl-mtg-%s-%s-en%s1", v.SetCode, SCGCardNumber, SCGFoil))
		productIDURL := fmt.Sprintf("%s-sgl-mtg-%s-%s-en%s", SCGCardName, v.SetCode, SCGCardNumber, SCGFoil)
		url := fmt.Sprintf("%s%s", scgCardUrl, productIDURL)
		for i := 1; i < 4; i++ {
			if i > 1 {
				productID = strings.ToUpper(fmt.Sprintf("SGL-MTG-%s%v-%s-en%s1", v.SetCode, i, SCGCardNumber, SCGFoil))
				url = fmt.Sprintf("%s%s-sgl-mtg-%s%v-%s-en%s", scgCardUrl, SCGCardName, v.SetCode, i, SCGCardNumber, SCGFoil)
			}
			if s.scgCardInfoFromURL(url) == 200 {
				fmt.Printf("Found card on SCG: %s\n", productID)
				s.SCGBuyList = append(s.SCGBuyList, SCGBuyListCard{
					Productid: productID,
					Language:  "en",
					Finish:    foiling(SCGFoil),
					Quantity:  v.Quantity,
				})
				break
			}
		}
	}

	err := s.writeSCGCSVToDisk()
	if err != nil {
		fmt.Printf("Error writing to disk: %v\n", err)
		return
	}
}

func (s *SCGTranslator) writeSCGCSVToDisk() error {
	if len(s.SCGBuyList) != 0 {
		file, err := os.OpenFile("StarCityGamesBuylist.csv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		err = gocsv.MarshalFile(&s.SCGBuyList, file) // Pass a pointer to the slice
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SCGTranslator) scgCardInfoFromURL(url string) int {
	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print("Error creating request: ", err)

	}

	// Set a User-Agent header to mimic a web browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Perform the request
	res, err := s.Client.Do(req)
	if err != nil {
		fmt.Print("Error fetching URL: ", err)

	}
	if res.StatusCode == 429 {
		fmt.Println("Rate limit exceeded, waiting for 5 minutes and trying again...")
		time.Sleep(s.RateLimit * 360)
		res, err = s.Client.Do(req)
		if err != nil {

		}
	}
	if res.StatusCode == 404 {
		return res.StatusCode
	}
	if res.StatusCode != 200 {
		fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	defer res.Body.Close()
	return res.StatusCode
}

func foiling(s string) string {
	if s == "f" {
		return "Foil"
	}
	return "Non-foil"
}
