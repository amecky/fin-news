package handler

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/amecky/fin-news/aggregator"
	"github.com/amecky/fin-news/model"
	"github.com/amecky/fin-news/repository"
	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	repo *repository.NewsFeedRepository
}

func NewNewsHandler(repo *repository.NewsFeedRepository) *NewsHandler {
	return &NewsHandler{
		repo: repo,
	}
}

func (h *NewsHandler) FindAllFeeds(c *gin.Context) {
	news := h.repo.FindAll()
	c.IndentedJSON(http.StatusOK, news)
}

func (h *NewsHandler) FindAll(c *gin.Context) {
	limit := 50
	ls := c.Query("limit")
	if ls != "" {
		limit, _ = strconv.Atoi(ls)
	}
	news := h.repo.FindRecentNews(limit)
	c.IndentedJSON(http.StatusOK, news)
}

func (h *NewsHandler) LoadNews(c *gin.Context) {
	items := aggregator.AggregateNews()
	cnt := 0
	for _, item := range items {
		rn := model.RssNews{
			FeedID:    item.FeedId,
			Title:     item.Title,
			Url:       item.Url,
			Timestamp: item.Timestamp,
		}
		if len(rn.Title) >= 250 {
			rn.Title = rn.Title[0:250]
		}
		hasher := md5.New()
		hasher.Write([]byte(item.UID))
		rn.Uid = hex.EncodeToString(hasher.Sum(nil))
		tmp := h.repo.FindNews(item.FeedId, rn.Uid)
		if tmp == nil {
			h.repo.CreateNews(rn)
			cnt++
		}
	}
	h.repo.CleanupNews(5)
}

func (h *NewsHandler) DeleteItem(c *gin.Context) {
	ids := c.Param("id")
	if ids != "" {
		id, _ := strconv.Atoi(ids)
		h.repo.Delete(id)
		c.IndentedJSON(http.StatusCreated, nil)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "ID required"})
	}
}
