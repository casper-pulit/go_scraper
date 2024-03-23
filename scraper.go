package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type item struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
	Book   string `json:"book"`
	Likes  int    `json:"likes"`
}

var (
	quoteLimit int = 100
	tag        string
	url        string = "https://www.goodreads.com/quotes?ref=nav_comm_quotes"
	quote      string
	author     string
	book       string
	like       int
	counter    int
)

func cleanQuote(text string) string {
	// regex to select quote from inbetween quote marks
	re_select := regexp.MustCompile(`^“.*?”`)
	// regex to remove quote marks from quote
	re_remove := regexp.MustCompile("[“”]")
	// apply regexes
	select_text := re_select.FindString(text)
	clean_text := re_remove.ReplaceAllLiteralString(select_text, "")

	return clean_text
}

func cleanAuthor(text string) (string, string) {
	// check for delimiter used when both author and the book are listed
	if strings.Contains(text, "\n  \n") {

		// split strings into slice based on delimiter
		split_text := strings.Split(text, "\n  \n")
		// author == first element of slice
		author := split_text[0]
		// trim white space
		author = strings.TrimSpace(author)

		// remove comma from end of authors name
		re_remove := regexp.MustCompile(",$")
		author = re_remove.ReplaceAllLiteralString(author, "")

		// book == second element of slice
		book := split_text[1]
		// trim white space
		book = strings.TrimSpace(book)

		return author, book
	} else {
		// if no delimiter book is empty string.
		return text, ""
	}

}

func cleanLikes(text string) int {
	re_select := regexp.MustCompile("[^0-9]+")
	num_likes_str := re_select.ReplaceAllString(text, "")
	num_likes_int, err := strconv.Atoi(num_likes_str)
	if err != nil {
		log.Fatal(err)
	}

	return (num_likes_int)

}

func writeScrape(items []item, filename string, counter int, limit int) {
	// if counter exceeds limit output
	if counter >= limit {
		data, err := json.Marshal(items)
		if err != nil {
			log.Fatal(err)
		}
		// create filename with .json extension
		filename = fmt.Sprintf("%s.json", filename)
		// write file
		os.WriteFile(filename, data, 0644)
		// quit program
		os.Exit(0)
	}
}

func scrape(limit int) {

	// init slice of item structs
	items := []item{}
	// Create a new Colly collector
	c := colly.NewCollector()

	// Define the URL you want to scrape

	if tag != "" {
		url = "https://www.goodreads.com/quotes/tag?utf8=%E2%9C%93&id=" + tag
	}

	c.OnHTML("div.quoteDetails", func(h *colly.HTMLElement) {
		quote = cleanQuote(h.ChildText("div.quoteText"))
		author, book = cleanAuthor(h.ChildText("span"))
		like = cleanLikes(h.ChildText("[class=smallText]"))
		// Tag scrape WIP
		//fmt.Println(h.ChildText("[class='greyText smallText left']"))

		i := item{
			Quote:  quote,
			Author: author,
			Book:   book,
			Likes:  like,
			//Tags:   tags,
		}
		// append quote struct to items slice
		items = append(items, i)

		// increment counter
		counter++

		writeScrape(items, "output", counter, limit)
	})

	// Visit next page
	c.OnHTML("[class=next_page]", func(h *colly.HTMLElement) {
		next_page := h.Request.AbsoluteURL(h.Attr("href"))
		c.Visit(next_page)
	})
	// if reaching the end
	c.OnHTML("[class='next_page disabled']", func(h *colly.HTMLElement) {
		fmt.Println("Scraper reached the end... outputting results.")
		// set limit to 0 so that counter is always greater than limit when finished scraping.
		writeScrape(items, "output", counter, 0)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	c.Visit(url)

}
func main() {
	scrape(quoteLimit)
}
