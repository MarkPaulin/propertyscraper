package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	_ "github.com/mattn/go-sqlite3"
)

func removeSpaces(s string) string {
	space := regexp.MustCompile(`\n*\s+`)
	comma := regexp.MustCompile(`\s+,`)

	result := space.ReplaceAllString(s, " ")
	result = comma.ReplaceAllString(result, ",")
	return result
}

func getID(u string) int {
	re := regexp.MustCompile(`\d+$`)
	s := re.FindString(u)

	id, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln(err)
	}

	return id
}

const query string = "REPLACE INTO properties (url, id, updated_time, lat, lon, " +
	"address, postcode, price_offers, price, price_min, price_max, agent, " +
	"branch, style, bedrooms, receptions, bathrooms, rates, heating, epc, " +
	"status, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

var counter = 0

type Property struct {
	URL         string
	ID          int
	UpdatedTime string
	Lat         float64
	Lon         float64
	Address     string
	Postcode    string
	PriceOffers string
	Price       string
	PriceMin    string
	PriceMax    string
	Agent       string
	Branch      string
	Style       string
	Bedrooms    int
	Receptions  int
	Bathrooms   int
	Rates       string
	Heating     string
	EPC         string
	Status      string
	Description string
}

func main() {
	properties := make([]Property, 0, 1000)
	c := colly.NewCollector(
		colly.AllowedDomains("www.propertypal.com"),
		colly.CacheDir("./propertypal_cache"),
	)

	c2 := c.Clone()

	c.OnHTML("div.propbox", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		if strings.Index(link, "/user/") != -1 {
			link = e.ChildAttr("a[href]:nth-of-type(2)", "href")
		}

		if strings.Index(link, "/premium") != -1 {
			return
		}

		url := e.Request.AbsoluteURL(link)
		c2.Visit(url)
	})

	c2.OnHTML("html", func(e *colly.HTMLElement) {
		db, err := sql.Open("sqlite3", "properties-db.sqlite")
		if err != nil {
			log.Fatalln("unable to open database", err)
		}
		defer db.Close()

		statement, err := db.Prepare(query)
		if err != nil {
			log.Fatalln("unable to prepare query", err)
		}
		defer statement.Close()

		property := Property{}

		property.URL = e.Request.URL.String()

		property.ID = getID(e.Request.URL.String())

		e.ForEach("meta", func(_ int, el *colly.HTMLElement) {
			switch el.Attr("property") {
			case "og:updated_time":
				property.UpdatedTime = el.Attr("content")
			case "place:location:latitude":
				lat, err := strconv.ParseFloat(el.Attr("content"), 64)
				if err != nil {
					log.Fatalln(err)
				}
				property.Lat = lat
			case "place:location:longitude":
				lon, err := strconv.ParseFloat(el.Attr("content"), 64)
				if err != nil {
					log.Fatalln(err)
				}
				property.Lon = lon
			case "og:title":
				property.Address = el.Attr("content")
			}
		})

		property.Postcode = e.ChildText("span.prop-summary-townPostcode > span.text-ib")

		property.PriceOffers = e.ChildText("div.prop-price-sm > span.price > span.price-offers")
		property.Price = e.ChildText("div.prop-price-sm > span.price > span.price-value")
		property.PriceMin = e.ChildText("div.prop-price-sm > span.price > span.price-min")
		property.PriceMax = e.ChildText("div.prop-price-sm > span.price > span.price-max")

		property.Agent = removeSpaces(e.ChildText("p.enquiry-org > .tokeniser-part1"))
		property.Branch = removeSpaces(e.ChildText("p.enquiry-org > .tokeniser-part2"))

		e.ForEach("table#key-info-table tr", func(_ int, el *colly.HTMLElement) {
			switch el.ChildText("th") {
			case "Style":
				property.Style = el.ChildText("td")
			case "Bedrooms":
				beds, err := strconv.Atoi(el.ChildText("td"))
				if err != nil {
					log.Fatalln(err)
				}
				property.Bedrooms = beds
			case "Receptions":
				rcpns, err := strconv.Atoi(el.ChildText("td"))
				if err != nil {
					log.Fatalln(err)
				}
				property.Receptions = rcpns
			case "Bathrooms":
				baths, err := strconv.Atoi(el.ChildText("td"))
				if err != nil {
					log.Fatalln(err)
				}
				property.Bathrooms = baths
			case "Rates":
				property.Rates = el.ChildText("td")
			case "Heating":
				property.Heating = el.ChildText("td")
			case "EPC Rating":
				property.EPC = removeSpaces(el.ChildText("td"))
			case "Status":
				property.Status = el.ChildText("td")
			}
		})

		property.Description = removeSpaces(e.ChildText("div.prop-descr-text"))

		_, err = statement.Exec(
			property.URL,
			property.ID,
			property.UpdatedTime,
			property.Lat,
			property.Lon,
			property.Address,
			property.Postcode,
			property.PriceOffers,
			property.Price,
			property.PriceMin,
			property.PriceMax,
			property.Agent,
			property.Branch,
			property.Style,
			property.Bedrooms,
			property.Receptions,
			property.Bathrooms,
			property.Rates,
			property.Heating,
			property.EPC,
			property.Status,
			property.Description,
		)
		if err != nil {
			log.Fatalln("db statement failed", err)
		}
	})

	c.OnHTML("a.paging-next", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		counter++
		if counter > 0 {
			fmt.Printf("Reached %d pages", counter)
			return
		}
		e.Request.Visit(link)
	})

	c2.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://www.propertypal.com/search?sta=forSale&sta=saleAgreed&sta=sold&st=sale&currency=GBP&term=15&pt=residential")

	file, err := json.MarshalIndent(properties, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile("test.json", file, 0644)
	if err != nil {
		log.Fatalln(err)
	}
}
