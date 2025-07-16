package processors

import (
	"github.com/gocarina/gocsv"
	"os"
)

type SellListItem struct {
	Name       string `csv:"Simple Name"`
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
