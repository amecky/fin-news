package aggregator

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/amecky/fin-news/model"
	"github.com/go-resty/resty/v2"
)

func RandomString(options []string) string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(options)
	return options[randNum]
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

func GET(target string) string {
	client := &http.Client{Transport: &http.Transport{}}

	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", RandomString(userAgents))

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		body, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return string(body)
		} else {
			fmt.Println(err)
		}
	}
	return ""
}

type NewsReader interface {
	ReadData() []model.NewsItem
}

type MotleyFoolReader struct{}

func NewMotleyFoolReader() NewsReader {
	return &MotleyFoolReader{}
}

func extract(txt, start, end string) string {
	i := strings.Index(txt, start)
	if i >= 0 {
		s := i + len(start)
		txt = txt[s:]
		j := strings.Index(txt, end)
		if j >= 0 {
			return txt[0:j]
		}
	}
	return ""
}

func (a *MotleyFoolReader) ReadData() []model.NewsItem {
	// https://www.fool.de/neueste-schlagzeilen/page/6/
	url := "https://www.fool.de/neueste-schlagzeilen/"
	data := GET(url)
	idx := strings.Index(data, " <ul class=\"article-list\">")
	if idx != -1 {
		data = data[idx:]
	}
	idx = strings.Index(data, "</ul>")
	if idx != -1 {
		data = data[0:idx]
	}
	var ret = make([]model.NewsItem, 0)
	re := regexp.MustCompile("<p class=\"byline\">(.|\n)*?</h2>")
	comments := re.FindAllString(data, -1)
	min := time.Now().Format("15:04")
	for _, c := range comments {
		line := extract(c, "<h2><a href", "</h2>")
		url := extract(line, "\"", "\"")
		entries := strings.Split(url, "/")
		dt := entries[3] + "-" + entries[4] + "-" + entries[5] + " " + min
		headline := extract(line, ">", "</a>")
		ret = append(ret, model.NewsItem{
			Title:     headline,
			FeedName:  "MotleyFool",
			Timestamp: dt,
			Url:       url,
			UID:       url,
			FeedId:    14,
		})
	}
	return ret
}

type SGZertifikateReader struct{}

func NewSGZertifikateReader() NewsReader {
	return &SGZertifikateReader{}
}

type SGEditorialEntry struct {
	Id              int    `json:"Id"`
	Title           string `json:"Title"`
	PublicationDate string `json:"PublicationDate"`
	Slug            string `json:"Slug"`
}
type SGEditorials struct {
	Entries []SGEditorialEntry `json:"Editorials"`
}

func (a *SGZertifikateReader) ReadData() []model.NewsItem {
	fmt.Println("Collecting SG-Zertifikate news")
	url := "https://www.sg-zertifikate.de/EmcWebApi/api/Editorials?categoryId=7617009&page=0&take=20&orderBy=PublicationDate%20desc&includeAttachments=false"
	rc := resty.New()
	res, _ := rc.R().
		SetResult(SGEditorials{}).
		Get(url)

	editorials := res.Result().(*SGEditorials)
	var ret = make([]model.NewsItem, 0)
	for _, e := range editorials.Entries {
		ret = append(ret, model.NewsItem{
			Title:     e.Title,
			FeedName:  "SG-Zertifikate",
			Timestamp: convertPubDate(e.PublicationDate),
			Url:       "https://sg-zertifikate.de/news-detail/" + e.Slug,
			UID:       e.Slug,
			FeedId:    11,
		})
	}
	return ret
}

type RssFeed struct {
	Id   int
	Name string
	Url  string
}

var RSS_URLS = []RssFeed{

	{1, "Finanzen.net", "https://www.finanzen.net/rss/analysen"},
	{2, "Godmode Trader", "https://www.godmode-trader.de/feeds/deutschland-und-europa"},
	{3, "Onvista", "https://www.onvista.de/news/feed/rss.xml?orderBy=datetime"},
	{4, "Finanznachrichten-Chartanalysen", "https://www.finanznachrichten.de/rss-chartanalysen-top"},
	{5, "Finanznachrichten-Analysen", "https://www.finanznachrichten.de/rss-aktien-analysen"},
	{6, "Ariva.de", "https://www.ariva.de/news/finanznachrichten/rss"},
	{7, "Yahoo", "https://finance.yahoo.com/news/rssindex"},
	{8, "CNBC", "https://www.cnbc.com/id/10000664/device/rss/rss.html"},
	{9, "BörseOnline", "https://www.boerse-online.de/rss"},
	{10, "Wikifolio", "https://www.wikifolio.com/de/de/blog/rss"},
	{12, "Reuters", "https://www.handelsblatt.com/contentexport/feed/finanzen"},
	{13, "SeekingAlpha", "https://seekingalpha.com/news/all/feed"},
	{14, "Finanznachrichten-Aktien", "https://www.finanznachrichten.de/rss-aktien-nachrichten"},
	//{15, "Stockworld-DAX-Analysen", "https://www.stock-world.de/nachrichten/rss_feed_nc.m?category=1007"},
	//{16, "Stockworld-MDAX-Analysen", "https://www.stock-world.de/nachrichten/rss_feed_nc.m?category=1008"},
}

type RSSReader struct{}

func NewRSSReader() NewsReader {
	return &RSSReader{}
}

var PUB_DATE_FORMATS = []string{
	time.RFC1123Z,
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"2006-01-02T15:04:05Z",
	"02 Jan 2006 15:04:05 -0700",
	"Mon, 02 Jan 2006 15:04:05 GMT",
	"2006-01-02T15:04:05",
}

func convertPubDate(pubDate string) string {
	for _, f := range PUB_DATE_FORMATS {
		dt, err := time.Parse(f, pubDate)
		if err == nil {
			return dt.Format("2006-01-02 15:04")
		}
	}
	fmt.Println("Unparseable date:", pubDate)
	return pubDate
}

func (a *RSSReader) ReadData() []model.NewsItem {
	var created []model.NewsItem
	for _, feed := range RSS_URLS {
		fmt.Println("calling", feed.Name)
		channel, err := ReadRSS(feed.Url)
		if err == nil {
			for _, item := range channel.Item {
				rn := model.NewsItem{
					FeedId:    feed.Id,
					FeedName:  feed.Name,
					Title:     item.Title,
					Url:       item.Link,
					UID:       item.Link,
					Timestamp: convertPubDate(item.PubDate),
				}
				if item.GUID != "" {
					rn.UID = item.GUID
				}
				if len(rn.Title) >= 250 {
					rn.Title = rn.Title[0:250]
				}
				created = append(created, rn)
			}
		} else {
			fmt.Println(err)
			fmt.Println(err)
		}
	}
	return created
}

var NEWS_READERS = []NewsReader{
	NewSGZertifikateReader(),
	NewRSSReader(),
	NewMotleyFoolReader(),
}

func AggregateNews() []model.NewsItem {
	var ret = make([]model.NewsItem, 0)
	for _, r := range NEWS_READERS {
		items := r.ReadData()
		ret = append(ret, items...)
	}
	return ret
}
