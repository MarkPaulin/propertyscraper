package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

var counter int = 0

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
		colly.CacheDir("./propertypal_cache"),
	)

	c.OnHTML("div.propbox", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		if strings.Index(link, "/user/") != -1 {
			link = e.ChildAttr("a[href]:nth-of-type(2)", "href")
		}

		if strings.Index(link, "/premium") != -1 {
			return
		}

		fmt.Printf("Link found: %s\n", link)
		scrapeProperty(link)
	})

	c.OnHTML("a.paging-next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if counter > 2 {
			fmt.Printf("Reached %d pages", counter)
			return
		}
		counter++
		e.Request.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://www.propertypal.com/search?sta=forSale&sta=saleAgreed&sta=sold&st=sale&currency=GBP&term=15&pt=residential")
}

func scrapeProperty(u string) {
	space := regexp.MustCompile(`\n+`)
	p := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
	)

	p.OnHTML("h1", func(e *colly.HTMLElement) {
		address := space.ReplaceAllString(e.Text, " ")
		fmt.Printf("Address: %s\n", address)
	})

	p.OnHTML("#key-info-table", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			switch el.ChildText("th") {
			case "Price":
				fmt.Printf("Price: %s\n", el.ChildText("td"))
			case "Style":
				fmt.Printf("Style: %s\n", el.ChildText("td"))
			case "Bedrooms":
				fmt.Printf("Bedrooms: %s\n", el.ChildText("td"))
			case "Receptions":
				fmt.Printf("Receptions: %s\n", el.ChildText("td"))
			}
		})
	})

	url := fmt.Sprintf("https://www.propertypal.com%s", u)
	p.Visit(url)
	fmt.Printf("\n")
}
