package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

// http://www.gocomics.com/calvinandhobbes/2016/04/06
// http://dilbert.com/strip/2016-04-07
// http://xkcd.com/info.0.json

const (
	CALVIN  = "http://www.gocomics.com/calvinandhobbes/"
	DILBERT = "http://dilbert.com/strip/"
	XKCD    = "http://xkcd.com/info.0.json"
)

type comicSlice []string

func (cs *comicSlice) String() string {
	return fmt.Sprintf("%s", *cs)
}

func (cs *comicSlice) Set(val string) error {
	*cs = append(*cs, val)
	return nil
}

var directory string
var comics comicSlice

func init() {
	flag.StringVar(&directory, "d", "home/babbar/Pictures/Comics", "default directory for downloading comics")
	flag.Var(&comics, "c", "calvin, dilbert")
}

func main() {

	fmt.Println("Parsing the flags")
	flag.Parse()

	fmt.Println("Started looking for your comics.")
	now := time.Now()

	done := make(chan bool)
	for _, comic := range comics {
		go fetch(comic, directory, now, done)
	}

	fmt.Println("Synchronizing the responses.")

	// Wait for as many responses as the comics.
	for range comics {
		<-done
	}
}

func fetch(comic, baseDir string, time time.Time, done chan bool) error {

	dFormat := dateFormat(comic, time)

	defer func() {
		fmt.Println("Replying the channel with true")
		done <- true
	}()

	switch comic {
	case "calvin":
		loc := baseDir + "/calvin-" + time.String()
		path, err := crawl(CALVIN+dFormat, calvinDocumentProcessor)
		if err != nil {
			fmt.Println("Unable to crawl calvin document" + err.Error())
			return errors.New("Unable to crawl for calvin document" + err.Error())
		}
		return downloadDocument(*path, loc)
	case "dilbert":
		loc := baseDir + "/dilbert-" + time.String()
		path, err := crawl(DILBERT+dFormat, dilbertDocumentProcessor)
		if err != nil {
			return errors.New("Unable to crawl for dilbert document" + err.Error())
		}
		return downloadDocument(*path, loc)
	case "xkcd":
		loc := baseDir + "/xkcd" + time.String()
		path, err := crawl(XKCD, xkcdDocumentProcessor)
		if err != nil {
			return errors.New("Unable to crawl for xkcd document" + err.Error())
		}
		return downloadDocument(*path, loc)

	default:
		fmt.Println("Not a valid comic for downloading: " + comic)
	}
	return nil
}

func dateFormat(comic string, time time.Time) string {

	format := ""
	year, month, date := time.Date()
	switch comic {
	case "calvin":
		format = strconv.Itoa(year) + "/" +
			strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(date)
	case "dilbert":
		format = strconv.Itoa(year) + "-" +
			strconv.Itoa(int(month)) + "-" +
			strconv.Itoa(date)
	default:
		format = strconv.Itoa(year) + "/" +
			strconv.Itoa(int(month)) + "/" +
			strconv.Itoa(date)
	}

	return format
}

type DocumentProcessor func(reader io.ReadCloser) (*string, error)

func crawl(url string, processor DocumentProcessor) (*string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Unable to fetch the webpage at url" + url)
	}

	defer resp.Body.Close()
	path, err := processor(resp.Body)

	if err != nil {
		return nil, errors.New("Unable to process the response body" + err.Error())
	}

	return path, nil
}

func xkcdDocumentProcessor(body io.ReadCloser) (*string, error) {

	var xkcd Xkcd
	decoder := json.NewDecoder(body)

	fmt.Println("Started crawling for xkcd document")

	if err := decoder.Decode(&xkcd); err != nil {
		return nil, errors.New("Unable to extract the xkcd details: " + err.Error())
	}

	fmt.Printf("%#v", xkcd)
	return &xkcd.Image, nil
}

func calvinDocumentProcessor(body io.ReadCloser) (*string, error) {

	parentFound := false

	z := html.NewTokenizer(body)
	for {
		tokenType := z.Next()
		switch tokenType {

		case html.StartTagToken:
			token := z.Token()
			if token.Data == "div" {
				for _, attr := range token.Attr {
					if attr.Key == "class" &&
						attr.Val == "feature" {
						// found the div for the comic.
						fmt.Println("Found the parent container div > Calvin")
						parentFound = true
					}
				}
			}

		case html.SelfClosingTagToken:
			token := z.Token()
			// Locate the image under the parent div
			if token.Data == "img" && parentFound {
				for _, attr := range token.Attr {

					// attributes key and values.
					if attr.Key == "alt" &&
						attr.Val == "Calvin and Hobbes" {

						fmt.Println("Found the img token > Calvin")
						// again iterate to check for the
						// values
						for _, attributes := range token.Attr {
							if attributes.Key == "src" {
								fmt.Println(attributes.Val)
								return &attributes.Val, nil
							}
						}
					}
				}
			}
		}
	}

	return nil, errors.New("Unable to locate the calvin comic strip")
}

func dilbertDocumentProcessor(body io.ReadCloser) (*string, error) {

	parentFound := false
	z := html.NewTokenizer(body)

	for {

		tokenType := z.Next()
		switch tokenType {

		case html.StartTagToken:
			token := z.Token()
			if token.Data == "div" {

				for _, attr := range token.Attr {
					if attr.Key == "class" &&
						attr.Val == "img-comic-container" {
						// found the div for the comic.
						fmt.Println("Found the parent container div > Dilbert")
						parentFound = true
						break
					}
				}
			}

		case html.SelfClosingTagToken:

			token := z.Token()
			// Locate the image under the parent div
			if token.Data == "img" && parentFound {
				fmt.Println("Found the image tag > Dilbert")
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						fmt.Println(attr.Val)
						return &attr.Val, nil
					}
				}
			}

		}
	}

	return nil, errors.New("Unable to locate the dilbert comic strip")
}

// downloadDocument: Download the document at the specified location.
func downloadDocument(url, loc string) error {

	fmt.Println("Going to download document, located at url: " + url + ", at location: " + loc)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Someting is wrong")
		return errors.New("Unable to download the document " + err.Error())
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Someting is wrong in reading contents")
		return errors.New("Unable to read the contents of strip")
	}

	format := getFormat(contents)

	if format == "" {
		fmt.Println("Someting is wrong in detecting image format")
		return errors.New("Unable to detect the correct format of document.")
	}

	ext := "." + (format)
	fmt.Println("extension: " + ext)

	if err := ioutil.WriteFile(loc+ext, contents, 0777); err != nil {
		fmt.Println("Someting is wrong while writing file")
		return errors.New("Unable to write the contents to the file." + err.Error())
	}

	return nil
}
