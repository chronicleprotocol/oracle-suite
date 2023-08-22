//  Copyright (C) 2021-2023 Chronicle Labs, Inc.
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

// Lightweight package that lets you use jQuery-like syntax to scrape web pages.
// The WithPreloaded* modifiers can be used if you need to query a single doc
// multiple times.

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"

	"github.com/chronicleprotocol/oracle-suite/pkg/log"
)

const LoggerTag = "SCRAPER"

type Scraper struct {
	preloadedDoc *goquery.Document
	log          log.Logger
}

type Element struct {
	Text  string
	Index int
}

func NewScraper() *Scraper {
	return &Scraper{}
}

// To be useful WithLogger should come ahead of With...() functions
func (o *Scraper) WithLogger(l log.Logger) *Scraper {
	o.log = l.WithField("tag", LoggerTag)
	return o
}

func (o *Scraper) WithPreloadedDoc(ctx context.Context, url string) (*Scraper, error) {
	doc, err := o.loadDoc(ctx, url)
	if err != nil {
		return nil, err
	}
	o.preloadedDoc = doc
	return o, nil
}

func (o *Scraper) WithPreloadedDocFromBytes(b []byte) (*Scraper, error) {
	if o.log != nil {
		o.log.WithField("document", string(b)).Debug("Preloaded document")
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	o.preloadedDoc = doc
	return o, nil
}

func (o *Scraper) FetchAndScrape(ctx context.Context, url, queryPath string, handler func(e Element)) error {
	doc, err := o.loadDoc(ctx, url)
	if err != nil {
		return err
	}
	doc.Find(queryPath).Each(
		func(i int, s *goquery.Selection) {
			handler(Element{
				Text:  s.Text(),
				Index: i,
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
				Text:  s.Text(),
				Index: i,
			})
		},
	)
	return nil
}

func (o *Scraper) loadDoc(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	if o.log != nil {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		o.log.WithField("document", string(body)).Debug("Preloaded document")
	}

	return goquery.NewDocumentFromReader(res.Body)
}
