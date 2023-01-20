package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/amecky/fin-news/model"
)

type NewsFeedRepository struct {
	db *sql.DB
}

func NewNewsFeedRepository(db *sql.DB) *NewsFeedRepository {
	return &NewsFeedRepository{
		db: db,
	}
}

// rss_news: id,feed_id,timestamp,title,url,uid
func (r *NewsFeedRepository) CreateNews(news model.RssNews) {
	_, err := r.db.Exec("INSERT INTO rss_news (feed_id,timestamp,title,url,uid) VALUES($1, $2, $3, $4, $5)", news.FeedID, news.Timestamp, news.Title, news.Url, news.Uid)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *NewsFeedRepository) CleanupNews(days int) {
	dt := time.Now().AddDate(0, 0, -1*days)
	date := dt.Format("2006-01-02")
	date += " 00:00"
	_, err := r.db.Exec("DELETE FROM rss_news where timestamp <= $1", date)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *NewsFeedRepository) FindNews(feedId int, uid string) *model.RssNews {
	row := r.db.QueryRow("SELECT id,feed_id,timestamp,title,url,uid from rss_news WHERE feed_id = $1 and uid = $2", feedId, uid)

	t := new(model.RssNews)
	err := row.Scan(&t.ID, &t.FeedID, &t.Timestamp, &t.Title, &t.Url, &t.Uid)
	if err == sql.ErrNoRows {
		return nil
	}
	return t
}

func (r *NewsFeedRepository) FindRecentNews(count int) []model.RssNews {
	var items []model.RssNews
	rows, err := r.db.Query("select id,feed_id,timestamp,title,url,uid from rss_news order by timestamp desc LIMIT $1", count)
	if err != nil {
		fmt.Println(err)
		return items
	}
	defer rows.Close()

	for rows.Next() {
		var t model.RssNews

		err := rows.Scan(&t.ID, &t.FeedID, &t.Timestamp, &t.Title, &t.Url, &t.Uid)
		if err != nil {
			fmt.Println(err)
			return items
		}

		items = append(items, t)
	}
	if err = rows.Err(); err != nil {
		fmt.Println(err)
	}
	return items
}

// id,text,created,closed,state,prio
/*


func (r *NewsFeedRepository) Get(id int) *model.NewsFeed {
	row := r.db.QueryRow("SELECT id,text,created,closed,state,prio from NewsFeeds WHERE id = $1", id)

	t := new(model.NewsFeed)
	err := row.Scan(&t.ID, &t.Text, &t.Created, &t.Closed, &t.State, &t.Prio)
	if err == sql.ErrNoRows {
		return nil
	}
	return t
}

func (r *NewsFeedRepository) Update(NewsFeed model.NewsFeed) *model.NewsFeed {
	_, err := r.db.Exec("UPDATE NewsFeeds set text = $1,closed = $2, state = $3,prio = $4 where id = $5", NewsFeed.Text, NewsFeed.Closed, NewsFeed.State, NewsFeed.Prio, NewsFeed.ID)
	if err != nil {
		log.Fatal(err)
	}
	return &NewsFeed
}
*/
func (r *NewsFeedRepository) Delete(id int) {
	_, err := r.db.Exec("DELETE from rss_news where id = $1", id)
	if err != nil {
		log.Fatal(err)
	}
}

func convertNewsRows(rows *sql.Rows) []model.RssNews {
	var items []model.RssNews
	for rows.Next() {
		var t model.RssNews

		err := rows.Scan(&t.ID, &t.FeedID, &t.Timestamp, &t.Title, &t.Url, &t.Uid)
		if err != nil {
			fmt.Println(err)
			return items
		}

		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
	}
	return items
}

func (r *NewsFeedRepository) FindAllByTokens(tokens []string) []model.RssNews {
	var wc string
	for i, t := range tokens {
		if i > 0 {
			wc += " AND "
		}
		wc += "lower(title) like '%" + strings.ToLower(t) + "%'"
	}
	rows, err := r.db.Query("select id,feed_id,timestamp,title,url,uid from rss_news where " + wc + " order by timestamp desc")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	news := convertNewsRows(rows)
	return news
}

func (r *NewsFeedRepository) FindAll() model.NewsFeeds {
	var feeds model.NewsFeeds
	rows, err := r.db.Query("select id,name,url,active from rss_feeds")
	if err != nil {
		fmt.Println(err)
		return feeds
	}
	defer rows.Close()

	for rows.Next() {
		var t model.NewsFeed

		err := rows.Scan(&t.ID, &t.Name, &t.Url, &t.Active)
		if err != nil {
			fmt.Println(err)
			return feeds
		}

		feeds = append(feeds, t)
	}
	if err = rows.Err(); err != nil {
		fmt.Println(err)
	}
	return feeds
}
