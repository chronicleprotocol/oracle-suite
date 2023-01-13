//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package webscraper

// Lightweight package that lets you use jQuery-like query syntax to scrape web
// pages. The WithPreloaded* modifiers can be used if you need to query a
// single doc multiple times.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	preloadedDoc *goquery.Document
	debug        bool
}

type Element struct {
	Selection *goquery.Selection
	Index     int
}

func NewScraper() *Scraper {
	return &Scraper{}
}

// WithDebug is a flag setter and MUST come before other With...() functions
func (o *Scraper) WithDebug() *Scraper {
	o.debug = true
	return o
}

func (o *Scraper) WithPreloadedDoc(url string) (*Scraper, error) {
	doc, err := o.loadDoc(url)
	if err != nil {
		return nil, err
	}
	o.preloadedDoc = doc
	return o, nil
}

func (o *Scraper) WithPreloadedDocFromBytes(b []byte) (*Scraper, error) {
	if o.debug {
		os.Stderr.WriteString(string(b))
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	o.preloadedDoc = doc
	return o, nil
}

func (o *Scraper) FetchAndScrape(url, queryPath string, handler func(e Element)) error {
	doc, err := o.loadDoc(url)
	if err != nil {
		return err
	}
	doc.Find(queryPath).Each(
		func(i int, s *goquery.Selection) {
			handler(Element{
				Selection: s,
				Index:     i,
			})
		},
	)
	return nil
}

func (o *Scraper) Scrape(queryPath string, handler func(e Element)) error {
	if o.preloadedDoc == nil {
		return fmt.Errorf("no preloaded web doc to scrape")
	}

	o.preloadedDoc.Find(queryPath).Each(
		func(i int, s *goquery.Selection) {
			handler(Element{
				Selection: s,
				Index:     i,
			})
		},
	)
	return nil
}

func (o *Scraper) loadDoc(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	if o.debug {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		os.Stderr.WriteString(string(body))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
