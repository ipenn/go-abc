package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/excel"
	file "github.com/chenqgp/abc/third/uFile"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func OrderList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	login := abc.ToInt(c.PostForm("login"))
	uid := abc.ToInt(c.MustGet("uid"))

	orderId := c.PostForm("order_id")
	symbol := c.PostForm("symbol")
	cmd := c.PostForm("cmd")
	start := c.PostForm("start")
	end := c.PostForm("end")

	profit := c.PostForm("profit")

	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	account := abc.GetAccountOne(fmt.Sprintf("login=%d and user_id=%d", login, uid))
	if account.Login <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	where := fmt.Sprintf("login=%d", login)
	if orderId != "" {
		where += fmt.Sprintf(" and order_id=%d", abc.ToInt(orderId))
	}
	if symbol != "" && !strings.Contains(symbol, "'") {
		where += fmt.Sprintf(" and symbol='%s'", symbol)
	}
	if cmd != "" {
		where += fmt.Sprintf(" and cmd=%d", abc.ToInt(cmd))
	}
	if start != "" && end != "" {
		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and close_time >= '%s' and close_time<='%s'", start, end)
		}
	}
	r := ResponseLimit{}
	r.Status = 1

	order := ""
	if profit != "" {
		switch profit {
		case "1":
			order += "profit asc,"
		case "2":
			order += "profit desc,"
		}
	}
	order += "close_time desc"
	m := make(map[string]interface{})
	r.Count, m["list"], m["total"] = abc.GetOrderList(page, size, where, order)
	r.Data = m
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func OrderAllUserList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))

	inviteName := c.PostForm("invite_name")
	trueName := c.PostForm("true_name")
	start := c.PostForm("start")
	end := c.PostForm("end")
	orderId := c.PostForm("order_id")
	login := c.PostForm("login")
	profit := c.PostForm("profit")

	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	where := fmt.Sprintf(` a.login > 0 `)
	if inviteName != "" && !strings.Contains(inviteName, "'") {
		users := abc.GetUsers(fmt.Sprintf("true_name like'%s%%' and path like'%%,%d,%%'", inviteName, uid))
		if len(users) == 0 {
			var p []abc.OrdersSimple
			r.Status, r.Count, r.Data = 1, 0, p
			c.JSON(http.StatusOK, r.Response(page, size, 0))
			return
		}
		for i, user := range users {
			if i == 0 {
				where += fmt.Sprintf(" and (u.path like '%%,%d,%%'", user.Id)
			} else {
				where += fmt.Sprintf(" or u.path like '%%,%d,%%'", user.Id)
			}
		}
		where += ")"
	}
	if trueName != "" && !strings.Contains(trueName, "'") {
		users := abc.GetUsers(fmt.Sprintf("true_name like'%s%%' and path like'%%,%d,%%'", trueName, uid))
		if len(users) == 0 {
			var p []abc.OrdersSimple
			r.Status, r.Count, r.Data = 1, 0, p
			c.JSON(http.StatusOK, r.Response(page, size, 0))
			return
		}
		for i, user := range users {
			if i == 0 {
				where += fmt.Sprintf(` and u.id in (%d`, user.Id)
			} else {
				where += fmt.Sprintf(",%d", user.Id)
			}
		}
		where += `)`
	}
	if inviteName == "" && trueName == "" {
		user := abc.GetUser(fmt.Sprintf("id=%d", uid))
		where += fmt.Sprintf(` and u.path like '%s%%'`, user.Path)
	}
	where2 := " and o.cmd in (0,1)"
	if start != "" && end != "" {
		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where2 += fmt.Sprintf(` and o.close_time>= '%s' and o.close_time<='%s'`, start, end)
		}
	}

	if orderId != "" {
		where2 += fmt.Sprintf(` and o.order_id=%d`, abc.ToInt(orderId))
	}
	if login != "" {
		where2 += fmt.Sprintf(` and o.login=%d`, abc.ToInt(login))
	}

	order := ""
	if profit != "" {
		switch profit {
		case "1":
			order += "o.profit asc,"
		case "2":
			order += "o.profit desc,"
		}
	}
	order += "o.close_time desc"
	r.Status = 1
	m := make(map[string]interface{})
	w := fmt.Sprintf(`	o.login IN (
	SELECT
		a.login 
	FROM
		user u
		LEFT JOIN account a ON u.id = a.user_id 
	WHERE %s
	)
	%s`, where, where2)
	re := false
	if page == 1 && inviteName == "" && trueName == "" && start == "" && end == "" && orderId == "" && login == "" &&
		profit == "" {
		re = true
	}

	r.Count, m["list"], m["total"] = abc.GetOrderAllUserList(re, uid, page, size, w, order)
	r.Data = m
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func OrderExcel(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	login := abc.ToInt(c.PostForm("login"))
	start := c.PostForm("start")
	end := c.PostForm("end")

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	account := abc.GetAccountOne(fmt.Sprintf("login=%d and user_id=%d", login, uid))
	if account.Login <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	if start == "" || end == "" {
		end = abc.FormatNow()
		start = time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	}
	if len(start) <= 10 {
		start += " 00:00:00"
	}
	if len(end) <= 10 {
		end += " 23:59:59"
	}
	where := ""
	if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
		where = fmt.Sprintf("login=%d and order_id > 0 and close_time >= '%s' and close_time<='%s'", login, start, end)
	}
	orders := abc.GetOrders(where)
	fileName := excel.OrderExcel(login, orders)
	path := file.UploadFile(fileName, uid)

	r.Status = 1
	r.Data = path
	c.JSON(http.StatusOK, r.Response())
}

func CustomerOrderInquiry(c *gin.Context) {
	r := &R{}
	response := ResponseLimit{}
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	name := c.PostForm("name")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	profit := abc.ToInt(c.PostForm("profit"))

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	//where := fmt.Sprintf(" o.login IN (SELECT a.login FROM user u LEFT JOIN account a ON u.id = a.user_id WHERE  a.login > 0  and u.path like '%v%%') and o.cmd in (0,1)", u.Path)
	where := fmt.Sprintf(" o.user_id like '%v%%'", u.Path)

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if startTime != "" {
		where += fmt.Sprintf(" AND o.close_time >= '%v'", startTime+" 00:00:00")
	}

	if endTime != "" {
		where += fmt.Sprintf(" AND o.close_time <= '%v'", endTime+" 23:59:59")
	}

	newName := strings.ReplaceAll(name, " ", "")
	if name != "" {
		where += fmt.Sprintf(" AND (REPLACE(u.true_name,' ','') = '%v' OR REPLACE(o.order_id,' ','') = '%v' OR REPLACE(o.login,' ','') = '%v')", newName, newName, newName)
	}

	order := " ORDER BY o.close_time DESC"

	if profit == 1 {
		order = " ORDER BY o.profit ASC"
	}

	if profit == 2 {
		order = " ORDER BY o.profit DESC"
	}

	res1, count, res2 := abc.CustomerOrderInquiry(where, order, page, size)

	m := make(map[string]interface{})
	m["list"] = res1
	if res2 != nil {
		m["volume"] = abc.ToFloat64(abc.PtoString(res2, "volume"))
		m["profit"] = abc.ToFloat64(abc.PtoString(res2, "profit"))
		m["storage"] = abc.ToFloat64(abc.PtoString(res2, "storage"))
		m["commission"] = abc.ToFloat64(abc.PtoString(res2, "commission"))
	}

	response.Data = m
	response.Status = 1
	c.JSON(200, response.Response(page, size, count))
}
