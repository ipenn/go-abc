package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/gin-gonic/gin"
)

func ActivityList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	tag := c.PostForm("tag")
	r := R{}

	user := abc.GetUserById(uid)
	t := abc.FormatNow()

	where := fmt.Sprintf(`lang='%s' and start_time<'%s'
		and (user_type = '0' or user_type like '%%%s%%')`, language, t, user.UserType)
	if tag != "" {
		where += fmt.Sprintf(" and tag=%d", abc.ToInt(tag))
	}

	if abc.FindActivityDisableOne("lottery", user.Path) {
		where += " and keyword!='lottery'"
	}
	if abc.FindActivityDisableOne("score", user.Path) {
		where += " and keyword!='score'"
	}

	r.Status = 1
	data := abc.GetActivities(where)
	for i, d := range data {
		if d.EndTime < t {
			data[i].Status = 0
		}
	}
	r.Data = data
	c.JSON(http.StatusOK, r.Response())
}

func LotteryConfig(c *gin.Context) {
	r := R{}
	r.Status, r.Data = 1, abc.FindLotteryConfig()
	c.JSON(http.StatusOK, r.Response())
}

func ActivityRecord(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	t := c.PostForm("type")
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
	if !strings.HasPrefix(t, "o") && !strings.HasPrefix(t, "f") && !strings.HasPrefix(t, "a") {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	count, detail := abc.LotteryDetailsLimit(uid, page, size, " and status=1", t)

	ActivityRecordReturn := struct {
		UnusedCount   int `json:"unused_count"`
		LotteryDetail any `json:"lottery_detail"`
	}{
		UnusedCount:   int(abc.LotteryCount(uid, t, " and status=0")),
		LotteryDetail: detail,
	}
	r.Status, r.Data = 1, ActivityRecordReturn
	c.JSON(http.StatusOK, r.Response(page, size, count))
}

func LotteryDraw(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	t := c.PostForm("type")
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}
	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()
	limit := 100
	user := abc.GetUser(fmt.Sprintf("id=%d", uid))

	if abc.FindActivityDisableOne("lottery", user.Path) {
		r.Msg = golbal.Wrong[language][10102]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if abc.FindActivityDisableOne("lottery", user.Path) {
		limit = 1
	}
	lottery := abc.LotteryDetails(uid, " and status=0", t)
	if len(lottery) == 0 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10106]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	award := abc.LotteryDraw(limit, lottery[0])
	time := abc.FormatNow()
	if award.Type == "BALANCE" {
		interest := abc.Interest{
			UserId:     uid,
			Login:      0,
			CreateTime: time,
			Fee:        award.Value,
			Comment:    lottery[0].Level,
			Type:       2,
		}
		if r.Status = interest.CreateInterestAndAddBalance(); r.Status == 0 {
			r.Status, r.Msg = 0, golbal.Wrong[language][10100]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else if award.Type == "COUPON" {
		coupon := abc.Coupon{
			Type:       0,
			UserId:     uid,
			Amount:     award.Value,
			CreateTime: time,
		}
		if r.Status = coupon.CreateCoupon(); r.Status == 0 {
			r.Status, r.Msg = 0, golbal.Wrong[language][10100]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else if award.Type == "CASHVOUCHER" {
		cash := abc.CashVoucher{
			Amount:     award.Value,
			Volume:     abc.VolAmountMap[award.Value],
			UserId:     uid,
			Status:     1,
			CreateTime: time,
			EndTime:    abc.ToAddDay(30),
			CashNo:     fmt.Sprintf("%s%d", abc.RandonNumber(8), uid),
			Comment:    fmt.Sprintf("[抽奖ID=%d,赠送现金券]", lottery[0].Id),
		}
		cash.CreateCashVoucher()
		creditDetail := abc.CreditDetail{
			UserId:     uid,
			Login:      0,
			CreateTime: t,
			OverTime:   abc.ToAddDay(30),
			Balance:    cash.Amount,
			Source:     3,
			Comment:    "",
			Volume:     cash.Volume,
			CouponNo:   fmt.Sprintf("%s%d", abc.RandonNumber(8), uid),
		}
		creditDetail.CreateCreditDetail()
	}
	abc.UpdateSql("lottery_detail", fmt.Sprintf("id=%d", lottery[0].Id), map[string]any{
		"status":      1,
		"result":      award.Value,
		"result_type": award.Type,
		"result_time": time,
		"comment":     award.Id,
	})
	r.Status, r.Data = 1, map[string]any{
		"id":    award.Id,
		"type":  award.Type,
		"value": fmt.Sprintf("%.2f", award.Value),
	}
	c.JSON(http.StatusOK, r.Response())
}

func RecommendPicture(c *gin.Context) {
	c.JSON(http.StatusOK,R{
		Status: 1,
		Msg:    "",
		Data:   abc.FindPoster("id>0"),
	})
}
