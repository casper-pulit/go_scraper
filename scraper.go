package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

type item struct {
	Quote        string   `json:"quote"`
	Author       string   `json:"author"`
	Book         string   `json:"book"`
	Book_Rel_URL string   `json:"book_rel_url"`
	Likes        int      `json:"likes"`
	Tags         []string `json:"tags"`
	Index        int      `json:"index"`
}

var (
	book_rel_url string
	tags         []string
	url          string = "https://www.goodreads.com/quotes?ref=nav_comm_quotes"
	quote        string
	author       string
	book         string
	like         int
	counter      int
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

//func getBookUrl(book_href []string) string {
//book_collector := colly.NewCollector()
//book_collector.
//Request.AbsoluteURL(strings.Join(book_href, ""))

//book_quote_url := book

//c.OnHTML("[class=next_page]", func(h *colly.HTMLElement) {
//	next_page := h.Request.AbsoluteURL(h.Attr("href"))
//	c.Visit(next_page)
//})
//re_select := regexp.MustCompile("[^0-9]+")
//book_id_str := re_select.ReplaceAllString(strings.Join(slice_str, ""), "")
//book_url := "https://www.goodreads.com/book/show/" + book_id_str

//return (book_url)
//}

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

func scrape(limit int, search_tag string, filename string) {

	// init slice of item structs
	items := []item{}
	// Create a new Colly collector
	c := colly.NewCollector()

	// Define the URL you want to scrape
	if search_tag != "" {
		url = "https://www.goodreads.com/quotes/tag?utf8=%E2%9C%93&id=" + search_tag
	}

	c.OnHTML("div.quoteDetails", func(h *colly.HTMLElement) {
		quote = cleanQuote(h.ChildText("div.quoteText"))

		book_href := h.ChildAttrs("[class='authorOrTitle']", "href")
		if len(book_href) > 0 {
			book_rel_url = strings.Join(book_href, "")
			/*
				Rework this chunk to get actual book URL to be async or move to separate script to be ran post main data scraping possibly on demand once script set up as a web app
				book_quote_url := h.Request.AbsoluteURL(strings.Join(book_href, ""))
				book_collector := colly.NewCollector()
				book_collector.OnHTML("[class = 'leftContainer']", func(book_h *colly.HTMLElement) {
					//fmt.Println(book_h.ChildAttr("[class = 'bookTitle']", "href"))
					book_url := book_h.Request.AbsoluteURL(book_h.ChildAttr("[class = 'bookTitle']", "href"))
					fmt.Println(book_url)

				})

				book_collector.Visit(book_quote_url)
			*/
		} else {
			book_rel_url = ""
		}

		author, book = cleanAuthor(h.ChildText("span"))
		like = cleanLikes(h.ChildText("[class=smallText]"))
		tags = h.ChildTexts("[class='greyText smallText left'] > a")

		i := item{
			Quote:        quote,
			Author:       author,
			Book:         book,
			Book_Rel_URL: book_rel_url,
			Likes:        like,
			Tags:         tags,
			Index:        counter,
		}
		// append quote struct to items slice
		items = append(items, i)

		// increment counter
		counter++

		writeScrape(items, filename, counter, limit)
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
		writeScrape(items, filename, counter, 0)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println(r.URL.String())
	})

	c.Visit(url)

}
func main() {
	quote_limit := flag.Int("limit", 100, "Set number of quotes to return if available")
	quote_tag := flag.String("tag", "", "Search for a specific tag. If empty will scrape most popular quotes.")
	filename := flag.String("out", "output", "Define name for output file")
	flag.Parse()
	scrape(*quote_limit, *quote_tag, *filename)
}
