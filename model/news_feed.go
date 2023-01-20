package model

type NewsFeed struct {
	ID     int
	Name   string
	Url    string
	Active int
}

type NewsFeeds []NewsFeed

type RssNews struct {
	ID        int
	FeedID    int
	Timestamp string
	Title     string
	Url       string
	Uid       string
}

type NewsItem struct {
	FeedId    int
	FeedName  string
	Timestamp string
	Title     string
	Url       string
	UID       string
}
