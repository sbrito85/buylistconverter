# Buylist converter for Magic TCG

## Why?

When submitting to buy lists, card shops will use different naming conventions for their sets.
This makes taking an exported csv from apps like Manabox and TCGPlayer difficult as well as keeping track of which cards are going to which buy list.

The initial version currently support cardkingdom buylist formats

Currently, TCGPlayer Exports are tested, although your csv export only needs to have the following fields

* Name: The name of the card
* Set Code: The set code(IE FIN, FIC)
* Card Number: The collector number of the card
* Printing: Foil/Non-Foil. If "Printing" is "Foil", we search a different link
* Quantity: The quantity of the specific card you have

## Usage

Export your tcgplayer list(or list from another app) and make sure it has the fields above before running.

Buy default we will build a Card Kingdom buy list and using the reject list we will build a start city list.

### Run All
```
buylistconverter -file path/to/tcgplayer.csv,path/to/other.csv
```

### Run only Card Kingdom buylist
```
buylistconverter -file path/to/tcgplayer.csv,path/to/other.csv -starcity=false
```

### Run only Star City buylist
```
buylistconverter -file path/to/tcgplayer.csv,path/to/other.csv -cardkingdom=false
```

Note: By default, it will look in the current directory for tcgplayer.csv if the file flag is not passed.

## Files

All files are placed in the working directory where you run the program.

- **CardKingdomBuyList.csv** A list formatted for Card Kingdoms buylist upload.
- **CardKingdomRejectList.csv** A list of cards that Card Kingdom is not accepting.
- **StarCityGamesBuyList** A List formatted for star city games buylist. We currently don't have a good way to tell if they are accepting a card or not.

### Example CSV

Name | Set Code | Card Number | Printing | Quantity
---  | ---      | ---         | ---      |--- 
Espers to Magicite | FIC | 114 | Foil | 1
Summon: Fat Chocobo | FIN | 371 | Foil | 1
"Zidane, Tantalus Thief" | FIN |405 |  | 1
The Masamune | FIN | 353 | Foil |1
