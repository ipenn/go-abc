package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	"github.com/chenqgp/abc/payment/tron"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/brevo"
	"github.com/chenqgp/abc/third/mt4"
	"github.com/chenqgp/abc/third/telegram"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

func Interest(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	t := abc.FormatNow()

	configDepositGoal := abc.GetConf(fmt.Sprintf("id=%d", 8))

	operator, err := abc.SqlOperator(fmt.Sprintf(`select IFNULL(sum(amount),0) as sums from payment where type = 'deposit' and status = 1 
				and create_time like '%%%s%%'`, t[0:7]))
	if err != nil {
		//todo telegram报错
		log.Println(" abc Interests1 ", err)
		telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf("自动任务复利计划利率查询获取失败：%s\n", err))
	}

	sums := abc.ToFloat64(abc.PtoString(operator, "sums"))
	goal, _ := strconv.ParseFloat(configDepositGoal.Value, 64)
	A := 1.0

	if sums/goal < 1 {
		A = math.Pow(math.E, 1-sums/goal)
	} else if sums > 0 && sums > goal {
		A = goal / sums
	}
	//configXAUUSD := GetConf(fmt.Sprintf("id=%d", 9))
	//conf.Flog.Debug("目标：%f， 已存款：%f,  A: %f, XAUUSD:%s", goal, sums, A, configXAUUSD.Value)
	params0 := map[string]interface{}{
		"A":           fmt.Sprintf("%.2f", A),
		"goal":        fmt.Sprintf("%.2f", goal),
		"deposits":    fmt.Sprintf("%.2f", sums),
		"create_time": t,
	}
	mt4.CrmPost("api/interest", params0)

	interest := abc.GetInterest(fmt.Sprintf("create_time > '%s' and type = 0 and login > 0 and user_id > 0", t[0:10]))
	if len(interest) > 0 {
		vip := make(map[int]float64, 0)
		//interest
		//config := GetConf(8)
		var uids []string
		for _, i := range interest {
			uids = append(uids, abc.ToString(i.UserId))
			done := abc.LimiterPer(nonConcurrent.Queue, i.UserId)

			err = abc.UpdateSql("user", fmt.Sprintf("id=%d", i.UserId), map[string]interface{}{
				"wallet_balance": gorm.Expr(fmt.Sprintf("wallet_balance + (%.2f)", i.Fee)),
			})
			if err != nil {
				//todo telegram报错
				log.Println(" abc Interests2 ", err)
				telegram.SendMsg(telegram.TEXT, telegram.TEST,
					fmt.Sprintf("自动任务复利计划用户钱包更新失败,UID:%d,Amount:%2f\n", i.UserId, i.Fee))
			}
			vip[i.UserId] += i.Fee
			time.Sleep(1 * time.Microsecond)
			done()
		}
		uservip := abc.GetUserVips(fmt.Sprintf("grade>2 and user_id in (%s)", strings.Join(uids, ",")))
		for _, user := range uservip {
			done := abc.LimiterPer(nonConcurrent.Queue, user.UserId)

			if vip[user.UserId] != 0 {
				abc.Interest{
					UserId:     user.UserId,
					CreateTime: t,
					Fee:        vip[user.UserId] * abc.InterestVip[user.Grade],
					Type:       4,
				}.CreateInterestAndAddBalance()
				time.Sleep(1 * time.Microsecond)
			}
			done()
		}
	}
}

