package kernel

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
)

type TimeMapping struct {
	XMLName      xml.Name `xml:"timemapping"`
	MappingField string   `xml:"field,attr"`
	Layout       string   `xml:",chardata"`
}

type RegexField struct {
	XMLName       xml.Name `xml:"regex"`
	Expression    string   `xml:",chardata"`
	CaptureGroups int      `xml:"capturegroups,attr"`
}

type Logfield struct {
	XMLName  xml.Name `xml:"logfield"`
	Name     string   `xml:"name,attr"`
	Datatype string   `xml:"datatype,attr"`
}

type ParserDirective struct {
	XMLName     xml.Name     `xml:"parserdirective"`
	Name        string       `xml:"name"`
	Description string       `xml:"description"`
	Regexes     []RegexField `xml:"regexes>regex"`
	Time        TimeMapping
	Logfields   []Logfield `xml:"logfields>logfield"`
}

func LoadParserDirective(xml_path string) (ParserDirective, error) {
	var parserDirective ParserDirective

	xmlFile, err := os.Open(xml_path)

	if err != nil {
		return parserDirective, errors.New("unable to load parser directive")
	}

	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)

	if err := xml.Unmarshal(byteValue, &parserDirective); err != nil {
		return parserDirective, errors.New("unable to parse parser directive")
	}

	return parserDirective, nil
}
