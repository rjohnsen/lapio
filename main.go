package main

import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "os"
	"github.com/akamensky/argparse"
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
* Main application logic
*/

func main() {
	parser := argparse.NewParser("lapio", "Lapio - Log Shovel. Shovel logs into Elastic Search")
	parser_directive_path := parser.String("d", "directive", &argparse.Options{
		Required: true, 
		Help: "Path to parser directive",
	})

	log_file_path := parser.String("l", "logpath", &argparse.Options{
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

		fmt.Printf("Index: %s", index_name)
		fmt.Printf("Parser Directive: %s\n", parser_directive.Name)
		fmt.Printf("Log file: %s\n", *log_file_path)
	}
}