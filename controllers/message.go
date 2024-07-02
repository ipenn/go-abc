package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func UserMessageList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Status = 1
	r.Count, r.Data = abc.UserMessageList(uid, page, size)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func AnnouncementList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	lang := abc.SwitchLanguage(language)
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Status = 1
	r.Count, r.Data = abc.GetAnnouncementList(uid, page, size, lang)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func AnnouncementRead(c *gin.Context) {
	id := abc.ToInt(c.PostForm("id"))
	abc.AnnouncementRead(id)
	r := R{}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func ReadUserMessage(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	r := R{}
	if id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	err := abc.UpdateSql("user_message", fmt.Sprintf("id=%d and user_id=%d", id, uid), map[string]interface{}{
		"status":    1,
		"read_time": abc.FormatNow(),
	})

	if err != nil {
		log.Println(" abc ReadUserMessage ", err)
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func ReadAllUserMessage(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	err := abc.UpdateSql("user_message", fmt.Sprintf("user_id=%d", uid), map[string]interface{}{
		"status": 1,
	})
	r := R{}
	if err != nil {
		log.Println(" abc ReadAllUserMessage ", err)
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func GetUserMessageNotRead(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}
	r.Status, r.Data = 1, abc.GetUserMessageCount(fmt.Sprintf("user_id=%d and status=0", uid))
	c.JSON(http.StatusOK, r.Response())
}
