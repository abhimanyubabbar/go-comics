package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

// http://www.gocomics.com/calvinandhobbes/2016/04/06
// http://dilbert.com/strip/2016-04-07

const (
	CALVIN  = "http://www.gocomics.com/calvinandhobbes/"
	DILBERT = "http://dilbert.com/strip/"
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
	flag.StringVar(&directory, "d", "~/Pictures/Comics", "default directory for downloading comics")
	flag.Var(&comics, "c", "calvin, dilbert")
}

func main() {

	fmt.Println("Parsing the flags")
	flag.Parse()

	fmt.Println("Started looking for your comics.")
	now := time.Now()

	done := make(chan bool)
	for _, comic := range comics {
		go fetch(comic, now, done)
	}

	fmt.Println("Synchronizing the responses.")

	// Wait for as many responses as the comics.
	for range comics {
		<-done
	}
}

func fetch(comic string, time time.Time, done chan bool) error {

	dFormat := dateFormat(comic, time)

	defer func() {
		fmt.Println("Replying the channel with true")
		done <- true
	}()

	switch comic {
	case "calvin":
		return crawl(CALVIN+dFormat, calvinDocumentProcessor)
	case "dilbert":
		return crawl(DILBERT+dFormat, dilbertDocumentProcessor)
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

type DocumentProcessor func(reader io.ReadCloser) error

func crawl(url string, processor DocumentProcessor) error {

	resp, err := http.Get(url)
	if err != nil {
		return errors.New("Unable to fetch the webpage at url" + url)
	}

	defer resp.Body.Close()
	if err := processor(resp.Body); err != nil {
		return errors.New("Unable to process the response body" + err.Error())
	}

	return nil
}

func calvinDocumentProcessor(body io.ReadCloser) error {

	articleFound := false
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
								articleFound = true
								break
							}
						}
					}
				}

				break
			}
		}
	}

	if !articleFound {
		return errors.New("Unable to locate the calvin strip.")
	}

	fmt.Println("Hunting for calvin strips finished.")
	return nil
}

func dilbertDocumentProcessor(body io.ReadCloser) error {

	parentFound := false
	articleFound := false

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
						articleFound = true
						break
					}
				}
				// Found what we are looking for.
				fmt.Println("Going to break out of main loop > Dilbert")
				break
			}

		}
	}

	if !articleFound {
		return errors.New("Unable to locate the dilbert comic strip")
	}

	return nil
}
