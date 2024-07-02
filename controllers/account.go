package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	task_guest "github.com/chenqgp/abc/task/task-guest"
	"github.com/chenqgp/abc/third/brevo"
	"github.com/chenqgp/abc/third/telegram"
	"net/http"
	"time"

	"github.com/chenqgp/abc/third/mt4"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetMyAccount(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res, _ := abc.SqlOperators(`SELECT login, group_name, leverage, balance, margin, enable, read_only  FROM account WHERE user_id = ? AND enable != -3`, uid)

	for _, v := range res {
		if strings.Contains(abc.PtoString(v, "group_name"), "STD") {
			v.(map[string]interface{})["account_type"] = 0
			switch abc.ToInt(abc.PtoString(v, "enable")) {
			case 0:
				v.(map[string]interface{})["status"] = 0
			case -1:
				v.(map[string]interface{})["status"] = 1
			case -2:
				v.(map[string]interface{})["status"] = 1
			case 1:
				v.(map[string]interface{})["status"] = 2
			case -4:
				v.(map[string]interface{})["status"] = -4
			}
		}
		if strings.Contains(abc.PtoString(v, "group_name"), "DMA") {
			v.(map[string]interface{})["account_type"] = 1
			switch abc.ToInt(abc.PtoString(v, "enable")) {
			case 0:
				v.(map[string]interface{})["status"] = 0

			case -1:
				v.(map[string]interface{})["status"] = 1
			case -2:
				v.(map[string]interface{})["status"] = 1
			case 1:
				switch abc.ToInt(abc.PtoString(v, "read_only")) {
				case 0:
					v.(map[string]interface{})["status"] = 2
				case 1:
					v.(map[string]interface{})["status"] = 3
				}
			case -4:
				v.(map[string]interface{})["status"] = -4

			}
		}
	}

	accountSlice := abc.GetMyAccount(uid)

	for _, v := range res {
		for _, vv := range accountSlice {
			if abc.ToInt(abc.PtoString(v, "login")) == vv.Login {
				a := GetMt4AccountInfo(vv.Login)
				if a.Code != 0 {
					v.(map[string]interface{})["leverage"] = abc.ToString(a.Leverage)
					v.(map[string]interface{})["balance"] = abc.ToString(a.Balance)
					v.(map[string]interface{})["margin"] = abc.ToString(a.MarginFree)
				}
			}
		}
	}

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}