func AccountLeverage(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	operators, err := abc.SqlOperators(`select * from account where user_id > 0 and login > 100
	and equity-credit >= 50000 and leverage >=800 and group_name like '%STD%' and leverage_fixed = 0`)
	if err != nil {
		//todo telegram消息
		log.Println(" abc AccountLeverage1 ", err)
	}
	if len(operators) > 0 {
		for _, operator := range operators {
			err = abc.UpdateSql("account", fmt.Sprintf("login=%s", abc.PtoString(operator, "login")), map[string]interface{}{
				"leverage": "500",
			})
			if err != nil {
				//todo telegram消息
				log.Println(" abc AccountLeverage2 ", err)
				telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
					"自动任务调整杠杆失败,login:%d,leverage:%d\n", abc.ToInt(abc.PtoString(operator, "login")), 500))
				continue
			}
			params0 := map[string]interface{}{
				"login": abc.PtoString(operator, "login"),
				"value": abc.PtoString(operator, "leverage"),
				"must":  "1",
			}
			mt4.CrmPost("api/leverage", params0)

			user := abc.GetUser(fmt.Sprintf("id=%s", abc.PtoString(operator, "user_id")))
			mail := abc.MailContent(111)
			brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName, abc.PtoString(operator, "login"), "500"), user.Email)
		}
	}

	sqlOperators, err := abc.SqlOperators(`select * from account where user_id > 0 and login > 100 and equity-credit >= 50000
			and leverage >=400 and group_name like '%DMA%' and leverage_fixed = 0`)
	if err != nil {
		return
	}
	if len(sqlOperators) > 0 {
		for _, operator := range sqlOperators {
			err = abc.UpdateSql("account", fmt.Sprintf("login=%s", abc.PtoString(operator, "login")), map[string]interface{}{
				"leverage": "200",
			})
			if err != nil {
				//todo telegram消息
				log.Println(" abc AccountLeverage2 ", err)
				telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
					"自动任务调整杠杆失败,login:%d,leverage:%d\n", abc.ToInt(abc.PtoString(operator, "login")), 200))
				continue
			}
			params0 := map[string]interface{}{
				"login": abc.PtoString(operator, "login"),
				"value": abc.PtoString(operator, "leverage"),
				"must":  "1",
			}
			mt4.CrmPost("api/leverage", params0)

			user := abc.GetUser(fmt.Sprintf("id=%s", abc.PtoString(operator, "user_id")))
			mail := abc.MailContent(111)
			brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName, abc.PtoString(operator, "login"), 200), user.Email)
		}
	}
}

func Margin(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	account := abc.GetAccounts(fmt.Sprintf(`user_id > 0 and margin > 0 and margin_level > 0 and margin_level <= %d
		and experience = 0`, abc.MarginLevel))
	for _, a := range account {
		if a.Login > 0 {
			user := abc.GetUser(fmt.Sprintf("id=%d", a.UserId))
			if user.Status != -1 {
				mail := abc.MailContent(24)
				brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName), user.Email)
			}
		}
	}
}

func PaymentExpired(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	times := time.Now()
	where := fmt.Sprintf(`create_time < '%s' and type ='deposit'
		and pay_name != 'Wire' and pay_name!='USDT' and status = 0`, times.Format("2006-01-02"))
	sql := fmt.Sprintf("update payment set status=-1 where ")
	payment := abc.GetPayment(where)
	for _, p := range payment {
		_, err := abc.SqlOperator(fmt.Sprintf("%s id=%d", sql, p.Id))
		if err != nil {
			log.Println(" abc PaymentExpired1 ", err)
			continue
		}
		user := abc.GetUser(fmt.Sprintf("id=%d", p.UserId))
		mail := abc.MailContent(76)
		brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName), user.Email)

		vip := abc.GetUserVipCashOne(fmt.Sprintf("pay_id=%d", p.Id))
		abc.UpdateSql(vip.TableName(), fmt.Sprintf("id=%d", vip.Id), map[string]interface{}{
			"status": 0,
			"pay_id": 0,
		})
	}

	where2 := fmt.Sprintf(`create_time < '%s' and type ='deposit'
		and pay_name = 'USDT' and status = 0`, times.Format("2006-01-02"))
	payment2 := abc.GetPayment(where2)
	for _, p := range payment2 {
		tron.Do(p.Id)
	}
	payment22 := abc.GetPayment(where2)
	for _, p := range payment22 {
		_, err := abc.SqlOperator(fmt.Sprintf("%s id=%d", sql, p.Id))
		if err != nil {
			log.Println(" abc PaymentExpired2 ", err)
			continue
		}
		user := abc.GetUser(fmt.Sprintf("id=%d", p.UserId))
		mail := abc.MailContent(76)
		brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName), user.Email)

		vip := abc.GetUserVipCashOne(fmt.Sprintf("pay_id=%d", p.Id))
		abc.UpdateSql(vip.TableName(), fmt.Sprintf("id=%d", vip.Id), map[string]interface{}{
			"status": 0,
			"pay_id": 0,
		})
	}

	times = times.AddDate(0, 0, -6)
	where1 := fmt.Sprintf(`create_time < '%s' and type ='deposit'
		and pay_name = 'Wire' and status = 0`, times.Format("2006-01-02"))
	payment1 := abc.GetPayment(where1)
	for _, p := range payment1 {
		_, err := abc.SqlOperator(fmt.Sprintf("%s id=%d", sql, p.Id))
		if err != nil {
			log.Println(" abc PaymentExpired3 ", err)
			continue
		}
		user := abc.GetUser(fmt.Sprintf("id=%d", p.UserId))
		mail := abc.MailContent(76)
		brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName), user.Email)

		vip := abc.GetUserVipCashOne(fmt.Sprintf("pay_id=%d", p.Id))
		abc.UpdateSql(vip.TableName(), fmt.Sprintf("id=%d", vip.Id), map[string]interface{}{
			"status": 0,
			"pay_id": 0,
		})
	}
}

