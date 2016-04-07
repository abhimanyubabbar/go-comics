package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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

	for _, comic := range comics {

		dFormat := dateFormat(comic, now)
		switch comic {
		case "calvin":
			resp, err := http.Get(CALVIN + dFormat)
			if err != nil {
				panic(err)
			}
			bytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(bytes))
		case "dilbert":
			resp, err := http.Get(DILBERT + dFormat)
			if err != nil {
				panic(err)
			}
			bytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(bytes))

		}
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
