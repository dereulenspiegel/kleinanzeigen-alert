package kleinanzeigen

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const url = "https://www.kleinanzeigen.de/seite:%v/s-%s/k0l%vr%v"

const cityURL = "https://www.kleinanzeigen.de/s-ort-empfehlungen.json?query=%s"

type Querier struct {
	log *slog.Logger
}

func NewQuerier(log *slog.Logger) *Querier {
	return &Querier{log}
}

func (q *Querier) QueryAds(query *Query) ([]*Ad, error) {
	cityCode, _, err := findCityID(q.log, query.cityName)
	if err != nil {
		return nil, fmt.Errorf("failed query city code for city %s: %w", query.cityName, err)
	}
	ads := getAds(q.log, 0, query.term, cityCode, query.radius, &query.maxPrice, &query.minPrice)
	return ads, nil
}

// GetAds gets the ads for the specified page serachterm citycode and radius
func getAds(log *slog.Logger, page int, term string, cityCode int, radius int, maxPrice *float64, minPrice *float64) []*Ad {
	log.Debug("scraping for ads")
	query := fmt.Sprintf(url, page, strings.ReplaceAll(term, " ", "-"), cityCode, radius)
	ads := make([]*Ad, 0)
	c := colly.NewCollector(
		colly.UserAgent("telegram-alert-bot/1.0"),
	)

	c.OnHTML("#srchrslt-adtable", func(adListEl *colly.HTMLElement) {
		adListEl.ForEach(".ad-listitem", func(_ int, e *colly.HTMLElement) {
			if !strings.Contains(e.DOM.Nodes[0].Attr[0].Val, "is-topad") {
				link := e.DOM.Find("a[class=ellipsis]")
				linkURL, _ := link.Attr("href")
				price := strings.TrimSpace(e.DOM.Find("p[class=aditem-main--middle--price-shipping--price]").Text())

				space := regexp.MustCompile(`\s+`)
				location := strings.TrimSpace(e.DOM.Find("div [class=aditem-main--top--left]").Last().Text())

				imageUrl := e.ChildAttr("div.aditem-image > a > div > img", "src")
				description := e.ChildText("div.aditem-main > div.aditem-main--middle > p")

				location = space.ReplaceAllString(location, " ")
				var priceValue float64
				var err error

				if maxPrice != nil && strings.ToLower(price) != "zu verschenken" {
					replacted := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.Trim(price, " "), "VB", ""), "€", ""), ".", ""), " ")

					if len(replacted) == 0 {
						return
					}

					priceValue, err = strconv.ParseFloat(replacted, 64)

					if err != nil {
						log.Warn("could not parse price from ad", "price-string", replacted)
						return
					}

					if priceValue >= *maxPrice {
						log.Debug("price is bigger than requested")
						return
					}

					if minPrice != nil && priceValue < *minPrice {
						log.Debug("price is lower than requested")
						return
					}
				}

				id, idExsits := e.DOM.Find("article[class=aditem]").Attr("data-adid")
				//details := e.DOM.Find("div[class=aditem-details]")
				title := link.Text()
				if idExsits {
					ads = append(ads, &Ad{
						title:       title,
						link:        "https://www.kleinanzeigen.de" + linkURL,
						id:          id,
						price:       priceValue,
						location:    &location,
						imageUrl:    imageUrl,
						description: description,
					})
				}
			}
		})
	})
	c.OnError(func(r *colly.Response, e error) {
		log.Error("error while scraping for ads", "err", e, "term", term, "radius", radius)
	})

	c.Visit(query)

	c.Wait()

	log.Debug("scraped ads for query", "query", term, "number_of_queries", len(ads))

	return ads
}

// FindCityID finds the city by the name/postal code
func findCityID(log *slog.Logger, untrimmed string) (int, string, error) {
	log.Debug("finding city id", "city_search_term", untrimmed)

	city := strings.Trim(untrimmed, " ")

	spaceClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(cityURL, city), nil)

	if err != nil {
		log.Error("could not create the request", "err", err)
		return 0, "", errors.New("could not make request")
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:74.0) Gecko/20100101 Firefox/74.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	res, getErr := spaceClient.Do(req)

	if getErr != nil {
		return 0, "", errors.New("could not send request")
	}

	if res.StatusCode != 200 {
		log.Error("received a wrong status code.", "status_code", res.Status)
		if res.StatusCode == 403 {
			log.Error("ip address might be blocked by kleinanzeigen.")
		}
		return 0, "", errors.New("request for city not successful")
	}

	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		return 0, "", errors.New("could not read response")
	}

	var cities map[string]string

	unmarshalErr := json.Unmarshal(body, &cities)

	if unmarshalErr != nil {
		return 0, "", errors.New("could not parse json")
	}

	if len(cities) == 0 {
		return 0, "", errors.New("could not find city")
	}

	for key, value := range cities {
		cityIDString := []rune(key)

		cityID, err := strconv.Atoi(strings.Trim(string(cityIDString[1:]), " "))

		if err != nil {
			return 0, "", errors.New("could not get cityId")
		}

		log.Debug("found city", "city_id", cityID, "city_name", value)

		return cityID, value, nil
	}

	return 0, "", errors.New("no city id found")
}
