package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
	)

	c.OnHTML("div.propbox", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		if strings.Index(link, "/user/") != -1 {
			link = e.ChildAttr("a[href]:nth-of-type(2)", "href")
		}

		fmt.Printf("Link found: %s\n", link)
		scrapeProperty(link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://www.propertypal.com/search?sta=forSale&sta=saleAgreed&sta=sold&st=sale&currency=GBP&term=15&pt=residential")
}

func scrapeProperty(u string) {
	p := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
	)

	p.OnHTML("h1", func(e *colly.HTMLElement) {
		address := e.Text
		fmt.Printf("Address: %s\n", address)
	})

	url := fmt.Sprintf("https://www.propertypal.com%s", u)
	p.Visit(url)
}
