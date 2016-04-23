package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
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

func mainUpdated() {

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

func main() {

	fmt.Println("Parsing the flags")
	flag.Parse()

	fmt.Println("Started looking for your comics.")
	now := time.Now()

	done := make(chan bool)

	// publisher channel over which
	// we publish the comics.
	comicChan := publisherRoutine()

	// start the fetch routine over which they
	// extract the information for the document.
	downloadChan := fetchRoutine(comicChan, directory, now)

	// We would like to split the
	// work between two downloaders.
	downloadWorker := 2
	for i := 0; i < downloadWorker; i++ {
		downloadRoutine(downloadChan, done)
	}

	// FIXME: Fix this step to a more elegant solution.
	// sync step to let the downloaders finish
	for i := 0; i < downloadWorker; i++ {
		<-done
	}

}

func publisherRoutine() chan string {
	comicChan := make(chan string)

	go func() {
		for _, comic := range comics {
			comicChan <- comic
		}
		close(comicChan)
	}()

	return comicChan
}

// fetchRoutine based on the comics generates the
// FIXME : How to handle the errors in this routine elegantly.
// If not handled properly, we may end up in deadlock
func fetchRoutine(comicChan <-chan string, baseDir string, time time.Time) chan ComicDownload {

	downloadChan := make(chan ComicDownload, 1)

	go func() {

		for comic := range comicChan {

			dFormat := dateFormat(comic, time)
			var loc, url string
			var processor DocumentProcessor

			switch comic {
			case "calvin":
				loc = baseDir + "/calvin-" + time.String()
				url = CALVIN + dFormat
				processor = calvinDocumentProcessor
			case "dilbert":
				loc = baseDir + "/dilbert-" + time.String()
				url = DILBERT + dFormat
				processor = dilbertDocumentProcessor
			case "xkcd":
				loc = baseDir + "/xkcd" + time.String()
				url = XKCD
				processor = xkcdDocumentProcessor
			default:
				fmt.Println("Not a valid comic for downloading: " + comic)
				continue
			}

			downloadLink, err := crawl(url, processor)
			if err != nil {
				fmt.Println("Unable to crawl calvin document" + err.Error())
				continue
			}

			comicDownload := ComicDownload{
				url:  *downloadLink,
				path: loc,
			}
			// Once we have successfully crawled a document, send the link generated
			// from the document.
			downloadChan <- comicDownload
		}

		close(downloadChan)
	}()

	return downloadChan
}

func fetch(comic, baseDir string, time time.Time, done chan bool) error {

	dFormat := dateFormat(comic, time)

	defer func() {
		fmt.Println("Replying the channel with true")
		done <- true
	}()

	var loc, url string
	var processor DocumentProcessor

	switch comic {
	case "calvin":
		loc = baseDir + "/calvin-" + time.String()
		url = CALVIN + dFormat
		processor = calvinDocumentProcessor
	case "dilbert":
		loc = baseDir + "/dilbert-" + time.String()
		url = DILBERT + dFormat
		processor = dilbertDocumentProcessor
	case "xkcd":
		loc = baseDir + "/xkcd" + time.String()
		url = XKCD
		processor = xkcdDocumentProcessor
	default:
		return errors.New("Not a valid comic for downloading: " + comic)
	}

	path, err := crawl(url, processor)
	if err != nil {
		fmt.Println("Unable to crawl calvin document" + err.Error())
		return errors.New("Unable to crawl for calvin document" + err.Error())
	}
	return downloadDocument(*path, loc)
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

func downloadRoutine(downloadChan <-chan ComicDownload, done chan bool) {

	go func() {

		defer func() {
			done <- true
		}()

		for obj := range downloadChan {

			fmt.Println("Going to download document, located at url: " +
				obj.url + ", at location: " + obj.path)

			resp, err := http.Get(obj.url)
			if err != nil {
				fmt.Println("Unable to download the document " + err.Error())
				continue
			}

			defer resp.Body.Close()

			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Unable to read the contents of strip")
			}

			format := getFormat(contents)

			if format == "" {
				fmt.Println("Unable to detect the correct format of document.")
				continue
			}

			ext := "." + (format)
			fmt.Println("extension: " + ext)

			if err := ioutil.WriteFile(obj.path+ext, contents, 0777); err != nil {
				fmt.Println("Unable to write the contents to the file." + err.Error())
				continue
			}
		}
	}()
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