func CreateAccount(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	groupName := c.PostForm("group_name")

	u := abc.GetUserById(uid)
	ui := abc.GetUserInfoForId(uid)

	if u.UserType == "user" {
		groupName = abc.GetRebate("id = ? and is_invite = 1", u.RebateId).GroupName
	}

	if u.Status != 1 || u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10027]
	}

	res, _ := abc.SqlOperator(`SELECT COUNT(login) num FROM account WHERE user_id = ? and login > 1000 and is_mam = 0`, uid)

	enable := -1
	readOnly := 0

	if strings.Contains(groupName, "DMA") {
		readOnly = 1
	}
	if res != nil {
		accountCount := abc.ToInt(abc.PtoString(res, "num"))

		if accountCount > 0 {
			if !abc.CheckUserIsDeposit(uid) {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10030]

				c.JSON(200, r.Response())

				return
			}
			enable = -2
		}

		if accountCount >= 5 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10028]

			c.JSON(200, r.Response())

			return
		}
	}

	if strings.Contains(u.UserType, "Level") {
		if !abc.CheckUserHaveCode(uid) || u.IbNo == "" {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10029]

			c.JSON(200, r.Response())

			return
		}
	}

	if abc.CheckUserApplying(uid) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10031]

		c.JSON(200, r.Response())

		return
	}

	rebate := abc.GetRebate("group_name = ? and is_invite = 1", groupName)

	if rebate.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	state, msg := abc.InterestAccount(uid, readOnly, groupName, enable, u.TrueName, ui.Country, ui.City, ui.Address, u.Mobile, u.Email, rebate.Id, "200", language, u.Path)

	if state == 0 {
		r.Status = state
		r.Msg = msg

		c.JSON(200, r.Response())

		return
	}

	abc.AddUserLog(uid, "Create Account", u.Email, abc.FormatNow(), c.ClientIP(), groupName)
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func EditLeverage(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	login := c.PostForm("login")
	leverage := abc.ToInt(c.PostForm("leverage"))
	code := c.PostForm("code")

	if login == "" || leverage == 0 || code == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	captcha := abc.VerifySmsCode(u.Mobile, code)

	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]

		c.JSON(200, r.Response())

		return
	}

	account := abc.GetUserAccount(uid, login)
	if account.Login == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10035]

		c.JSON(200, r.Response())

		return
	}

	if account.Login < 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10040]

		c.JSON(200, r.Response())

		return
	}

	if account.Experience != 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10036]

		c.JSON(200, r.Response())

		return
	}

	//if account.LeverageFixed == 1 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10518]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	if leverage != 50 && leverage != 100 && leverage != 200 && leverage != 500 && leverage != 400 && leverage != 800 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	if strings.Contains(account.GroupName, "DMA") && leverage > 400 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10037]

		c.JSON(200, r.Response())

		return
	}

	data := GetMt4AccountPosition(abc.ToInt(login))

	if len(data.List) != 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10512]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	profit := 0.0
	res, _ := abc.SqlOperator(`SELECT IFNULL(sum(profit),0) profit FROM orders WHERE cmd = 6 and login = ?`, login)

	if res != nil {
		profit = abc.ToFloat64(abc.PtoString(res, "profit"))

		if leverage == 800 && profit >= 50000 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10038]

			c.JSON(200, r.Response())

			return
		}

		if strings.Contains(account.GroupName, "DMA") && leverage == 400 && profit >= 50000 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10039]

			c.JSON(200, r.Response())

			return
		}
	}

	var accountInfo abc.AccountInfoData

	m := map[string]interface{}{
		"login": fmt.Sprintf("%d", account.Login),
	}
	res1 := mt4.CrmPost("api/account_info", m)
	str := abc.ToJSON(res1)
	json.Unmarshal(str, &accountInfo)
	if accountInfo.Code == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10095]

		c.JSON(200, r.Response())

		return
	}
	accountInfo.Balance = fmt.Sprintf("%.2f", accountInfo.Balance)
	accountInfo.Credit = fmt.Sprintf("%.2f", accountInfo.Credit)
	accountInfo.Equity = fmt.Sprintf("%.2f", accountInfo.Equity)
	accountInfo.Margin = fmt.Sprintf("%.2f", accountInfo.Margin)
	accountInfo.MarginLevel = fmt.Sprintf("%.2f", accountInfo.MarginLevel)
	accountInfo.Margin = fmt.Sprintf("%.2f", abc.ToFloat64(accountInfo.Margin))
	accountInfo.MarginFree = fmt.Sprintf("%.2f", accountInfo.MarginFree)
	accountInfo.Volume = fmt.Sprintf("%.2f", accountInfo.Volume)

	newAmount := abc.ToFloat64(accountInfo.Equity)

	flag := true
	if strings.Contains(account.GroupName, "STD") && newAmount < 50000 {
		if leverage > 800 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "STD") && (newAmount >= 50000 && newAmount < 100000) {
		if leverage > 500 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "STD") && (newAmount >= 100000 && newAmount < 1500000) {
		if leverage > 200 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "STD") && (newAmount >= 1500000 && newAmount < 2000000) {
		if leverage > 100 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "STD") && newAmount >= 200000 {
		if leverage > 50 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "DMA") && newAmount < 50000 {
		if leverage > 400 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "DMA") && (newAmount >= 50000 && newAmount < 100000) {
		if leverage > 200 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "DMA") && (newAmount >= 100000 && newAmount < 1500000) {
		if leverage > 100 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "DMA") && (newAmount >= 1500000 && newAmount < 2000000) {
		if leverage > 50 {
			flag = false
		}
	}

	if strings.Contains(account.GroupName, "DMA") && newAmount >= 200000 {
		if leverage > 25 {
			flag = false
		}
	}

	if !flag {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10519]

		c.JSON(200, r.Response())

		return
	}

	//需要MT4接口
	//如果MT4接口成功
	if abc.ToInt(account.Leverage) != leverage {
		result := mt4.CrmPost("api/leverage", map[string]interface{}{
			"login": login,
			"value": leverage,
		})

		if result == nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]

			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v", login, "修改杠杆", result))
			return
		}

		mtcode, ok := result["code"]

		if !ok {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]

			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v", login, "修改杠杆", result))
			return
		}
		mtcode = abc.ToInt(mtcode)

		if mtcode == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]

			c.JSON(200, r.Response())

			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v", login, "修改杠杆", result))
			return
		}

		if mtcode == 1 {
			abc.UpdateSql("account", fmt.Sprintf("user_id = %v and login = %v ", uid, login), map[string]interface{}{
				"leverage": abc.ToString(leverage),
			})

			mail := abc.MailContent(14)
			content := fmt.Sprintf(mail.Content, abc.ToString(uid), abc.ToString(login), abc.ToString(leverage))
			brevo.Send(mail.Title, content, u.Email)

			message := abc.GetMessageConfig(9)
			abc.SendMessage(uid, 14, fmt.Sprintf(message.ContentZh, abc.ToString(login), abc.ToString(leverage)), fmt.Sprintf(message.ContentHk, abc.ToString(login), abc.ToString(leverage)), fmt.Sprintf(message.ContentEn, abc.ToString(login), abc.ToString(leverage)))

			abc.AddUserLog(uid, "Leverage Modify", u.Email, time.Now().Format("2006-01-02 15:04:05"), c.ClientIP(), fmt.Sprintf("%v %v => %v", login, leverage, leverage))
			r.Status = 1
			r.Msg = ""
			c.JSON(200, r.Response())

			return
		}

		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v", login, "修改杠杆", result))
		return
	}

	abc.DeleteSmsOrMail(captcha.Id)
	r.Status = 0
	r.Msg = golbal.Wrong[language][10094]

	c.JSON(200, r.Response())
}

