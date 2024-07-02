package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

func AgreeInterestPlan(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if abc.FindActivityDisableOne("extra", user.Path) {
		r.Msg = golbal.Wrong[language][10102]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	margin := 0.00
	acc := abc.GetAccounts(fmt.Sprintf("user_id=%d and enable=1 and read_only=0 and experience=0", uid))
	for _, account := range acc {
		margin += account.Equity - account.Margin
	}
	if margin < 5000 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10101]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	more := abc.GetUserMore(uid)
	if more.UserId == 0 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	more.ExtraDoc = 1
	if r.Status = abc.SaveUserMore(more); r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status = 1
	r.Msg = golbal.Wrong[language][10122]
	c.JSON(http.StatusOK, r.Response())
}

func InterestData(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}
	r.Status, r.Msg, r.Data = 1, "", abc.GetInterestData(uid)
	c.JSON(http.StatusOK, r.Response())
}
func InterestDateData(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	dateType := c.PostForm("date_type")
	dateNum := abc.ToInt(c.PostForm("date_num"))
	r := R{}
	r.Status, r.Data = abc.GetInterestDateData(uid, dateType, dateNum)
	c.JSON(http.StatusOK, r.Response())
}

func InterestList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	login := c.PostForm("login")
	start := c.PostForm("start")
	end := c.PostForm("end")
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	where := fmt.Sprintf(" user_id=%d", uid)
	if t != "" {
		where += fmt.Sprintf(" and type=%d", abc.ToInt(t))
	}
	if login != "" && !strings.Contains(login, "'") {
		where += fmt.Sprintf(" and login='%s'", login)
	}
	if start != "" && end != "" {

		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", start, end)
		}
	}

	m := make(map[string]interface{})
	m["total"] = abc.InterestSum(where)
	r := ResponseLimit{}
	r.Status, r.Count, m["list"] = abc.GetInterestList(page, size, where)
	r.Data = m
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func InterestSum(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	login := c.PostForm("login")
	start := c.PostForm("start")
	end := c.PostForm("end")
	r := R{}
	where := fmt.Sprintf(" user_id=%d", uid)
	if t != "" {
		where += fmt.Sprintf(" and type=%d", abc.ToInt(t))
	}
	if login != "" && !strings.Contains(login, "'") {
		where += fmt.Sprintf(" and login='%s'", login)
	}
	if start != "" && end != "" {
		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", start, end)
		}
	}

	r.Status, r.Data = 1, abc.InterestSum(where)
	c.JSON(http.StatusOK, r.Response())
}

func CouponList(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	coupon := abc.GetCoupon(fmt.Sprintf("user_id=%d", uid))
	vipCash := abc.GetUserVipCash(fmt.Sprintf("user_id=%d", uid))
	cashVoucher := abc.GetCashVoucher(fmt.Sprintf("user_id=%d", uid))
	r := R{}
	data := struct {
		CouponList []abc.CouponReturn `json:"coupon_list"`
		TotalCount int                `json:"total_count"`
		UseCount   int                `json:"use_count"`
		UsingCount int                `json:"using_count"`
		UsedCount  int                `json:"used_count"`
	}{}
	for _, cou := range coupon {
		end := abc.StringToUnix(cou.UsedEndTime)
		if cou.UsedEndTime == "" {
			end = abc.GetTimer(cou.CreateTime).AddDate(0, 0, 30).Unix()
		}
		data.TotalCount++
		if end < time.Now().Unix() {
			cou.Status = -1
		}
		t := 0
		if cou.Type <= 1 {
			t = cou.Type + 1
		} else if cou.Type == 2 {
			t = 6
		}
		data.CouponList = append(data.CouponList, abc.CouponReturn{
			Id:         cou.Id,
			Type:       t,
			CouponId:   cou.CouponNo,
			Amount:     cou.Amount,
			Status:     cou.Status,
			CreateTime: cou.CreateTime,
			EndTime:    time.Unix(end, 0).Format("2006-01-02 15:04:05"),
		})
		switch cou.Status {
		case 0:
			data.UseCount++
		case 1, 2:
			data.UsingCount++
		case -1:
			data.UsedCount++
		}
	}
	for _, vc := range vipCash {
		data.TotalCount++

		if vc.Status == 1 {
			vc.Status = 2
		}
		t, add := 3, 90
		if vc.PayAmount == 0 {
			t = 4
			add = 1
		}
		end := abc.GetTimer(vc.CreateTime).AddDate(0, 0, add)
		if end.Unix() < time.Now().Unix() {
			vc.Status = -1
		}
		data.CouponList = append(data.CouponList, abc.CouponReturn{
			Id:         vc.Id,
			Type:       t,
			CouponId:   vc.OrderNo,
			Amount:     abc.ToFloat64(vc.DeductionAmount),
			Status:     vc.Status,
			CreateTime: vc.CreateTime,
			EndTime:    end.Format("2006-01-02 15:04:05"),
		})
		switch vc.Status {
		case 0:
			data.UseCount++
		case 1, 2:
			data.UsingCount++
		case -1:
			data.UsedCount++
		}
	}
	for _, voucher := range cashVoucher {
		if abc.GetTimer(voucher.EndTime).Unix() < time.Now().Unix() {
			voucher.Status = -1
		}
		data.TotalCount++
		data.CouponList = append(data.CouponList, abc.CouponReturn{
			Id:         voucher.Id,
			Type:       5,
			CouponId:   voucher.CashNo,
			Amount:     voucher.Amount,
			Status:     voucher.Status,
			CreateTime: voucher.CreateTime,
			EndTime:    voucher.EndTime,
		})
		switch voucher.Status {
		case 0:
			data.UseCount++
		case 1, 2:
			data.UsingCount++
		case -1:
			data.UsedCount++
		}
	}
	data.CouponList = SortCouponList(data.CouponList)
	r.Status, r.Data = 1, data
	c.JSON(http.StatusOK, r.Response())
}

