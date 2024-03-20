package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/gocolly/colly/v2"
)

type item struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
	//	Tag    string "json:tag"
}

var (
	quote   string
	counter int
	limit   int = 1000
)

func main() {
	// init slice of item structs
	items := []item{}
	// Create a new Colly collector
	c := colly.NewCollector()

	// Define the URL you want to scrape
	url := "https://www.goodreads.com/quotes"

	c.OnHTML("div.quoteDetails", func(h *colly.HTMLElement) {
		// regex to select quote from inbetween quote marks
		re_select := regexp.MustCompile(`^“.*?”`)
		// regex to remove quote marks from quote
		re_remove := regexp.MustCompile("[“”]")
		// apply regexes
		quote = re_select.FindString(h.ChildText("div.quoteText"))
		quote = re_remove.ReplaceAllLiteralString(quote, "")
		i := item{
			Quote:  quote,
			Author: h.ChildText("span"),
		}
		// append quote struct to items slice
		items = append(items, i)

		counter += 1

		if counter >= limit {
			data, err := json.Marshal(items)
			if err != nil {
				log.Fatal(err)
			}
			os.WriteFile("quotes.json", data, 0644)
			os.Exit(0)
		}

	})
	// Visit next page
	c.OnHTML("[class=next_page]", func(h *colly.HTMLElement) {
		next_page := h.Request.AbsoluteURL(h.Attr("href"))
		fmt.Println(next_page)
		c.Visit(next_page)
	})

	c.Visit(url)

}