func UpdateAccountPassword(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	login := c.PostForm("login")
	code := c.PostForm("code")
	password := c.PostForm("password")
	cType := c.PostForm("type")

	investor := cType
	if !abc.VerifyPasswordFormat(password, 1) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10501]

		c.JSON(200, r.Response())

		return
	}

	if login == "" || code == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	if u.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	capture := abc.VerifySmsCode(u.Mobile, code)

	if capture.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]

		c.JSON(200, r.Response())

		return
	}

	res := mt4.CrmPost("api/password", map[string]interface{}{
		"xlogin":    login,
		"xpassword": password,
		"investor":  investor,
	})

	if res == nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v-%v", login, "修改mt4密码", res, investor))
		return
	}

	co, ok := res["code"]

	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v-%v", login, "修改mt4密码", res, investor))
		return
	}

	if abc.ToInt(co) != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v:%v原因:%v-%v", login, "修改mt4密码", res, investor))
		return
	}

	abc.DeleteSmsOrMail(capture.Id)

	message := abc.GetMessageConfig(6)
	if investor == "1" {
		message.ContentZh = strings.ReplaceAll(message.ContentZh, "交易", "观摩")
		message.ContentHk = strings.ReplaceAll(message.ContentHk, "交易", "观摩")
	}

	abc.SendMessage(uid, 6, fmt.Sprintf(message.ContentZh, login), fmt.Sprintf(message.ContentHk, login), fmt.Sprintf(message.ContentEn, login))

	mail := abc.MailContent(21)
	content := fmt.Sprintf(mail.Content, abc.ToString(login), password)

	brevo.Send(mail.Title, content, u.Email)

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func GetMt4AccountPosition(login int) abc.PositionData {
	var data abc.PositionData
	redisData := abc.RDB.Get(abc.Rctx, fmt.Sprintf("%d-list", login))
	result := redisData.Val()
	if result == "" {
		m := map[string]interface{}{
			"login": fmt.Sprintf("%d", login),
		}
		res := mt4.CrmPost("api/position", m)
		//fmt.Println(fmt.Sprintf("%+v", res))

		str := abc.ToJSON(res)
		json.Unmarshal(str, &data)
		if data.Code == 0 {
			return data
		}
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Lots = fmt.Sprintf("%.2f", data.Lots)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.Profit = fmt.Sprintf("%.2f", data.Profit)
		abc.RDB.Set(abc.Rctx, fmt.Sprintf("%d-list", login), str, 30*time.Second)
	} else {
		json.Unmarshal([]byte(result), &data)
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Lots = fmt.Sprintf("%.2f", data.Lots)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.Profit = fmt.Sprintf("%.2f", data.Profit)
	}
	return data
}

