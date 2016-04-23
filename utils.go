package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

type Xkcd struct {
	Title string `json:"title"`
	Image string `json:"img"`
}

type ComicDownload struct {
	url  string `json:"url"`
	path string `json:"path"`
}

func getFormat(contents []byte) string {
	if len(contents) < 4 {

		fmt.Println("Not enough bytes")
		return ""
	}

	if contents[0] == 0x89 && contents[1] == 0x50 &&
		contents[2] == 0x4E && contents[3] == 0x47 {
		return "png"
	}

	if contents[0] == 0xFF && contents[1] == 0xD8 {
		return "jpg"
	}

	if contents[0] == 0x47 && contents[1] == 0x49 &&
		contents[2] == 0x46 && contents[3] == 0x38 {
		return "gif"
	}

	if contents[0] == 0x42 && contents[1] == 0x4D {
		return "bmp"
	}

	return ""
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
