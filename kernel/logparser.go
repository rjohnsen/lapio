package kernel

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type ParserStatus struct {
	ErrorCount     int
	IndexedEntries int
	RowCount       int
}

func ParseLog(parserDirective ParserDirective, settings Settings, indexName string, logPath string) (ParserStatus, error) {
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

	retryBackoff := backoff.NewExponentialBackOff()

	// Elastic client
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			settings.Host,
		},
		Username: settings.Username,
		Password: settings.Password,

		// Retry on these status codes
		RetryOnStatus: []int{502, 503, 504, 429},

		// Backoff function
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Maximum number of retries
		MaxRetries: 5,
	})

	if err != nil {
		return parserStatus, errors.New("unable to create elastic client")
	}

	// BulkIndexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         indexName,
		Client:        es,
		NumWorkers:    settings.Workers,
		FlushBytes:    int(settings.Flushbytes),
		FlushInterval: time.Duration(settings.Flushinterval) * time.Second,
	})

	if err != nil {
		return parserStatus, errors.New("Unable to create BulkIndexer")
	}

	for scanner.Scan() {
		line := scanner.Text()
		parserStatus.RowCount += 1
		regexMatched := false

		for _, regx := range parserDirective.Regexes {
			re := regexp.MustCompile(regx.Expression)
			matches := re.FindStringSubmatch(line)

			if len(matches) == regx.CaptureGroups+1 {
				doc := map[string]interface{}{}

				for index, name := range re.SubexpNames() {
					if index == 0 {
						doc["message"] = string(matches[0])
						doc["log_origin"] = logPath
						continue
					}

					// Handling log time
					if name == parserDirective.Time.MappingField {
						timeStr := matches[index]
						parsedTime, err := time.Parse(parserDirective.Time.Layout, timeStr)

						if err != nil {
							fmt.Println("Error parsing time:", err)
						} else {
							rfc3339Time := parsedTime.Format(time.RFC3339)
							doc["@timestamp"] = rfc3339Time
						}
					}

					doc[name] = string(matches[index])
				}

				document, err := json.Marshal(doc)

				if err != nil {
					log.Println(err)
				}

				// Calulate document ID
				hasher := md5.New()
				hasher.Write([]byte(line))
				docId := hex.EncodeToString(hasher.Sum(nil))

				err = bi.Add(
					context.Background(),
					esutil.BulkIndexerItem{
						Action:     "index",
						DocumentID: docId,
						Body:       bytes.NewReader(document),
						OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
							parserStatus.IndexedEntries += 1
						},
						OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
							if err != nil {
								log.Printf("ERROR: %s", err)
								parserStatus.ErrorCount += 1
							} else {
								log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
								parserStatus.ErrorCount += 1
							}
						},
					},
				)

				if err != nil {
					log.Fatalf("Unexpected error: %s", err)
				} else {
					fmt.Printf("[ _id: %s ] ItemNum: %d - Errors: %d\n", docId, parserStatus.IndexedEntries, parserStatus.ErrorCount)
					regexMatched = true
					break
				}
			}
		}

		if !regexMatched {
			if _, err := errorFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
				log.Println(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return parserStatus, nil
}