func Position(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	login := abc.ToInt(c.PostForm("login"))
	uid := abc.ToInt(c.MustGet("uid"))
	orderId := c.PostForm("order_id")
	symbol := c.PostForm("symbol")
	start := c.PostForm("start")
	end := c.PostForm("end")
	profit := c.PostForm("profit")
	r := R{}
	account := abc.GetAccountOne(fmt.Sprintf("login=%d and user_id=%d", login, uid))
	if account.Login <= 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	data := GetMt4AccountPosition(login)

	var res []abc.OrdersData
	for _, datum := range data.List {
		if orderId != "" {
			if abc.ToInt(datum.OrderId) != abc.ToInt(orderId) {
				continue
			}
		}
		if symbol != "" {
			if abc.ToString(datum.Symbol) != symbol {
				continue
			}
		}
		if start != "" && end != "" {
			s, e, t := abc.StringToUnix(start), abc.StringToUnix(end), abc.StringToUnix(abc.ToString(datum.OpenTime))
			if t < s || t > e {
				continue
			}
		}
		res = append(res, datum)
	}
	if profit != "" {
		res = abc.SortIntList(profit, res)
	}
	data.List = res
	r.Status, r.Data = 1, data
	c.JSON(200, r.Response())
}

func AccountInfo(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	login := abc.ToInt(c.PostForm("login"))
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}

	account := abc.GetAccountOne(fmt.Sprintf("login=%d and user_id=%d", login, uid))
	if account.Login <= 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	data := GetMt4AccountInfo(login)
	if data.Code == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10100]
	} else {
		r.Status, r.Data = 1, data
	}

	c.JSON(200, r.Response())
}

func GetMt4AccountInfo(login int) abc.AccountInfoData {
	var data abc.AccountInfoData
	redisData := abc.RDB.Get(abc.Rctx, fmt.Sprintf("%d-info", login))
	result := redisData.Val()
	if result == "" {
		m := map[string]interface{}{
			"login": fmt.Sprintf("%d", login),
		}
		res := mt4.CrmPost("api/account_info", m)
		str := abc.ToJSON(res)
		json.Unmarshal(str, &data)
		if data.Code == 0 {
			return data
		}
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Credit = fmt.Sprintf("%.2f", data.Credit)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.MarginLevel = fmt.Sprintf("%.2f", data.MarginLevel)
		data.Margin = fmt.Sprintf("%.2f", abc.ToFloat64(data.Margin))
		data.MarginFree = fmt.Sprintf("%.2f", data.MarginFree)
		data.Volume = fmt.Sprintf("%.2f", data.Volume)
		abc.RDB.Set(abc.Rctx, fmt.Sprintf("%d-info", login), str, 30*time.Second)
	} else {
		json.Unmarshal([]byte(result), &data)
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Credit = fmt.Sprintf("%.2f", data.Credit)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.MarginLevel = fmt.Sprintf("%.2f", data.MarginLevel)
		data.Margin = fmt.Sprintf("%.2f", abc.ToFloat64(data.Margin))
		data.MarginFree = fmt.Sprintf("%.2f", data.MarginFree)
		data.Volume = fmt.Sprintf("%.2f", data.Volume)
	}
	return data
}

