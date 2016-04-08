package main

import (
	"flag"
	"fmt"
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

func fetch(comic string, time time.Time, done chan bool) {

	dFormat := dateFormat(comic, time)

	defer func() {
		done <- true
	}()

	switch comic {
	case "calvin":
		resp, err := http.Get(CALVIN + dFormat)
		if err != nil {
			return
		}

		parentFound := false

		defer resp.Body.Close()
		z := html.NewTokenizer(resp.Body)

		for {
			tokenType := z.Next()

			switch tokenType {
			case html.ErrorToken:
				return

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
									break
								}
							}
						}
					}
				}

			}
		}
	case "dilbert":
		resp, err := http.Get(DILBERT + dFormat)
		if err != nil {
			return
		}
		parentFound := false
		defer resp.Body.Close()

		html.NewTokenizer(resp.Body)
		z := html.NewTokenizer(resp.Body)

		for {
			tokenType := z.Next()

			switch tokenType {
			case html.ErrorToken:
				return

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
							break
						}
					}
				}

			}
		}
	default:
		fmt.Println("Not a valid comic for downloading: " + comic)
	}
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

func crawl(url string) error {
	return nil
}
