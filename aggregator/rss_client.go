package aggregator

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"net/http"
	"time"

	"golang.org/x/text/encoding/ianaindex"
)

type RSSChannel struct {
	XMLName       xml.Name  `xml:"channel"`
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Item          []RSSItem `xml:"item"`
}

//ItemEnclosure struct for each Item Enclosure
type ItemEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

//Item struct for each Item in the Channel
type RSSItem struct {
	Title       string          `xml:"title"`
	Link        string          `xml:"link"`
	Comments    string          `xml:"comments"`
	PubDate     string          `xml:"pubDate"`
	GUID        string          `xml:"guid"`
	Category    []string        `xml:"category"`
	Enclosure   []ItemEnclosure `xml:"enclosure"`
	Description string          `xml:"description"`
	Author      string          `xml:"author"`
	Content     string          `xml:"content"`
	FullText    string          `xml:"full-text"`
	Isin        string          `xml:"fn:isin"`
}

func parseFeed(resp *http.Response) (*RSSChannel, error) {
	defer resp.Body.Close()
	var rss struct {
		Channel RSSChannel `xml:"channel"`
	}
	body, _ := ioutil.ReadAll(resp.Body)
	decoder := xml.NewDecoder(bytes.NewBuffer(body))
	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, fmt.Errorf("charset %s: %s", charset, err.Error())
		}
		if enc == nil {
			// Assume it's compatible with (a subset of) UTF-8 encoding
			// Bug: https://github.com/golang/go/issues/19421
			return reader, nil
		}
		return enc.NewDecoder().Reader(reader), nil
	}

	if err := decoder.Decode(&rss); err != nil {
		return nil, err
	}
	return &rss.Channel, nil

}

//https://www.finanznachrichten.de/rss-nachrichten-meistgelesen

// https://www.finanzen.net/rss/analysen
// url := "https://www.finanznachrichten.de/rss-nachrichten-meistgelesen"

func ReadRSS(url string) (*RSSChannel, error) {

	httpClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return nil, getErr
	}

	if res.StatusCode != 200 {
		fmt.Println("Wrong status code", res.StatusCode)
		return nil, errors.New(fmt.Sprintf("Wrong status code: %d", res.StatusCode))
	}

	channel, err := parseFeed(res)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return channel, nil
}
