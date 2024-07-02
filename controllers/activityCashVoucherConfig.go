package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ActivityCashVoucher(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}
	t := abc.FormatNow()

	data, err := abc.SqlOperator(fmt.Sprintf(`SELECT * FROM activity_cash_voucher_config
		WHERE user_id=%d and is_del = 0 and start_time <= '%s' and end_time > '%s'`,
		uid, t[:10], t[:10]))
	if err != nil {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status, r.Data = 1, data
	c.JSON(http.StatusOK, r.Response())
}

func ActivityCashVoucherUsedList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	act := abc.GetActivityCashVoucherConfigOne(fmt.Sprintf("user_id=%d and id=%d", uid, id))
	if act.Id == 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Count, r.Data = abc.GetActivityCashVoucherUsedList(page, size, act.No)
	r.Status = 1
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func UseActivityCashVoucher(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	amount := abc.ToFloat64(c.PostForm("amount"))
	email := c.PostForm("email")
	r := R{}
	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()
	u := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if u.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	user := abc.GetUser(fmt.Sprintf("username='%s'", email))
	if user.Id == 0 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10043]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if !strings.Contains(user.Path, fmt.Sprintf(",%d,", u.Id)) {
		r.Status, r.Msg = 0, golbal.Wrong[language][10121]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status, r.Msg = abc.UseActivityCashVoucher(uid, user.Id, 1, amount, language)
	c.JSON(http.StatusOK, r.Response())
}