func UserDisable(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	t := time.Now().AddDate(0, -2, 0).Format("2006-01-02 15:04:05")
	t2 := abc.FormatNow()
	operators, err := abc.SqlOperators(fmt.Sprintf(`SELECT
	u.id,
	u.login_time,
	u.true_name,
	u.status, 
	u.email, 
	GROUP_CONCAT(a.login) as login,
	SUM(a.balance) as balance,
	SUM(a.volume) as volume
FROM
	user u
	LEFT JOIN account a ON u.id = a.user_id 
WHERE
	u.login_time < '%s' AND u.login_time > '2019-01-01' 
	AND u.status = 1
	and balance<100 and volume=0
	GROUP BY u.id`, t))
	if err != nil {
		log.Println(" abc UserDisable1 ", err)
		telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
			"获取自动禁用用户失败"))
	}
	if len(operators) > 0 {
		for _, operator := range operators {
			id := abc.PtoString(operator, "id")
			trueName := abc.PtoString(operator, "true_name")
			err = abc.UpdateSql("user", fmt.Sprintf("id=%s", id), map[string]interface{}{
				"status":     -1,
				"login_time": t2,
			})
			if err != nil {
				log.Println(" abc UserDisable2 ", err)
				continue
			}
			msg := fmt.Sprintf("超过2个月没有登录过的 客户 + 代理 自动禁用 id=%s name=%s", id, trueName)
			alog := abc.AdminLog{}
			alog.AdminName = trueName
			alog.Comment = msg
			alog.Keys = "自动禁用"
			alog.CreateTime = abc.FormatNow()
			alog.CreateAdminLog()

			mail := abc.MailContent(79)
			brevo.Send(mail.Title, fmt.Sprintf(mail.Content, abc.PtoString(operator, "true_name")), abc.PtoString(operator, "email"))

			time.Sleep(1 * time.Microsecond)
		}
	}
	t3 := time.Now().AddDate(0, -1, 0).Format("2006-01-02 15:04:05")
	users := abc.GetUsers(fmt.Sprintf("login_time like '%s%%' and status = 1", t3[0:10]))
	for _, user := range users {
		mail := abc.MailContent(78)
		brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName), user.Email)
	}
}

func SalesDisable(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	t := time.Now().AddDate(0, 0, -11).Format("2006-01-02 15:04:05")
	now := abc.FormatNow()
	users := abc.GetUsers(fmt.Sprintf(`login_time < '%s' and login_time > '2019-01-01' 
		and user_type='sales' and status = 1 and sales_type = 'admin'`, t))
	for _, user := range users {
		err := abc.UpdateSql("user", fmt.Sprintf("id=%d", user.Id), map[string]interface{}{
			"status":     -1,
			"login_time": now,
		})
		if err != nil {
			log.Println(" abc SalesDisable ", err)
			continue
		}
		msg := fmt.Sprintf("超过10天没有登录过的公司业务员自动禁用 id=%d name=%s", user.Id, user.TrueName)
		alog := abc.AdminLog{}
		alog.AdminName = user.TrueName
		alog.Comment = msg
		alog.Keys = "自动禁用"
		alog.CreateTime = abc.FormatNow()
		alog.CreateAdminLog()
	}
}

