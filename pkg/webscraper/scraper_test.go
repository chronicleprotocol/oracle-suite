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

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chronicleprotocol/oracle-suite/pkg/webscraper"
	"github.com/stretchr/testify/assert"
)

func setupServer(resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, resp)
	}))
}

func TestScrape(t *testing.T) {
	resp := `<li class="navAmount " data-col="fundHeader.fundNav.navAmount" data-path="">
<span class="header-nav-label navAmount">
NAV as of 09/Jan/2023
</span>
<span class="header-nav-data">
USD 5.43
</span>
<span class="header-info-bubble">
</span>
<br>
<span class="fiftyTwoWeekData">
52 WK: 5.11 - 5.37
</span>
</li>`
	srv := setupServer(resp)
	defer srv.Close()

	expected := "USD 5.43"
	w := webscraper.NewScraper()
	err := w.FetchAndScrape(srv.URL, "span.header-nav-data", func(e webscraper.Element) {
		txt := strings.Replace(e.Selection.Text(), "\n", "", -1)
		if strings.HasPrefix(txt, "USD ") {
			assert.True(t, txt == expected, "Got the expected value from scrape")
		}
	})
	assert.True(t, err == nil, "No error on scrape")
}

func TestScrapeNotFound(t *testing.T) {
	resp := `<li class="navAmount " data-col="fundHeader.fundNav.navAmount" data-path="">
<span class="header-nav-label navAmount">
NAV as of 09/Jan/2023
</span>
<span>
USD 5.43
</span>
<span class="header-info-bubble">
</span>
<br>
<span class="fiftyTwoWeekData">
52 WK: 5.11 - 5.37
</span>
</li>`
	srv := setupServer(resp)
	defer srv.Close()

	expected := ""
	w := webscraper.NewScraper()
	err := w.FetchAndScrape(srv.URL, "span.header-nav-data", func(e webscraper.Element) {
		txt := strings.Replace(e.Selection.Text(), "\n", "", -1)
		if strings.HasPrefix(txt, "USD ") {
			expected = txt
		}
	})
	assert.True(t, expected == "", "Scrape for price not found")
	assert.True(t, err == nil, "No error on scrape")
}

func TestScrapeMissingPrice(t *testing.T) {
	resp := `<li class="navAmount " data-col="fundHeader.fundNav.navAmount" data-path="">
<span class="header-nav-label navAmount">
NAV as of 09/Jan/2023
</span>
<span class="header-info-bubble">
</span>
<br>
<span class="fiftyTwoWeekData">
52 WK: 5.11 - 5.37
</span>
</li>`
	srv := setupServer(resp)
	defer srv.Close()

	w, err := webscraper.NewScraper().WithPreloadedDoc(srv.URL)
	assert.True(t, err == nil, "No error on Scraper init")

	expected := ""
	err = w.Scrape("span.header-nav-data", func(e webscraper.Element) {
		txt := strings.Replace(e.Selection.Text(), "\n", "", -1)
		if strings.HasPrefix(txt, "USD ") {
			expected = txt
		}
	})
	assert.True(t, expected == "", "Scrape for price not found")
	assert.True(t, err == nil, "No error on scrape")
}
