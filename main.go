package main

import (
	"database/sql"
	"fmt"

	"github.com/amecky/config"
	"github.com/amecky/fin-news/handler"
	"github.com/amecky/fin-news/repository"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Config struct {
	Database struct {
		Url      string `yaml:"url"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

func connectDatabase(url string, user string, pwd string) *sql.DB {
	uri := "postgres://" + user + ":" + pwd + "@" + url + "?sslmode=disable"
	db, err := sql.Open("postgres", uri)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func main() {
	var cfg = Config{}
	err := config.ReadConfig(&cfg)
	if err == nil {
		fmt.Println(cfg)
		con := connectDatabase(cfg.Database.Url, cfg.Database.User, cfg.Database.Password)
		repo := repository.NewNewsFeedRepository(con)
		newsHandler := handler.NewNewsHandler(repo)
		r := gin.Default()
		r.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
		r.GET("/feeds", newsHandler.FindAllFeeds)
		// feeds/status
		r.GET("/news", newsHandler.FindAll)
		r.POST("/news/commands/load", newsHandler.LoadNews)
		r.DELETE("/news/:id", newsHandler.DeleteItem)
		r.Run(":8184")
	} else {
		fmt.Println(err)
	}
}
