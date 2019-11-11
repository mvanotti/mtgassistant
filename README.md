# MTG Arena Assistant
Helper code for Magic The Gathering: Arena

# Requirements
You will need a computer that can run `Magic The Gathering: Arena`, and set it up so it exports logs.
To do so, open the game, go to Settings > View Account > Enable Logs, then restart the game and exit.

These programs work by parsing the resource files inside your installation of MTG: Arena, as well as
the game logs. If those things are not in the standard locations, you will need to specify those to
the programs via command-line flags.

# What can I do?
Right now the assistant only has two binaries: a collection exporter, and a deck helper.

## Collection Exporter
Collection Exporter is a program that will parse your MTG:A collection and print it in the MTG:A format.
To run it, just do:

```
$ go run collectionexporter/main.go
```

## Deck Helper
Given a decklist in the MTG:A format, deck helper will parse your MTG:A collection and will tell you
which cards are missing from the decks (cards that you need to craft). Currently it only supports the
standard rotation, but that can be easily changed from the code.

To run it:

```
$ go run deckhelper.go -deck=<path-to-your-deck>
```

# Libraries

There's a `carddb` library that parses the resource files and creates a database of magic cards. You can
use it to make queries based on card names, card ids, or just iterate over it and run the code that you
want.

There's also a `collectionfinder` library that parses the game logs and gets your card collection.