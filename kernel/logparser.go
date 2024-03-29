package kernel

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/elastic/go-elasticsearch/v8"
)

type ParserStatus struct {
	ErrorCount     int
	IndexedEntries int
	RowCount       int
}

func ParseLog(parserDirective ParserDirective, elasticCredentials Settings, indexName string, logPath string) (ParserStatus, error) {
	var parserStatus ParserStatus
	logFile, err := os.Open(*&logPath)

	if err != nil {
		return parserStatus, errors.New("unable to load log file")
	}

	defer logFile.Close()

	scanner := bufio.NewScanner(logFile)

	errorFile, err := os.OpenFile("error.data", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return parserStatus, errors.New("unable to create error file")
	}

	defer errorFile.Close()

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			elasticCredentials.Host,
		},
		Username: elasticCredentials.Username,
		Password: elasticCredentials.Password,
	})

	if err != nil {
		return parserStatus, errors.New("unable to create elastic client")
	}

	for scanner.Scan() {
		line := scanner.Text()
		parserStatus.RowCount += 1

		for index, regx := range parserDirective.Regexes {
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

				res, err := es.Index(*&indexName, bytes.NewReader(document))

				if err != nil {
					return parserStatus, errors.New("unable to index row. Please check if ElasticSearch is running and port 9200 is open.")
				} else {
					fmt.Println(res)
				}

				parserStatus.IndexedEntries += 1
				break
			}

			if len(parserDirective.Regexes) == index+1 {
				parserStatus.ErrorCount += 1

				if _, err := errorFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
					log.Println(err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return parserStatus, nil
}