// 排序
func SortCouponList(arr []abc.CouponReturn) []abc.CouponReturn {
	if len(arr) <= 1 {
		return arr
	}
	for i := 1; i < len(arr); i++ {
		for j := i - 1; j >= 0; j-- {
			if abc.GetTimer(arr[j].CreateTime).Unix() < abc.GetTimer(arr[j+1].CreateTime).Unix() {
				swap(arr, j, j+1)
			}

		}
	}
	return arr
}

func swap(arr []abc.CouponReturn, i, j int) []abc.CouponReturn {
	temp := arr[j]
	arr[j] = arr[i]
	arr[i] = temp
	return arr
}

func UseBirthdayCash(c *gin.Context) {
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

	cash := abc.GetUserVipCashOne(fmt.Sprintf("id=%d", id))
	if cash.Id == 0 || cash.UserId != uid {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
	}
	if cash.Status == 1 {
		r.Msg = golbal.Wrong[language][10112]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if abc.GetTimer(cash.CreateTime).AddDate(0, 0, 1).Unix() < time.Now().Unix() || cash.Status == -1 {
		r.Msg = golbal.Wrong[language][10111]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if cash.CreateTime[0:10] != time.Now().Format("2006-01-02") {
		r.Msg = golbal.Wrong[language][10113]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	r.Status = abc.Interest{
		UserId:     cash.UserId,
		CreateTime: abc.FormatNow(),
		Fee:        abc.ToFloat64(cash.DeductionAmount),
		Type:       5,
	}.CreateInterestAndAddBalance()
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	err := abc.UpdateSql("user_vip_cash", fmt.Sprintf("id=%d", cash.Id), map[string]interface{}{
		"status": 1,
	})
	if err != nil {
		log.Println("controllers UseBirthdayCash ", err)
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	//todo 发送telegram消息
	//service.TeleSendData(6, fmt.Sprintf("生日礼金转余额，姓名：%s，金额：%.2f", c.MustGet("true_name").(string), float64(cash.DeductionAmount)))
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func UseCoupon(c *gin.Context) {
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

	coupon := abc.GetCouponOne(fmt.Sprintf("id=%d", id))
	if coupon.Id == 0 || coupon.UserId != uid {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if coupon.Status == 1 || coupon.Status == 2 {
		r.Msg = golbal.Wrong[language][10112]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	//abc.GetTimer(coupon.CreateTime).AddDate(0, 0, 30).Unix() < time.Now().Unix() ||
	if coupon.Status == -1 {
		r.Msg = golbal.Wrong[language][10111]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	//t2 := abc.ToAddDay(1)
	err := abc.UpdateSql("coupon", fmt.Sprintf("id=%d", id), map[string]interface{}{
		"used_start_time": abc.FormatNow(),
		//"used_end_time":   t2,
		"status": 1,
	})
	if err != nil {
		log.Println("controllers UseCoupon ", err)
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func CouponUseCount(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	end := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	coupon := abc.GetCoupon(fmt.Sprintf("user_id=%d and status=0 and create_time>'%s'", uid, end))
	end2 := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	vipCash := abc.GetUserVipCash(fmt.Sprintf("user_id=%d and status=0 and create_time>'%s'", uid, end2))
	cashVoucher := abc.GetCashVoucher(fmt.Sprintf("user_id=%d and status=0 and end_time>'%s'", uid, abc.FormatNow()))
	vipcount := 0
	for _, cash := range vipCash {
		if cash.PayAmount == 0 {
			if cash.CreateTime[0:10] == abc.FormatNow()[0:10] {
				vipcount++
			}
		} else {
			vipcount++
		}
	}
	r := R{}
	r.Status = 1
	r.Data = struct {
		Count int `json:"count"`
	}{
		Count: len(coupon) + len(cashVoucher) + vipcount,
	}
	c.JSON(http.StatusOK, r.Response())
}
