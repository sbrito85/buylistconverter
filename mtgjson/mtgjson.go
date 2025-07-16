package mtgjson

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MTGSets struct {
	sets map[string]Card
}

type MTGJSONResponse struct {
	Data SetData `json:"data"`
}

type SetData struct {
	Cards []Card `json:"cards"`
}

// Card represents a single Magic: The Gathering card.
type Card struct {
	HasFoil      bool              `json:"hasFoil"`
	HasNonFoil   bool              `json:"hasNonFoil"`
	IsPromo      bool              `json:"isPromo"`
	Name         string            `json:"name"`
	Number       string            `json:"number"`
	PurchaseUrls map[string]string `json:"purchaseUrls"`
}

const mtgJSONApi = "https://mtgjson.com/api/v5/%s.json"

func CardKingdomUrl(cardnumber, setcode, printing string) string {
	client := &http.Client{}
	//u, _ := url.Parse(filepath.Join(mtgJSONApi, "FIC.json"))
	req, err := http.NewRequest("GET", fmt.Sprintf(mtgJSONApi, setcode), nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var mtgjson MTGJSONResponse
	err = json.NewDecoder(res.Body).Decode(&mtgjson)
	if err != nil {
		fmt.Errorf("%v", err)
	}

	for _, v := range mtgjson.Data.Cards {
		if cardnumber == v.Number {
			if printing == "Foil" {
				return v.PurchaseUrls["cardKingdomFoil"]
			}
			return v.PurchaseUrls["cardKingdom"]
		}
	}
	return ""
}