func AccountExamine(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	accounts := abc.GetAccounts("enable = -1 and user_id > 0 and login < 0")
	for _, account := range accounts {
		user := abc.GetUser(fmt.Sprintf("id=%d", account.UserId))
		if user.Id == 0 {
			//conf.Flog.Info("自动审核 用户不存在...uid=%d group=%s", list.UserId, list.GroupName)
			continue
		}
		//conf.Flog.Info("自动审核开户申请...uid=%d group=%s", list.UserId, list.GroupName)
		readOnly := 0
		experience := 0
		zipcode := "99"
		moveToGroup := ""
		if strings.Index(account.GroupName, "-DMA-") > -1 {
			readOnly = 1
		}
		old := abc.GetAccountOne(fmt.Sprintf("login > 100000 and user_id = %d", account.UserId))

		if old.Login > 100000 {
			params0 := map[string]interface{}{
				"login": old.Login,
			}
			m := mt4.CrmPost("api/admin_user_info", params0)

			zipcode = abc.ToString(m["zipcode"])
			rm := abc.GetRiskgroupMapping(fmt.Sprintf(
				"current_risk_group = '%s' and client_apply='%s'", abc.ToString(m["group"]), account.GroupName))
			if rm.Id > 0 {
				moveToGroup = rm.NewRiskGroup
				//conf.Flog.Info("查询newriskgroup=%s", moveToGroup)
			}
		}
		last := abc.GetAccountOne(fmt.Sprintf("group_name='%s' and enable=1", account.GroupName))
		params := map[string]interface{}{
			"login":           last.Login + 1,
			"name":            account.Name,
			"email":           "",
			"group":           account.GroupName,
			"country":         "",
			"city":            "",
			"address":         "",
			"phone":           "",
			"zipcode":         zipcode,
			"read_only":       fmt.Sprintf("%d", readOnly),
			"id":              fmt.Sprintf("%d", account.UserId),
			"noswap":          "0",
			"experience":      fmt.Sprintf("%d", experience),
			"send_reports":    "1",
			"leverage":        "200",
			"leverage_status": "1",
		}
		//fmt.Println(fmt.Sprintf("%+v", params))
		m := mt4.CrmPost("api/openaccount", params)
		//fmt.Println(fmt.Sprintf("%+v", m))
		if m == nil {
			continue
		}
		code, ok := m["code"]
		if !ok {
			continue
		}
		if abc.ToInt(code) == 1 {
			acc := abc.ToString(m["data"])
			a := strings.Split(acc, ",")
			if len(a) < 3 {
				//conf.Flog.Info("自动开户返回数据不正确...uid=%d", user.Id)
				//go service.TeleSendData(3, fmt.Sprintf("name=%s email=%s 开户生成账户失败，请手动处理！", user.TrueName, user.Email))
				continue
			}
			login := abc.ToInt(a[0])
			if login == 0 {
				//go service.TeleSendData(3, fmt.Sprintf("name=%s email=%s 开户生成账户失败，请手动处理！", user.TrueName, user.Email))
				continue
			}
			if len(moveToGroup) > 0 && account.GroupName != moveToGroup {
				time.Sleep(2 * time.Second)
				//移动到风控组
				param := map[string]interface{}{
					"login": login,
					"k":     "group",
					"v":     moveToGroup,
				}
				risk := mt4.CrmPost("api/admin_user_info", param)

				if abc.ToInt(risk["code"]) == 1 {
					//写入admin—log
					alog := abc.AdminLog{
						AdminName:  "user:",
						CreateTime: abc.FormatNow(),
						Keys:       "移动到风控组",
						Comment:    fmt.Sprintf(" 将账户 %s 从：%s 移动到：%s", a[0], account.GroupName, moveToGroup),
					}
					alog.CreateAdminLog()
				} else {
					continue
				}
			}
			err := abc.UpdateSql("account", fmt.Sprintf(`login='%d' and user_id=%d`, account.Login, account.UserId),
				map[string]interface{}{
					"reg_time":  abc.FormatNow(),
					"user_path": user.Path,
					"login":     login,
					"enable":    1,
				})
			if err != nil {
				//go service.TeleSendData(3, fmt.Sprintf("name=%s email=%s 开户生成账户失败，请手动处理！", user.TrueName, user.Email))
				telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
					"自动任务自动开户增加账户失败,UID:%d,login:%d,password:%s", user.Id, login, a[1]))
				continue
			}
			mail := abc.MailContent(4)
			brevo.Send(mail.Title, fmt.Sprintf(mail.Content, user.TrueName, a[0], a[1]), user.Email)
			//telMsg := fmt.Sprintf("自动审核账号 uid=%d login=%d type=%s group=%s", user.Id, login, user.UserType, list.GroupName)
			//go service.TeleSendData(13, telMsg)
		}
	}
}