func GetTransferAccount(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res, _ := abc.SqlOperators(`SELECT login FROM account WHERE user_id = ? and enable = 1 and login > 0`, uid)

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func GetServerTime(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	limiter, _, consum := abc.Limiter(c.ClientIP(), task_guest.QueueGuest, abc.ForLimiterSecond)
	defer consum()
	if limiter.Burst() == 0 {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(200, r.Response())
		return
	}
	loc := "GMT+3"
	location, _ := time.LoadLocation("Europe/Moscow")
	t := time.Now().In(location)
	//t := abc.GetTimer(c.PostForm("time"))
	mouth := t.Month()
	f := false
	if mouth >= 11 || mouth <= 3 {
		day := t.Day()
		if mouth == 11 {
			if day < 7 {
				w := 0
				for i := 1; i <= 7; i++ {
					week := time.Date(t.Year(), t.Month(), i,
						0, 0, 0, 0, location).Weekday()
					if week == time.Sunday {
						w = i
					}
				}
				if day >= w {
					f = true
				}
			} else {
				f = true
			}
		} else if mouth == 3 {
			if day < 7 {
				w := 0
				for i := 1; i <= 7; i++ {
					week := time.Date(t.Year(), t.Month(), i,
						0, 0, 0, 0, location).Weekday()
					if week == time.Sunday {
						w = i
					}
				}
				if day < w {
					f = true
				}
			}
		} else {
			f = true
		}
	}
	if f {
		loc = "GMT+2"
		t = t.Add(time.Hour)
	}
	r.Status, r.Data = 1, fmt.Sprintf("%s %s", t.Format("2006-01-02 15:04:05"), loc)
	c.JSON(200, r.Response())
}

func UnfreezeAccount(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	code := c.PostForm("code")
	login := abc.ToInt(c.PostForm("login"))
	language := abc.ToString(c.MustGet("language"))

	r := R{}

	u := abc.GetUser(fmt.Sprintf("id=%d", uid))
	captcha := abc.VerifySmsCode(u.Mobile, code)

	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]
		c.JSON(200, r.Response())
		return
	}

	acc := abc.GetAccountOne(fmt.Sprintf("login=%d", login))
	if acc.Login == 0 && acc.Login < 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10035]
		c.JSON(200, r.Response())
		return
	}
	if acc.Enable != -4 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}

	mt4.CrmPost("api/enable", map[string]interface{}{
		"login": acc.Login,
		"value": "1",
	})
	//cod := res["code"].(int)
	//if cod == 1 {
	//	abc.UpdateAccountStatus(fmt.Sprintf("login = %v", acc.Login), map[string]interface{}{
	//		"enable": 1,
	//	})
	//	r.Status = 1
	//	abc.WriteUserLog(uid, "Unfreeze The Account", c.ClientIP(), abc.ToString(acc.Login))
	//} else {
	//	r.Msg = golbal.Wrong[language][10094]
	//}

	abc.UpdateAccountStatus(fmt.Sprintf("login = %v", acc.Login), map[string]interface{}{
		"enable": 1,
	})
	r.Status = 1
	abc.WriteUserLog(uid, "Unfreeze The Account", c.ClientIP(), abc.ToString(acc.Login))

	c.JSON(http.StatusOK, r.Response())
}
