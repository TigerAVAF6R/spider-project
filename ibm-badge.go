package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"github.com/tealeg/xlsx"
)

var (
	// BadgeItemList : data list for final result
	BadgeItemList map[string]*BadgeItem
)

// BadgeItem : model for one badge item details
type BadgeItem struct {
	Title  string
	Link   string
	Labels map[string]string
	Skills []string
}

func main() {
	BadgeItemList = make(map[string]*BadgeItem)

	currentDetailLink := ""

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("www.youracclaim.com"),
	)

	// clone a collector to get badge details
	detailCollector := c.Clone()

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.HasPrefix(link, "/org/ibm/badge/") {
			title := e.Attr("title")

			// Print link
			//fmt.Printf("Link found: %s -> %s\n", title, link)

			fullLink := e.Request.AbsoluteURL(link)

			_, ok := BadgeItemList[fullLink]
			if ok {
				// already in list, do nothing
			} else {
				// add a new object into final list
				var item = new(BadgeItem)
				item.Title = title
				item.Link = fullLink
				BadgeItemList[fullLink] = item
			}

			detailCollector.Visit(fullLink)
		}

	})

	detailCollector.OnHTML("ul[class]", func(e *colly.HTMLElement) {
		if e.Attr("class") == "cr-badges-template-attributes cr-badges-template-attributes--normal hide-mobile" {
			e.ForEach("li.cr-badges-template-attributes__item", func(_ int, e1 *colly.HTMLElement) {
				label := e1.ChildText(".cr-badges-template-attributes__label")
				value := e1.ChildText(".cr-badges-template-attributes__value")
				//fmt.Printf("ChildText: %s -> %s\n", label, value)

				item, ok := BadgeItemList[currentDetailLink]
				if ok {
					// already in list, get existing object, to add labels value
					if item.Labels != nil {
						labels := item.Labels
						labels[label] = value
					} else {
						var labels = make(map[string]string)
						labels[label] = value
						item.Labels = labels
					}
				}
			})
		}
	})

	detailCollector.OnHTML("ul[class]", func(e *colly.HTMLElement) {
		if e.Attr("class") == "cr-badges-badge-skills__skills" {
			var skills []string
			e.ForEach("li", func(_ int, e1 *colly.HTMLElement) {
				skill := e1.DOM.Find("a").Text()

				//fmt.Printf("skill: %s\n", skill)
				skills = append(skills, skill)
			})

			item, ok := BadgeItemList[currentDetailLink]
			if ok {
				if item.Skills == nil {
					item.Skills = skills
				}
			}
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	detailCollector.OnRequest(func(r *colly.Request) {
		detailLink := r.URL.String()
		//fmt.Println("Visiting", detailLink)
		currentDetailLink = detailLink
	})

	// Start scraping on https://hackerspaces.org
	url := "https://www.youracclaim.com/organizations/ibm/badges?page=:page"
	maxPage := 46
	//maxPage := 1

	for page := 1; page <= maxPage; page++ {
		newURL := strings.ReplaceAll(url, ":page", fmt.Sprintf("%d", page))
		c.Visit(newURL)
	}

	//printResult()
	generateExcel()
	fmt.Println("########################## Over #######################")
}

func printResult() {
	fmt.Println("########################## Start print final list #######################")
	for k, v := range BadgeItemList {
		fmt.Printf("%s -> %s\n", k, v.Title)
		for kLabel, vLabel := range v.Labels {
			fmt.Printf("%s -> %s\n", kLabel, vLabel)
		}

		for kSkill, vSkill := range v.Skills {
			fmt.Printf("%d -> %s\n", kSkill, vSkill)
		}
	}
}

func generateExcel() {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}

	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "Badge Title"

	cell = row.AddCell()
	cell.Value = "Badge Link"

	cell = row.AddCell()
	cell.Value = "Badge Type"

	cell = row.AddCell()
	cell.Value = "Badge Level"

	cell = row.AddCell()
	cell.Value = "Badge Time"

	cell = row.AddCell()
	cell.Value = "Badge Cost"

	cell = row.AddCell()
	cell.Value = "Badge Skills"

	for k, v := range BadgeItemList {
		row = sheet.AddRow()

		cell = row.AddCell()
		cell.Value = v.Title

		cell = row.AddCell()
		cell.Value = k

		// lables columns start
		noValue := "N/A"
		labelMap := v.Labels
		valueType, ok := labelMap["Type"]
		if ok {
			cell = row.AddCell()
			cell.Value = valueType
		} else {
			cell = row.AddCell()
			cell.Value = noValue
		}

		valueLevel, ok := labelMap["Level"]
		if ok {
			cell = row.AddCell()
			cell.Value = valueLevel
		} else {
			cell = row.AddCell()
			cell.Value = noValue
		}

		valueTime, ok := labelMap["Time"]
		if ok {
			cell = row.AddCell()
			cell.Value = valueTime
		} else {
			cell = row.AddCell()
			cell.Value = noValue
		}

		valueCost, ok := labelMap["Cost"]
		if ok {
			cell = row.AddCell()
			cell.Value = valueCost
		} else {
			cell = row.AddCell()
			cell.Value = noValue
		}
		// lables columns end

		// skills set related to this badge
		if v.Skills != nil {
			strValue := strings.Replace(strings.Trim(fmt.Sprint(v.Skills), "[]"), " ", ",", -1)
			cell = row.AddCell()
			cell.Value = strValue
		} else {
			cell = row.AddCell()
			cell.Value = noValue
		}
	}

	err = file.Save("IBM-Badges.xlsx")
	if err != nil {
		fmt.Printf(err.Error())
	}
}
