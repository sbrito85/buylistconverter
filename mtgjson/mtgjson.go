package mtgjson

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MTGSets struct {
	SealedProducts map[string][]SealedProduct
	sets           map[string][]Card
	SetNames       map[string]string
}

type MTGJSONResponse struct {
	Data SetData `json:"data"`
}

type SetData struct {
	Name           string          `json:"name"`
	Cards          []Card          `json:"cards"`
	SealedProducts []SealedProduct `json:"sealedProduct"`
}

type SealedProduct struct {
	Name         string            `json:"name"`
	SubType      string            `json:"subtype"`
	Category     string            `json:"category"`
	PurchaseUrls map[string]string `json:"purchaseUrls"`
}

// Card represents a single Magic: The Gathering card.
type Card struct {
	HasFoil      bool              `json:"hasFoil"`
	HasNonFoil   bool              `json:"hasNonFoil"`
	IsPromo      bool              `json:"isPromo"`
	Name         string            `json:"name"`
	Number       string            `json:"number"`
	SetCode      string            `json:"setCode"`
	PurchaseUrls map[string]string `json:"purchaseUrls"`
	SetName      string
	Printing     string
}

type SellCard struct {
	ProductLine  string
	Name         string
	Number       string
	SetCode      string
	PurchaseUrls map[string]string
	SetName      string
	Printing     string
	Quantity     int
}

const mtgJSONApi = "https://mtgjson.com/api/v5/%s.json"

// InitMTGJSON initializes the MTGSets cache so we don't fetch mtgjson data every time.
func InitMTGJSON() *MTGSets {
	var mtgSets MTGSets
	mtgSets.sets = make(map[string][]Card)
	mtgSets.SetNames = make(map[string]string)
	mtgSets.SealedProducts = make(map[string][]SealedProduct)
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
			// Special cards may not have a foil link.
			if printing == "Foil" && v.PurchaseUrls["cardKingdomFoil"] != "" {
				return v.PurchaseUrls["cardKingdomFoil"]
			}
			return v.PurchaseUrls["cardKingdom"]
		}
	}
	return ""
}

func (m *MTGSets) FetchSellCardInfo(cardnumber, setCode, printing string, quantity int) *SellCard {
	if _, ok := m.sets[setCode]; !ok {
		if err := m.FetchSet(setCode); err != nil {
			log.Printf("Error fetching set %s: %v", setCode, err)
			return nil
		}
	}
	for _, v := range m.sets[setCode] {
		if cardnumber == v.Number {
			return &SellCard{
				ProductLine: "Magic",
				Name:        v.Name,
				Number:      v.Number,
				SetCode:     v.SetCode,
				SetName:     m.SetNames[setCode],
				Printing:    printing,
				Quantity:    quantity,
			}
		}
	}
	return nil
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
	m.SetNames[setCode] = mtgjson.Data.Name
	m.SealedProducts[setCode] = mtgjson.Data.SealedProducts

	return nil
}
