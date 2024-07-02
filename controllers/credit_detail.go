package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func CreditDetailList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	status := c.PostForm("status")
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	where := fmt.Sprintf("user_id=%d", uid)
	if status != "" {
		s := abc.ToInt(status)
		switch s {
		case 0:
			times := abc.FormatNow()
			where += fmt.Sprintf(" and status=0 and over_time>'%s'", times)
		case 1:
			where += fmt.Sprintf(" and status=1")
		case -1:
			times := abc.FormatNow()
			where += fmt.Sprintf(" and status=0 and over_time<'%s'", times)
		}
	}
	//r := R{}
	//var result struct {
	//	Using  []abc.CreditDetailReturn `json:"using"`
	//	Used   []abc.CreditDetailReturn `json:"used"`
	//	Exceed []abc.CreditDetailReturn `json:"exceed"`
	//}
	var result []abc.CreditDetailReturn
	count, data := abc.GetCreditDetailList(page, size, where)

	maxTime := "2006-01-02 15:04:05"
	minTime := "2006-01-02 15:04:05"
	var order []abc.Orders
	if len(data) > 0 {
		for _, datum := range data {
			if minTime < datum.CreateTime {
				minTime = datum.CreateTime
			}
			if maxTime < datum.OverTime {
				maxTime = datum.OverTime
			}
		}
		m := abc.GetTimer(minTime).Add(-6 * time.Hour).Format("2006-01-02 15:04:05")
		order = abc.GetAvailableVolumeOrders(fmt.Sprintf(`login in (select login from account where user_id = %d) and cmd < 2
		and symbol_type < 2  and volume - spent > 0 and close_time>='%s' and close_time<='%s'`, uid, m, maxTime))
	}

	for _, detail := range data {
		res := abc.CreditDetailReturn{
			Id:         detail.Id,
			UserId:     detail.UserId,
			Login:      detail.Login,
			CreateTime: detail.CreateTime,
			OverTime:   detail.OverTime,
			DeductTime: detail.DeductTime,
			Amount:     detail.Balance,
			Source:     detail.Source,
			Comment:    detail.Comment,
			Volume:     detail.Volume,
			CouponNo:   detail.CouponNo,
			Deposit:    detail.Deposit,

			Status: detail.Status,
		}
		if detail.Status == 0 && abc.GetTimer(detail.OverTime).Unix() > time.Now().Unix() {
			for _, spent := range order {

				ct := abc.GetTimer(detail.CreateTime).Add(-6 * time.Hour).Unix()

				if abc.StringToUnix(spent.CloseTime) >= ct && abc.StringToUnix(spent.CloseTime) <= abc.StringToUnix(detail.OverTime) {
					if detail.Login == 0 || (detail.Login != 0 && detail.Login == spent.Login) {
						res.UsableVolume += spent.Volume - spent.Spent
					}
				}
			}
			//result.Using = append(result.Using, res)
		} else if detail.Status == 0 && abc.GetTimer(detail.OverTime).Unix() < time.Now().Unix() {
			res.Status = -1
			//result.Exceed = append(result.Exceed, res)
		} else if detail.Status == 1 {
			//result.Used = append(result.Used, res)
		}
		result = append(result, res)
	}
	r.Status, r.Data = 1, result
	c.JSON(http.StatusOK, r.Response(page, size, count))
}

func UseCreditDetail(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	credit := abc.GetCreditDetailOne(fmt.Sprintf("id=%d", id))
	if credit.Id == 0 || credit.UserId != uid {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if credit.Status == 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10112]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if abc.GetTimer(credit.OverTime).Unix() < time.Now().Unix() {
		r.Status, r.Msg = 0, golbal.Wrong[language][10111]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	where := fmt.Sprintf(`cmd < 2 and user_id like '%%,%d,'
		and symbol_type < 2 and close_time > '%s' and close_time<'%s' and volume - spent > 0`, uid,
		abc.UnixTimeToStr(abc.GetTimer(credit.CreateTime).Unix()-6*3600), credit.OverTime)
	order := abc.GetAvailableVolumeOrders(where)
	if credit.Volume>0{
		sum := 0.00
		for _, o := range order {
			sum += o.Volume - o.Spent
			if sum >= credit.Volume {
				r.Status = 1
				break
			}
		}
		if r.Status == 0 {
			r.Status, r.Msg = 0, golbal.Wrong[language][10114]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	}
	interest := abc.Interest{}
	interest.UserId = credit.UserId
	interest.CreateTime = abc.FormatNow()
	interest.Fee = credit.Balance
	if credit.Source == 2 {
		if abc.GetTimer(credit.CreateTime).Unix() > time.Now().AddDate(0, 0, 30).Unix() {
			r.Status, r.Msg = 0, golbal.Wrong[language][10114]
			c.JSON(http.StatusOK, r.Response())
			return
		}
		interest.Type = 6
	} else if credit.Source == 3 {
		interest.Type = 7
	}
	if interest.Type == 0 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if r.Status = interest.CreateInterestAndAddBalance(); r.Status == 0 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	abc.UpdateSql("credit_detail", fmt.Sprintf("id=%d", credit.Id), map[string]interface{}{
		"status": 1,
		"deduct_time": abc.FormatNow(),
	})
	abc.UpdateOrderSpent(credit.Volume, order)
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}
