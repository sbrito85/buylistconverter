package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sbrito85/buylistconverter/mtgjson"
	"github.com/sbrito85/buylistconverter/processors"
	"github.com/sbrito85/buylistconverter/translators"
)

func main() {
	var sellList []processors.SellListItem
	csvlist := flag.String("file", "TCGplayer.csv", "Path to the file TCGPlayer export")
	flag.Parse()
	for _, v := range strings.Split(*csvlist, ",") {
		_, err := os.Stat(v) // Attempt to get file info
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Printf("File %v does not exist\n", v) // File does not exist
				continue
			}
		}

		sellList = append(sellList, processors.ProcessCSV(v)...)
	}
	mtgSets := mtgjson.InitMTGJSON()
	ckTranslate := translators.NewCKTranslator(mtgSets)
	ckTranslate.TranslateBuyList(sellList)
}