func ScoreExpired(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	t := time.Now().AddDate(0, -12, 0).Format("2006-01")
	operators, err := abc.SqlOperators(fmt.Sprintf(`select user_id, sum(amount - use_amount) amount from score_log 
			where create_time < '%s' and use_amount < amount and amount > 0 group by user_id`, t))
	if err != nil {
		//todo telegram消息
		log.Println(" abc ScoreExpired1 ", err)
	}
	if len(operators) > 0 {
		for _, operator := range operators {
			uid := abc.ToInt(abc.PtoString(operator, "user_id"))
			_, err = abc.SqlOperator(fmt.Sprintf(`update score_log set use_amount = amount 
			where create_time < '%s' and use_amount < amount and amount > 0 and user_id = %d`, t, uid))
			if err != nil {
				//todo telegram消息
				log.Println(" abc ScoreExpired2 ", err)
				telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
					"自动任务积分过期更新失败,UID:%d", uid))
				continue
			}
			messageConfig := abc.GetMessageConfig(27)
			amount := abc.PtoString(operator, "amount")
			abc.SendMessage(uid, 27, fmt.Sprintf(messageConfig.ContentZh, amount),
				fmt.Sprintf(messageConfig.ContentHk, amount), fmt.Sprintf(messageConfig.ContentEn, amount))
		}
	}

	t2 := time.Now().AddDate(0, -11, 0).Format("2006-01")
	operators2, err := abc.SqlOperators(fmt.Sprintf(`select user_id, sum(amount - use_amount) amount from score_log 
			where create_time < '%s' and use_amount < amount and amount > 0 group by user_id`, t2))
	if err != nil {
		//todo telegram消息
		log.Println(" abc ScoreExpired3 ", err)
		telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
			"自动任务积分过期提醒查询失败"))
	}
	if len(operators2) > 0 {
		for _, i := range operators2 {
			uid := abc.ToInt(abc.PtoString(i, "user_id"))
			amount := abc.PtoString(i, "amount")
			messageConfig := abc.GetMessageConfig(28)
			abc.SendMessage(uid, 28, fmt.Sprintf(messageConfig.ContentZh, t2, amount),
				fmt.Sprintf(messageConfig.ContentHk, t2, amount), fmt.Sprintf(messageConfig.ContentEn, t2, amount))
		}
	}
}

func VipBirth(c *gin.Context) {
	key := c.PostForm("key")
	if key != "cron-run" {
		return
	}
	t := abc.FormatNow()

	config := abc.GetVipConfig()
	amount := map[int]int{}
	for _, con := range config {
		amount[con.GradeId] = con.BirthCash
	}
	lists := abc.GetUserGrade(t)
	var value []string
	if len(lists) > 0 {
		for _, item := range lists {
			value = append(value, fmt.Sprintf(`(%d,'%s',%d)`,
				item.UserId, t, amount[item.Grade]))
		}
		sql := fmt.Sprintf(`insert into user_vip_cash 
    ( user_id, create_time, deduction_amount) values %s`, strings.Join(value, ","))
		_, err := abc.SqlOperator(sql)
		if err != nil {
			//todo telegram消息
			telegram.SendMsg(telegram.TEXT, telegram.TEST, fmt.Sprintf(
				"自动任务生日礼金插入失败:%s", sql))
		}
	}
}
