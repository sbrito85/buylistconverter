# Buylist converter for Magic TCG

## Why?

When submitting to buy lists, card shops will use different naming conventions for their sets.
This makes taking an exported csv from apps like Manabox and TCGPlayer difficult.

The initial version currently support cardkingdom buylist formats

Currently, TCGPlayer Exports are tested, although your csv export only needs to have the following fields

* Name: The name of the card
* SetCode: The set code(IE FIN, FIC)
* CardNumber: The collector number of the card
* Printing: Foil/Non-Foil. If "Printing" is "Foil", we search a different link
* Quantity: The quantity of the specific card you have

## Usage

Export your tcgplayer list and make sure it has the fields above before running the following

```
buylistconverter -file path/to/tcgplayer.csv
```

Note: By default, it will look in the current directory for tcgplayer.csv if the file flag is not passed.

A file named "CardKingdomBuyList.csv" will be placed in your current directory

## Future ideas

* Caching of data: Currently we are sending requests for every card. We can cache sets to send less requests.
* Web UI: Not everyone wants to run command line tools
* Other buylist
* Make use of the cards we know aren't currently being accepted