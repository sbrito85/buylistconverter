package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sbrito85/buylistconverter/processors"
	"github.com/sbrito85/buylistconverter/translators"
)

func main() {

	csvlist := flag.String("file", "TCGplayer.csv", "Path to the file TCGPlayer export")
	flag.Parse()
	_, err := os.Stat(*csvlist) // Attempt to get file info
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {

			fmt.Printf("File %v does not exist\n", *csvlist) // File does not exist
			os.Exit(1)
		}
	}
	sellList := processors.ProcessCSV(*csvlist)

	ckTranslate := translators.NewCKTranslator()
	ckTranslate.TranslateBuyList(sellList)
}
