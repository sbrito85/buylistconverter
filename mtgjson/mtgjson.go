package mtgjson

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MTGSets struct {
	sets map[string][]Card
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

// InitMTGJSON initializes the MTGSets cache so we don't fetch mtgjson data every time.
func InitMTGJSON() *MTGSets {
	var mtgSets MTGSets
	mtgSets.sets = make(map[string][]Card)
	return &mtgSets
}

func (m *MTGSets) CardKingdomUrl(cardnumber, setcode, printing string) string {

	if _, ok := m.sets[setcode]; !ok {
		if err := m.FetchSet(setcode); err != nil {
			log.Printf("Error fetching set %s: %v", setcode, err)
			return ""
		}
	}
	for _, v := range m.sets[setcode] {
		if cardnumber == v.Number {
			if printing == "Foil" {
				return v.PurchaseUrls["cardKingdomFoil"]
			}
			return v.PurchaseUrls["cardKingdom"]
		}
	}
	return ""
}

func (m *MTGSets) FetchSet(setCode string) error {
	client := &http.Client{}
	fmt.Printf("Fetching MTGJSON data for set: %v\n", setCode)
	req, err := http.NewRequest("GET", fmt.Sprintf(mtgJSONApi, setCode), nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var mtgjson MTGJSONResponse
	err = json.NewDecoder(res.Body).Decode(&mtgjson)
	if err != nil {
		return err
	}
	m.sets[setCode] = mtgjson.Data.Cards

	return nil
}
