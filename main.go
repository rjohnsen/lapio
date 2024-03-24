package main

import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "os"
	"github.com/akamensky/argparse"
	"bufio"
	"log"
	"regexp"
)

/*
 * Directive section
 */

type Logfield struct {
	XMLName		xml.Name	`xml:"logfield"`
	Name		string		`xml:"name,attr"`
	Datatype	string		`xml:"datatype,attr"`
}

type ParserDirective struct {
	XMLName		xml.Name	`xml:"parserdirective"`
	Name		string		`xml:"name"`
	Description	string		`xml:"description"`
	Regex		string		`xml:"regex"`
	Logfields	[]Logfield	`xml:"logfields>logfield"`
}

func load_parser_directive(xml_path string) ParserDirective {
	xmlFile, err := os.Open(xml_path)

	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var parser_directive ParserDirective

	if err :=xml.Unmarshal(byteValue, &parser_directive); err != nil {
        panic(err)
    }

	return parser_directive
}

/*
 * Parse log
 */

func parse_log(parser_directive ParserDirective, index_name *string, log_path *string) {
	fmt.Printf("Index: %s\n", *index_name)
	fmt.Printf("Parser Directive: %s\n", parser_directive.Name)
	fmt.Printf("Log file: %s\n", *log_path)

	file, err := os.Open(*log_path)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	line_counter := 0

	// Tracking errors
	errors := 0
	entries_indexed := 0

	// Write lines with error to file for follow up
	error_file, err := os.OpenFile("error.data", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
	}
	
	defer error_file.Close()

	for scanner.Scan() {
		line := scanner.Text()
		line_counter += 1

		re := regexp.MustCompile(parser_directive.Regex)
		matches := re.FindStringSubmatch(line)

		if len(matches) >= 8 {
			fmt.Printf("%d - %d - %s => %s\n", line_counter, len(matches), matches[re.SubexpIndex("http_method")], matches[re.SubexpIndex("url")])
			entries_indexed += 1
		} else {
			errors += 1
			
			if _, err := error_file.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
				log.Println(err)
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
	parser_directive_path := parser.String("d", "directive", &argparse.Options{
		Required: true, 
		Help: "Path to parser directive",
	})

	log_path := parser.String("l", "logpath", &argparse.Options{
		Required: true,
		Help: "Path to log file",
	})

	index_name := parser.String("i", "index", &argparse.Options{
		Required: true,
		Help: "Name of index (where to store logs)",
	})

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
	} else {
		var parser_directive = load_parser_directive(*parser_directive_path)
		parse_log(
			parser_directive,
			index_name,
			log_path,
		)
	}
}