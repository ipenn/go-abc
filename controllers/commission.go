package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

func CommissionList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	start := c.PostForm("start")
	end := c.PostForm("end")
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	t := c.PostForm("type")
	ticket := c.PostForm("ticket")
	login := c.PostForm("login")

	where := fmt.Sprintf("ib_id=%d", uid)
	if t != "" {
		where += fmt.Sprintf(" and commission_type=%d", abc.ToInt(t))
	}
	if ticket != "" {
		where += fmt.Sprintf(" and ticket=%d", abc.ToInt(ticket))
	}
	if login != "" {
		where += fmt.Sprintf(" and login=%d", abc.ToInt(login))
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
	m := make(map[string]interface{})
	m["total"] = abc.GetCommissionSum(where)
	r.Status, r.Count, m["list"] = abc.GetCommissionList(page, size, where)
	r.Data = m
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func GetWageData(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	commission_type := abc.ToInt(c.PostForm("commission_type"))
	r := R{}
	wageData := struct {
		WageList []abc.WageReturn `json:"wage_list"`
		Amount   float64          `json:"amount"`
	}{}

	if commission_type == 1 {
		//wage := abc.GetWageData(fmt.Sprintf("user_id=%d and status=1", uid))
		wage := abc.GetSalesCommission(fmt.Sprintf("user_id=%d and status=1", uid))
		for _, w := range wage {
			wageData.Amount += w.Amount
			wageData.WageList = append(wageData.WageList, abc.WageReturn{
				Id:         w.Id,
				UserId:     w.UserId,
				CreateTime: w.CreateTime,
				Status:     w.Status,
				//Type:           w.type,
				Amount:         w.Amount,
				CommissionType: 1,
			})
		}
	} else if commission_type == 2 {
		pay := abc.GetPayment(fmt.Sprintf("user_id=%d and type='transfer' and transfer_login=-1 and status=1", uid))
		for _, payment := range pay {
			wageData.Amount += payment.Amount
			wageData.WageList = append(wageData.WageList, abc.WageReturn{
				Id:             payment.Id,
				UserId:         payment.UserId,
				CreateTime:     payment.CreateTime,
				Status:         payment.Status,
				Type:           payment.Type,
				Amount:         payment.Amount,
				CommissionType: 2,
			})
		}
	} else {
		operators, _ := abc.SqlOperators(fmt.Sprintf(`select *
from ((select id,
              user_id,
              create_time,
              status,
              '' as type,
              amount,
              1  as commission_type
       from sales_commission
       where user_id = %d
         and status = 1
       order by create_time desc)
      union all
      (select id,
              user_id,
              create_time,
              status,
              type,
              amount,
              2 as commission_type
       from payment
       where user_id = %d
         and type = 'transfer'
         and transfer_login = -1
         and status = 1
       order by create_time desc)) w
order by w.create_time desc`, uid, uid))
		if len(operators) > 0 {
			for _, res := range operators {
				amount := abc.ToFloat64(abc.PtoString(res, "amount"))
				wageData.Amount += amount
				wageData.WageList = append(wageData.WageList, abc.WageReturn{
					Id:             abc.ToInt(abc.PtoString(res, "id")),
					UserId:         abc.ToInt(abc.PtoString(res, "user_id")),
					CreateTime:     abc.PtoString(res, "create_time"),
					Status:         abc.ToInt(abc.PtoString(res, "status")),
					Type:           abc.PtoString(res, "type"),
					Amount:         amount,
					CommissionType: abc.ToInt(abc.PtoString(res, "commission_type")),
				})
			}
		}
	}
	r.Status, r.Data = 1, wageData
	c.JSON(http.StatusOK, r.Response())
}

func GetCommissionSum(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	id := c.PostForm("id")
	login := c.PostForm("login")
	where := fmt.Sprintf("uid=%d", uid)

	if t != "" {
		where += fmt.Sprintf(" and commission_type=%d", abc.ToInt(t))
	}
	if id != "" {
		where += fmt.Sprintf(" and id=%d", abc.ToInt(id))
	}
	if login != "" && !strings.Contains(login, "'") {
		where += fmt.Sprintf(" and login=%d", abc.ToInt(login))
	}
	r := R{}
	r.Status, r.Data = 1, abc.GetCommissionSum(where)
	c.JSON(http.StatusOK, r.Response())
}

func IncomeStatisticsDetails(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	t1 := time.Now().Format("2006-01-02") + " 00:00:00"
	t2 := time.Now().Format("2006-01") + "-01 00:00:00"
	todayIncome, todayVolume := abc.StatisticalIncome(uid, t1)
	monthIncome, monthVolume := abc.StatisticalIncome(uid, t2)

	r.Status = 1
	r.Msg = ""
	r.Data = map[string]interface{}{
		"todayIncome": todayIncome,
		"todayVolume": todayVolume,
		"monthIncome": monthIncome,
		"monthVolume": monthVolume,
	}

	c.JSON(200, r.Response())
}

func RevenueList(c *gin.Context) {
	r := &R{}
	response := &ResponseLimit{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	name := c.PostForm("name")
	cType := abc.ToInt(c.PostForm("type"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
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

	where := fmt.Sprintf("c.ib_id = %v", uid)
	if startTime != "" {
		where += fmt.Sprintf(" AND c.close_time >= '%v'", startTime+" 00:00:00")
	}

	if endTime != "" {
		where += fmt.Sprintf(" AND c.close_time <= '%v'", endTime+" 23:59:59")
	}
	newName := strings.ReplaceAll(name, " ", "")
	if name != "" {
		where += fmt.Sprintf(" AND (REPLACE(u1.true_name,' ','') = '%v' OR REPLACE(u2.true_name,' ','') = '%v' OR REPLACE(c.ticket,' ','') = '%v' OR REPLACE(c.login,' ','') = '%v')", newName, newName, newName, newName)
	}

	switch cType {
	//case 0:
	//	where += fmt.Sprintf(" AND FIND_IN_SET(c.commission_type,'0,1,2')")
	case 1:
		where += fmt.Sprintf(" AND c.commission_type = 0")
	case 2:
		where += fmt.Sprintf(" AND c.commission_type IN (1,3)")
	case 3:
		where += fmt.Sprintf(" AND c.commission_type = 2")
	}

	totalVolume := 0.0
	totalFee := 0.0
	res1, count, res2, res3 := abc.RevenueList(where, page, size)

	if res2 != nil {
		totalVolume = abc.ToFloat64(abc.PtoString(res3, "volume"))
		totalFee = abc.ToFloat64(abc.PtoString(res2, "fee"))
	}

	response.Data = map[string]interface{}{
		"list":         res1,
		"total_volume": totalVolume,
		"total_fee":    totalFee,
	}
	response.Status = 1

	c.JSON(200, response.Response(page, size, count))
}

func MyCommissionRatio(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	u := abc.GetUserById(uid)

	r.Status = 1
	r.Msg = ""
	r.Data = abc.GetMyCommissionAuthority(uid, u.RebateCate, abc.GetCommissionType(uid))

	c.JSON(200, r.Response())
}

func NewRevenueList(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	closeTime := c.PostForm("close_time")
	language := abc.ToString(c.MustGet("language"))
	r := &R{}

	if closeTime == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}

	cac := abc.GetCache("uid=? and time=?", uid, closeTime)

	if cac.Id == 0 || cac.Comment == "" {
		operators, err := abc.SqlOperators(`select o.uid,
       o.login,
       u.username,
       u.true_name,
       u.path,
       o.commission_type,
       sum(o.volume)                         as volume,
       sum(o.fee)                            as fee,
       sum(if(o.commission_type >= 0, 1, 0)) as count,
	   group_concat(o.ticket) as tickets
from commission o
         left join user u on o.uid = u.id
where o.ib_id = ?
  and date_format(o.close_time, '%Y-%m-%d') = ?
group by o.uid, o.login, o.commission_type
order by o.uid;`, uid, closeTime)
		if err != nil {
			log.Println(" abc NewRevenueList ", err)
			return
		}
		count := 0
		//loginMap := make(map[string][]string)
		for i, operator := range operators {
			//login := abc.PtoString(operator, "login")
			tickets := strings.Split(abc.PtoString(operator, "tickets"), ",")
			//tickets = append(tickets, loginMap[login]...)
			slices.Sort(tickets)
			tickets = slices.Compact(tickets)
			//loginMap[login] = tickets
			count += abc.ToInt(abc.PtoString(operator, "count"))
			operators[i].(map[string]any)["count"] = abc.ToString(len(tickets))
		}

		//for i, operator := range operators {
		//	login := abc.PtoString(operator, "login")
		//	operators[i].(map[string]any)["count"] = abc.ToString(len(loginMap[login]))
		//	operators[i].(map[string]any)["tickets"] = ""
		//}

		if count >= 500000 {
			marshal, _ := json.Marshal(operators)
			cac.Comment = string(marshal)
			if cac.Id == 0 {
				cac.Uid = uid
				cac.Time = closeTime
				cac.Type = 3
				abc.CreateCache(cac)
			} else {
				abc.SqlOperator(`update cache set comment=? where uid=? and time=?`, cac.Comment, uid, closeTime)
			}
		}
		r.Data = operators
	} else {
		var operators []interface{}
		json.Unmarshal([]byte(cac.Comment), &operators)
		r.Data = operators
	}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func RevenueSum(c *gin.Context) {
	uid := c.MustGet("uid").(int)
	closeTime := c.PostForm("close_time")
	language := abc.ToString(c.MustGet("language"))

	r := &R{}

	if closeTime == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}

	res := struct {
		TotalVolume               float64 `json:"total_volume"`
		TotalFee                  float64 `json:"total_fee"`
		TotalCommission           float64 `json:"total_commission"`
		TotalCommissionDifference float64 `json:"total_commission_difference"`
		TotalCount                int     `json:"total_count"`
		MonthVolume               float64 `json:"Month_volume"`
		MonthFee                  float64 `json:"Month_fee"`
		MonthCommission           float64 `json:"Month_commission"`
		MonthCommissionDifference float64 `json:"Month_commission_difference"`
		MonthCount                int     `json:"Month_count"`
	}{}

	loc, _ := time.LoadLocation("Local") //获取当地时区
	t, _ := time.ParseInLocation("2006-01-02", closeTime, loc)

	operators, err := abc.SqlOperators(`select 
       o.commission_type,
       sum(o.volume)                         as volume,
       sum(o.fee)                            as fee,
       sum(if(o.commission_type >= 0, 1, 0)) as count,
		group_concat(o.ticket) as tickets
from commission o
where o.ib_id = ?
  and date_format(o.close_time, '%Y-%m-%d') = ?
 	group by o.commission_type`, uid, time.Now().Format("2006-01-02"))
	if err != nil {
		log.Println(" abc NewRevenueList ", err)
		return
	}
	var tickets []string
	if len(operators) > 0 {
		for _, operator := range operators {
			tickets = strings.Split(abc.PtoString(operator, "tickets"), ",")
			slices.Sort(tickets)
			tickets = slices.Compact(tickets)
			if abc.PtoString(operator, "commission_type") == "0" {
				if t.Format("2006-01") == time.Now().Format("2006-01") {
					res.MonthCommission += abc.ToFloat64(abc.PtoString(operator, "fee"))
				}
				res.TotalCommission += abc.ToFloat64(abc.PtoString(operator, "fee"))
			}
			if abc.PtoString(operator, "commission_type") == "1" {
				if t.Format("2006-01") == time.Now().Format("2006-01") {
					res.MonthCommissionDifference += abc.ToFloat64(abc.PtoString(operator, "fee"))
				}
				res.TotalCommissionDifference += abc.ToFloat64(abc.PtoString(operator, "fee"))
			}
			if abc.PtoString(operator, "commission_type") == "2" {
				if t.Format("2006-01") == time.Now().Format("2006-01") {
					res.MonthFee += abc.ToFloat64(abc.PtoString(operator, "fee"))
				}
				res.TotalFee += abc.ToFloat64(abc.PtoString(operator, "fee"))
			}
		}
	}
	slices.Sort(tickets)
	tickets = slices.Compact(tickets)
	res.TotalCount += len(tickets)
	res.MonthCount += len(tickets)

	operators2, err := abc.SqlOperator(`select
    sum(volume) as volume
from commission o
where o.ib_id = ?
  and date_format(o.close_time, '%Y-%m-%d') = ?
        group by o.ticket;`, uid, time.Now().Format("2006-01-02"))
	if err != nil {
		log.Println(" abc NewRevenueList ", err)
		return
	}
	if operators2 != nil {
		if t.Format("2006-01") == time.Now().Format("2006-01") {
			res.MonthVolume += abc.ToFloat64(abc.PtoString(operators2, "volume"))
		}
		res.TotalVolume += abc.ToFloat64(abc.PtoString(operators2, "volume"))
	}

	cas := abc.FindCache("uid=? and type in (1,2,3)", uid)
	for _, cache := range cas {
		res.TotalVolume += cache.Volume
		res.TotalFee += cache.Fee
		res.TotalCommission += cache.Commission
		res.TotalCommissionDifference += cache.CommissionDifference
		res.TotalCount += cache.Quantity

		if cache.Type == 2 {
			t2, _ := time.ParseInLocation("2006-01", cache.Time, loc)
			if t.Format("2006-01") == t2.Format("2006-01") {
				res.MonthVolume += cache.Volume
				res.MonthFee += cache.Fee
				res.MonthCommission += cache.Commission
				res.MonthCommissionDifference += cache.CommissionDifference
				res.MonthCount += cache.Quantity
			}
		}
		if cache.Type == 3 {
			t2, _ := time.ParseInLocation("2006-01-02", cache.Time, loc)
			if t.Format("2006-01") == t2.Format("2006-01") {
				res.MonthVolume += cache.Volume
				res.MonthFee += cache.Fee
				res.MonthCommission += cache.Commission
				res.MonthCommissionDifference += cache.CommissionDifference
				res.MonthCount += cache.Quantity
			}
		}
	}
	r.Status, r.Data = 1, res
	c.JSON(http.StatusOK, r.Response())
}

