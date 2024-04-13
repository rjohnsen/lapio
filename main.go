package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/rjohnsen/lapio/kernel"
)

/*
 * Main application logic
 */

func main() {
	parser := argparse.NewParser("lapio", "Lapio - Log Shovel. Shovel logs into Elastic Search")
	parserDirectivePath := parser.String("d", "directive", &argparse.Options{
		Required: true,
		Help:     "Path to parser directive",
	})

	logPath := parser.String("l", "logpath", &argparse.Options{
		Required: true,
		Help:     "Path to log file",
	})

	indexName := parser.String("i", "index", &argparse.Options{
		Required: true,
		Help:     "Name of index (where to store logs)",
	})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
	} else {
		var parserDirective, err = kernel.LoadParserDirective(*parserDirectivePath)

		if err != nil {
			fmt.Println(fmt.Errorf("[ ERR ] %s", err))
			os.Exit(1)
		}

		// Get Elastic credentials
		settings, err := kernel.LoadSettings("elastic.toml")

		if err != nil {
			fmt.Println(fmt.Errorf("[ ERR ] %s", err))
			os.Exit(1)
		}

		fmt.Printf("Index: %s\n", *indexName)
		fmt.Printf("Parser Directive: %s\n", parserDirective.Name)
		fmt.Printf("Log file: %s\n", *logPath)
		fmt.Printf("Elastic: %s\n", settings.Host)
		fmt.Println("-------------------------------------------------------\n")

		parserStatus, err := kernel.ParseLog(parserDirective, settings, *indexName, *logPath)

		if err != nil {
			fmt.Println(fmt.Errorf("[ ERR ] %s", err))
			os.Exit(1)
		}

		fmt.Println("\nIndexing done. Here are som vital stats:")
		fmt.Printf("Entries total: %d\n", parserStatus.RowCount)
		fmt.Printf("Entries indexed: %d\n", parserStatus.IndexedEntries)
		fmt.Printf("Errors: %d\n", parserStatus.ErrorCount)

		if parserStatus.ErrorCount > 0 {
			fmt.Println("\nPlease see file 'error.data' in current directory for any unparsed log entries. Please adjust your regexes accordingly.")
		}
	}
}
