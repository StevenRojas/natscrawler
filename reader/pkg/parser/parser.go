package parser

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
)

const urlColNum = 0

// CSVParser interface
type CSVParser interface {
	Parse(ctx context.Context, wg*sync.WaitGroup, urlCh chan<- string)
}

type csvParser struct {
	file string
	f *os.File
	skipRows int
	urlDomain string
}

// NewCSVParser validate that the CSV file exists and returns an instance of CSVParser
func NewCSVParser(file string, skipRows int, urlDomain string) (CSVParser, error) {
	f, err := os.Open(file)
	if err != nil {
		log.Printf("unable to open file handler: %s\n", err.Error())
		return nil, err
	}
	return &csvParser{
		file: file,
		f: f,
		skipRows: skipRows,
		urlDomain: urlDomain,
	}, nil
}

// Parse a CSV file sending each row though the channel
func (c *csvParser) Parse(ctx context.Context, wg*sync.WaitGroup, urlCh chan<- string) {
	defer wg.Done()
	log.Println("parsing CSV file")
	reader := csv.NewReader(c.f)
	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		row, err := reader.Read()
		if err == io.EOF {
			log.Println("CSV end of file")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		count++
		if count <= c.skipRows {
			continue
		}
		uri := row[urlColNum]
		if c.validate(uri) {
			urlCh <- uri
		} else {
			log.Printf("Invalid URL: %s\n", uri)
		}

	}
	close(urlCh)
	err := c.f.Close()
	if err != nil {
		log.Fatalf("unable to close file handler: %s\n", err.Error())
	}
}

// validate if a URL is formatted correctly and is for the defined URL domain
func (c *csvParser) validate(uri string) bool {
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}
	if !strings.Contains(uri, c.urlDomain) {
		return false
	}
	return true
}