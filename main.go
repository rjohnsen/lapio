package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/akamensky/argparse"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rjohnsen/lapio/kernel"
)

/*
 * Parse log
 */

func parse_log(parser_directive kernel.ParserDirective, index_name *string, log_path *string) {
	// Get Elastic credentials
	elastic_credentials, err := kernel.LoadSettings("elastic.toml")

	if err != nil {
		fmt.Println(fmt.Errorf("[ ERR ] %s", err))
		os.Exit(1)
	}

	fmt.Printf("Index: %s\n", *index_name)
	fmt.Printf("Parser Directive: %s\n", parser_directive.Name)
	fmt.Printf("Log file: %s\n", *log_path)

	file, err := os.Open(*log_path)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Write lines with error to file for follow up
	error_file, err := os.OpenFile("error.data", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
	}

	defer error_file.Close()

	// Tracking errors
	errors := 0
	entries_indexed := 0
	line_counter := 0

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			elastic_credentials.Host,
		},
		Username: elastic_credentials.Username,
		Password: elastic_credentials.Password,
	})

	if err != nil {
		fmt.Println("Error creating the client: %s", err)
	}

	for scanner.Scan() {
		line := scanner.Text()
		line_counter += 1

		for index, regx := range parser_directive.Regexes {
			re := regexp.MustCompile(regx.Expression)
			matches := re.FindStringSubmatch(line)

			if len(matches) == regx.CaptureGroups+1 {
				hasher := md5.New()
				hasher.Write([]byte(line))
				doc_id := hex.EncodeToString(hasher.Sum(nil))
				doc_id = doc_id
				doc := map[string]interface{}{}

				for index, name := range re.SubexpNames() {
					if index == 0 {
						doc["message"] = string(matches[0])
						continue
					}

					doc[name] = string(matches[index])
				}

				document, err := json.Marshal(doc)

				if err != nil {
					log.Println(err)
				}

				res, err := es.Index(*index_name, bytes.NewReader(document))

				if err != nil {
					panic("Indexing failed")
				} else {
					fmt.Println(res)
				}

				entries_indexed += 1
				break
			}

			if len(parser_directive.Regexes) == index+1 {
				errors += 1

				if _, err := error_file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
					log.Println(err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nIndexing done. Here are som vital stats:")
	fmt.Printf("Entries total: %d\n", line_counter)
	fmt.Printf("Entries indexed: %d\n", entries_indexed)
	fmt.Printf("Errors: %d\n", errors)

}

/*
 * Main application logic
 */

func main() {
	parser := argparse.NewParser("lapio", "Lapio - Log Shovel. Shovel logs into Elastic Search")
	parserDirectivePath := parser.String("d", "directive", &argparse.Options{
		Required: true,
		Help:     "Path to parser directive",
	})

	log_path := parser.String("l", "logpath", &argparse.Options{
		Required: true,
		Help:     "Path to log file",
	})

	index_name := parser.String("i", "index", &argparse.Options{
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

		parse_log(
			parserDirective,
			index_name,
			log_path,
		)
	}
}
