package controllers

import (
	"fmt"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/telegram"
	"net/http"
	"time"

	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	"github.com/gin-gonic/gin"
)

func ScoreCount(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}
	var detail abc.ScoreCountDetail

	month := abc.ScoreCountDetail{}
	month.ScoreCountDetails(fmt.Sprintf("where user_id=%d and close_time>'%s'", uid, time.Now().Format("2006-01")))
	detail.ScoreCountDetails(fmt.Sprintf("where user_id=%d and amount>0", uid))
	data := struct {
		Total  float64 `json:"total"`
		Used   float64 `json:"used"`
		Using  float64 `json:"using"`
		Month  float64 `json:"month"`
		Expire float64 `json:"expire"`
	}{
		Total:  detail.Total,
		Used:   detail.Used,
		Using:  detail.Using,
		Month:  month.Total,
		Expire: abc.ExpireScore(uid),
	}

	r.Status, r.Msg, r.Data = 1, "", data
	c.JSON(http.StatusOK, r.Response())
}

func ScoreDetails(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Status, r.Msg, r.Count, r.Data = abc.ScoreDetails(uid, t, startTime, endTime, page, size)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func ScoreUsed(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Status, r.Msg, r.Count, r.Data = abc.ScoreUsed(uid, page, size, startTime, endTime)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func GetScoreConfig(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	r := R{}
	uid := abc.ToInt(c.MustGet("uid"))
	r.Status, r.Msg, r.Data = abc.GetScoreConfig(uid)
	c.JSON(http.StatusOK, r.Response())
}

func ExchangeGoods(c *gin.Context) {
	r := R{}
	language := abc.ToString(c.MustGet("language"))
	id := abc.ToInt(c.PostForm("id"))
	uid := abc.ToInt(c.MustGet("uid"))
	addressId := abc.ToInt(c.PostForm("address_id"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))

	if abc.FindActivityDisableOne("score", user.Path) {
		r.Msg = golbal.Wrong[language][10102]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	status, msg, truename, payamount, goods, balance,isGoods := abc.ExchangeGoods(id, uid, addressId, language)
	r.Status, r.Msg = status, msg
	if status == 1 &&isGoods==0{
		telegram.SendMsg(telegram.TEXT, 7,
			fmt.Sprintf("兑换通知，用户：%s，消耗积分：%.2f， 产品：%s， 剩余积分：%.2f", truename, payamount, goods, balance))
	}else if status == 1 &&isGoods==1{
		telegram.SendMsg(telegram.TEXT, 19,
			fmt.Sprintf("兑换通知，用户：%s，消耗积分：%.2f， 产品：%s， 剩余积分：%.2f", truename, payamount, goods, balance))
	}
	c.JSON(http.StatusOK, r.Response())
}
