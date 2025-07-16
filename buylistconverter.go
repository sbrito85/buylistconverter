package main

import (
	"flag"

	"github.com/sbrito85/buylistconverter/processors"
	"github.com/sbrito85/buylistconverter/translators"
)

func main() {
	csvlist := flag.String("file", "TCGplayer.csv", "Path to the file TCGPlayer export")

	sellList := processors.ProcessCSV(*csvlist)

	ckTranslate := translators.NewCKTranslator()
	ckTranslate.TranslateBuyList(sellList)
}
