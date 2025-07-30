package processors

import (
	"os"

	"github.com/gocarina/gocsv"
)

type SellListItem struct {
	Name       string `csv:"Name"`
	SetCode    string `csv:"Set Code"`
	CardNumber string `csv:"Card Number"`
	Printing   string `csv:"Printing"`
	Quantity   int    `csv:"Quantity"`
}

func ProcessCSV(path string) []SellListItem {
	in, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	var sellList []SellListItem

	if err := gocsv.UnmarshalFile(in, &sellList); err != nil {
		panic(err)
	}
	return sellList
}
