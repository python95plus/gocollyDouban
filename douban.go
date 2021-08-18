package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type Movie struct {
	idx    string
	title  string
	year   string
	info   string
	rating string
	url    string
}

func main() {
	fName := "dob.csv"

	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("创建失败 %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"Idx", "Title", "Year", "Info", "Rating", "URL"})
	startUrl := "https://movie.douban.com/top250"

	c := colly.NewCollector()
	extensions.RandomUserAgent(c)
	// c.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 5 * time.Second,
		Parallelism: 5,
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Println(err.Error())
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("start visit: ", request.URL.String())
	})

	c.OnHTML("ol.grid_view", func(e *colly.HTMLElement) {
		e.ForEach("div.info>div.hd>a", func(_ int, el *colly.HTMLElement) {
			href := el.Attr("href")
			if href != "" {
				parseDetail(c, href, writer)
				log.Println("first href ", href)
			}
		})
	})

	c.OnHTML("div.paginator>span.next>a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href != "" {
			e.Request.Visit(e.Request.AbsoluteURL(href))
		}
	})
	c.Visit(startUrl)
}

func parseDetail(collector *colly.Collector, url string, writer *csv.Writer) {
	newCollector := collector.Clone()
	newCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second,
	})
	newCollector.OnRequest(func(request *colly.Request) {
		log.Println("start visit: ", request.URL.String())
	})
	newCollector.OnHTML("body", func(element *colly.HTMLElement) {
		selection := element.DOM.Find("div#content")
		idx := selection.Find("div.top250>span.top250-no").Text()
		title := selection.Find("h1 >span").First().Text()
		year := selection.Find("h1 > span.year").Text()
		info := selection.Find("div#info").Text()
		info = strings.ReplaceAll(info, " ", "")
		info = strings.ReplaceAll(info, "\n", ";")
		rating := selection.Find("strong.rating_num").Text()
		movie := Movie{
			idx:    idx,
			title:  title,
			year:   year,
			info:   info,
			rating: rating,
			url:    element.Request.URL.String(),
		}
		writer.Write([]string{
			idx,
			title,
			year,
			info,
			rating,
			element.Request.URL.String(),
		})
		log.Printf("%+v\n", movie)

	})

	newCollector.Visit(url)
}
