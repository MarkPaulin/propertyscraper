package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
)

var counter int = 0

type Property struct {
	Timestamp		string
	URL					string
	Address			string
	Price				string
	Style				string
	Bedrooms		string
	Receptions	string
	Rates				string
	Heating			string
	EPC					string
	Status			string
	Description	string
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
		colly.CacheDir("./propertypal_cache"),
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:		"propertypal.*",
		Delay:			  10*time.Second,
	})

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
	c.Wait()
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
			case "Rates":
				fmt.Printf("Rates: %s\n", el.ChildText("td"))
			case "Heating":
				fmt.Printf("Heating: %s\n", el.ChildText("td"))
			case "EPC Rating":
				fmt.Printf("EPC Rating: %s\n", space.ReplaceAllString(el.ChildText("td"), " "))
			case "Status":
				fmt.Printf("Status: %s\n", el.ChildText("td"))
			}
		})
	})

	p.OnHTML("p.enquiry-org", func(e *colly.HTMLElement) {
		fmt.Printf("Agent: %s\n", space.ReplaceAllString(e.Text, " "))
	})

	p.OnHTML("div.prop-descr-text", func(e *colly.HTMLElement) {
		description := space.ReplaceAllString(e.Text, " ")
		fmt.Printf("Description: %s\n", description)
	})

	url := fmt.Sprintf("https://www.propertypal.com%s", u)
	p.Visit(url)

	fmt.Printf("\n")
}
