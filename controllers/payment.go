package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc/conf"
	alpapay "github.com/chenqgp/abc/payment/yinlian/yinlian-alpapay"
	exlink "github.com/chenqgp/abc/payment/yinlian/yinlian-exlink"
	yinlian_fastPort "github.com/chenqgp/abc/payment/yinlian/yinlian-fastPort"
	yinlian_tops "github.com/chenqgp/abc/payment/yinlian/yinlian-tops"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/chenqgp/abc/payNotice"
	validator "github.com/chenqgp/abc/third/google"

	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	"github.com/chenqgp/abc/payment/h2pay"
	"github.com/chenqgp/abc/payment/tron"
	bft "github.com/chenqgp/abc/payment/yinlian/yinlian-bft"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	nep "github.com/chenqgp/abc/payment/yinlian/yinlian-nep"
	teleport "github.com/chenqgp/abc/payment/yinlian/yinlian-teleport"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	wallet2 "github.com/chenqgp/abc/task/task-wallet"
	"github.com/chenqgp/abc/third/brevo"
	"github.com/chenqgp/abc/third/mt4"
	"github.com/chenqgp/abc/third/telegram"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Help2PayBack struct {
	Id            string `form:"id" json:"id"`
	Amount        string `form:"Amount" json:"Amount"`
	Currency      string `form:"Currency" json:"Currency"`
	Language      string `form:"Language" json:"Language"`
	Reference     string `form:"Reference" json:"Reference"`
	Datetime      string `form:"Datetime" json:"Datetime"`
	Status        string `form:"Status" json:"Status"`
	Key           string `form:"Key" json:"Key"`
	Note          string `form:"Note" json:"Note"`
	StatementDate string `form:"StatementDate" json:"StatementDate"`
	DepositFee    string `form:"DepositFee" json:"DepositFee"`
	ErrorCode     string `form:"ErrorCode" json:"ErrorCode"`
}

type NepSuccessBack struct {
	BillNo string `form:"bill_no" json:"bill_no"`
	Amount string `form:"amount" json:"amount"`
	SysNo  string `form:"sys_no" json:"sys_no"`
	Sign   string `form:"sign" json:"sign"`
}

type NepCancelBack struct {
	BillNo     string `form:"bill_no" json:"bill_no"`
	BillStatus int    `form:"bill_status" json:"bill_status"`
	SysNo      string `form:"sys_no" json:"sys_no"`
	Sign       string `form:"sign" json:"sign"`
}

type FastSuccessBack struct {
	BillNo string `form:"bill_no" json:"bill_no"`
	Amount string `form:"amount" json:"amount"`
	SysNo  string `form:"sys_no" json:"sys_no"`
	Sign   string `form:"sign" json:"sign"`
}

type FastCancelBack struct {
	BillNo     string `form:"bill_no" json:"bill_no"`
	BillStatus int    `form:"bill_status" json:"bill_status"`
	SysNo      string `form:"sys_no" json:"sys_no"`
	Sign       string `form:"sign" json:"sign"`
}

type TeleportBack struct {
	SysNo      string `form:"sys_no" json:"sys_no" bind"`
	BillNo     string `form:"bill_no" json:"bill_no" binding:"required"`
	Amount     string `form:"amount" json:"amount"`
	BillStatus int    `form:"bill_status" json:"bill_status"`
	Sign       string `form:"sign" json:"sign" binding:"required"`
}

type ChipBack struct {
	CompanyOrderNum string `form:"companyOrderNum" json:"companyOrderNum"`
	IntentOrderNo   string `form:"intentOrderNo" json:"intentOrderNo"`
	CoinAmount      string `form:"coinAmount" json:"coinAmount"`
	CoinSign        string `form:"coinSign" json:"coinSign"`
	TradeStatus     string `form:"tradeStatus" json:"tradeStatus"`
	TradeOrderTime  string `form:"tradeOrderTime" json:"tradeOrderTime"`
	UnitPrice       string `form:"unitPrice" json:"unitPrice"`
	Total           string `form:"total" json:"total"`
	SuccessAmount   string `form:"successAmount" json:"successAmount"`
	Sign            string `form:"sign" json:"sign"`
}

type Star7Back struct {
	OrderNo     string `form:"orderNo" json:"orderNo"`
	OrderTime   string `form:"orderTime" json:"orderTime"`
	PayAmt      string `form:"payAmt" json:"payAmt"`
	PayStatus   string `form:"payStatus" json:"payStatus"`
	SuccessTime string `form:"successTime" json:"successTime"`
}

type Pay7Back struct {
	SuccessTime string `form:"successTime" json:"successTime"`
	OrderTime   string `form:"orderTime" json:"orderTime"`
	OrderNo     string `form:"orderNo" json:"orderNo"`
	PayStatus   string `form:"payStatus" json:"payStatus"`
	PayAmt      string `form:"payAmt" json:"payAmt"`
	Sign        string `form:"sign" json:"sign"`
}

//type BftBack struct {
//	ApiOrderNo  string `json:"apiOrderNo"`
//	Money       string `json:"money"`
//	TradeStatus string `json:"tradeStatus"`
//	TradeId     string `json:"tradeId"`
//	UniqueCode  string `json:"uniqueCode"`
//	Signature   string `json:"signature"`
//}

type BftBack struct {
	APIOrderNo  string `json:"apiOrderNo"`
	Money       string `json:"money"`
	UniqueCode  string `json:"uniqueCode"`
	Signature   string `json:"signature"`
	TradeStatus string `json:"tradeStatus"`
	TradeID     string `json:"tradeId"`
}

type EXlinkBack struct {
	APIOrderNo  string `json:"apiOrderNo"`
	Money       string `json:"money"`
	UniqueCode  string `json:"uniqueCode"`
	Signature   string `json:"signature"`
	TradeStatus string `json:"tradeStatus"`
	TradeID     string `json:"tradeId"`
}

func PaymentList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	status := c.PostForm("status")
	payName := c.PostForm("pay_name")
	start := c.PostForm("start_time")
	end := c.PostForm("end_time")
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))

	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	where := fmt.Sprintf("user_id=%d", uid)
	if t == "deposit" || t == "withdraw" || t == "transfer" || t == "commission" {
		where += fmt.Sprintf(" and type='%s'", t)
	}
	if status != "" {
		s := abc.ToInt(status)
		if s < 4 {
			if status == "-2" {
				where += fmt.Sprintf(" and status in (-2,-3)")
			} else {
				where += fmt.Sprintf(" and status=%d", abc.ToInt(status))
			}
		} else if s == 4 {
			where += fmt.Sprintf(" and status=0 and (a_status=0 or b_status=0)")
		} else if s == 5 {
			where += fmt.Sprintf(" and status=0 and a_status=1 and b_status=1")
		} else if s == 6 {
			where += fmt.Sprintf(" and status=0 and c_status=1 ")
		}
	}
	if payName != "" && !strings.Contains(payName, "'") {
		where += fmt.Sprintf(" and pay_name='%s'", payName)
	}
	if start != "" && end != "" {
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", start, end)
		}
	}

	r := ResponseLimit{}
	m := make(map[string]interface{})

	total := abc.GetPaymentAmount(where)
	if t == "withdraw" {
		total.Amount = math.Abs(total.Amount)
	}
	m["total"] = total
	r.Status, r.Count, m["list"] = abc.GetPaymentList(page, size, where)
	r.Data = m
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func UnionPayToPay(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	payName := c.PostForm("pay_name")
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	cashId := abc.ToInt(c.PostForm("cashId"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	//if amount < 200 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10059]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	ui := abc.GetUserInfoById(uid)
	u := abc.GetUserById(uid)
	pConfigSlice := abc.GetPaymentChannel()

	//if ui.IdentityType == "Identity card" && u.Phonectcode != "+86" {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10515]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	var pConfig abc.PaymentConfig

	flag := false
	for _, v := range pConfigSlice {
		if v.Name == payName {
			pConfig = v
			flag = true
			break
		}
	}

	if !flag {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10053]

		c.JSON(200, r.Response())

		return
	}

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("deposits", u.Path) || abc.FindActivityDisableOne("CloseUnionPay", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10062]

		c.JSON(200, r.Response())

		return
	}

	t := time.Now().Format("15:04:05")

	t1 := abc.FormatNow()
	if t < pConfig.OpenTime || t > pConfig.CloseTime {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10055]

		c.JSON(200, r.Response())

		return
	}

	if amount < pConfig.MinAmount {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10056]

		c.JSON(200, r.Response())

		return
	}

	if amount > pConfig.MaxAmount {
		r.Status = 0
		r.Msg = fmt.Sprintf(golbal.Wrong[language][10057], pConfig.MaxAmount)

		c.JSON(200, r.Response())

		return
	}

	//如果存在未支付订单
	//p := abc.QueryUnpaidOrders(uid)
	//if p.Id > 0 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10058]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	var uc abc.UserVipCash
	if cashId != 0 {
		uc = abc.CheckCouponExist(uid, cashId)

		if uc.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10060]

			c.JSON(200, r.Response())

			return
		}

		if amount < abc.ToFloat64(uc.PayAmount) {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10061]

			c.JSON(200, r.Response())

			return
		}
	}

	orderNo := fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid)

	rate := 0.0

	uf := abc.GetUserFile(uid)
	newAmount := amount - abc.ToFloat64(uc.DeductionAmount)
	switch payName {
	case "ChipPay":
		idFront := ""
		idBack := ""
		for _, v := range uf {
			if abc.ToInt(v.Front) == 1 {
				idFront = v.FileName
			}
			if abc.ToInt(v.Front) == 2 {
				idBack = v.FileName
			}
		}

		bank := abc.GetAuditedBankOne(uid)
		r.Status, r.Msg, r.Data, rate = chip.ChipPay(orderNo, u.Phonectcode, u.Mobile, bank.TrueName, ui.Identity, idFront, idBack, language, amount-abc.ToFloat64(uc.DeductionAmount))

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "ChipPay", uid))
			return
		}
	case "Teleport":
		pay := abc.GetPayConfigByName("Teleport")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}
		r.Status, r.Msg, r.Data = teleport.TeleportPay(orderNo, t1, ui.ChineseName, language, amount-float64(uc.DeductionAmount), uid, rate, c.ClientIP())

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "Teleport", uid))
			return
		}
	case "BFT":
		pay := abc.GetPayConfigByName("BFT")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}
		r.Status, r.Msg, r.Data = bft.BftPay(orderNo, u.TrueName, language, amount-float64(uc.DeductionAmount), rate, uid)

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "BFT", uid))
			return
		}
	case "Neptune":
		pay := abc.GetPayConfigByName("TOPS")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}

		if pay.Status > 0{
			r.Status, r.Msg, r.Data = yinlian_tops.TopsPay(orderNo, u.TrueName, u.Mobile, ui.Identity, language, amount-float64(uc.DeductionAmount), rate)
			if r.Status > 0 {
				payName = "TOPS"
			}
		}else{
			r.Status = 0
		}

		if r.Status == 0 {
			pay = abc.GetPayConfigByName("Neptune")
			if pay.ExchangeRate > 6 {
				rate = pay.ExchangeRate
			} else {
				rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
			}

			r.Status, r.Msg, r.Data = nep.NeptunePay(orderNo, ui.ChineseName, c.ClientIP(), abc.FormatNow(), language, amount-float64(uc.DeductionAmount), rate, uid)

			if r.Status == 0 {
				c.JSON(200, r.Response())
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "Neptune", uid))
				return
			}
		}
	case "7Star":
		pay := abc.GetPayConfigByName("7Star")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}

		r.Status, r.Msg, r.Data = alpapay.AlpapayPay(orderNo, language, amount-float64(uc.DeductionAmount), rate)

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "Neptune", uid))
			return
		}
	case "EXlink":
		pay := abc.GetPayConfigByName("TOPS")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}
		if pay.Status > 0{
			r.Status, r.Msg, r.Data = yinlian_tops.TopsPay(orderNo, u.TrueName, u.Mobile, ui.Identity, language, amount-float64(uc.DeductionAmount), rate)
			if r.Status > 0 {
				payName = "TOPS"
			}
		}else{
			r.Status = 0
		}
		if r.Status == 0{
			pay = abc.GetPayConfigByName("EXlink")
			if pay.ExchangeRate > 6 {
				rate = pay.ExchangeRate
			} else {
				rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
			}

			r.Status, r.Msg, r.Data = exlink.EXlinkPay(orderNo, u.TrueName, language, amount-float64(uc.DeductionAmount), rate, uid)

			if r.Status == 0 {
				c.JSON(200, r.Response())
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "EXlink", uid))
				return
			}
		}
	case "Fastport":
		pay := abc.GetPayConfigByName("TOPS")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}
		if pay.Status > 0{
			r.Status, r.Msg, r.Data = yinlian_tops.TopsPay(orderNo, u.TrueName, u.Mobile, ui.Identity, language, amount-float64(uc.DeductionAmount), rate)
			if r.Status > 0 {
				payName = "TOPS"
			}
		}else{
			r.Status = 0
		}
		if r.Status == 0{
			pay = abc.GetPayConfigByName("Fastport")
			if pay.ExchangeRate > 6 {
				rate = pay.ExchangeRate
			} else {
				rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
			}

			r.Status, r.Msg, r.Data = yinlian_fastPort.FastPortPay(orderNo, ui.ChineseName, c.ClientIP(), abc.FormatNow(), language, amount-float64(uc.DeductionAmount), rate, uid)

			if r.Status == 0 {
				c.JSON(200, r.Response())
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "Fastport", uid))
				return
			}
		}


	case "TOPS":
		pay := abc.GetPayConfigByName("TOPS")
		if pay.ExchangeRate > 6 {
			rate = pay.ExchangeRate
		} else {
			rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
		}
		if pay.Status > 0 {
			r.Status, r.Msg, r.Data = yinlian_tops.TopsPay(orderNo, u.TrueName, u.Mobile, ui.Identity, language, amount-float64(uc.DeductionAmount), rate)
			if r.Status > 0 {
				payName = "TOPS"
			}
		}else{
			r.Status = 0
		}
		if r.Status == 0 {
			payName = "Neptune"
			//c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "TOPS", uid))
			//return
			pay = abc.GetPayConfigByName("Neptune")
			if pay.ExchangeRate > 6 {
				rate = pay.ExchangeRate
			} else {
				rate, _ = decimal.NewFromFloat(teleport.GetBaseExchangeRate() * pConfig.ExchangeRate).Round(2).Float64()
			}

			r.Status, r.Msg, r.Data = nep.NeptunePay(orderNo, ui.ChineseName, c.ClientIP(), abc.FormatNow(), language, amount-float64(uc.DeductionAmount), rate, uid)

			if r.Status == 0 {
				c.JSON(200, r.Response())
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "Neptune", uid))
				return
			}
		}
	default:
		r.Status = 0
		r.Msg = golbal.Wrong[language][10053]

		c.JSON(200, r.Response())

		return
	}

	pId := abc.CreatePayment(orderNo, t1, "UnionPay", payName, "", u.Path, r.Data.(string), "", newAmount, float64(uc.DeductionAmount), rate, uid, 1, 1)

	if pId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v创建订单失败", "银联存款", uid))
		return
	}

	if uc.Id > 0 {
		abc.UpdateSql("user_vip_cash", fmt.Sprintf("id = %v", uc.Id), map[string]interface{}{
			"status": 1,
			"pay_id": pId,
		})
	}

	abc.AddUserLog(uid, "Create Deposit", u.Email, abc.FormatNow(), c.ClientIP(), payName)
	c.JSON(200, r.Response())

}

func UsdtToPay(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	payName := c.PostForm("pay_name")
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	cashId := abc.ToInt(c.PostForm("cashId"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if amount < 200 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10059]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	if payName != "USDT" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10053]

		c.JSON(200, r.Response())

		return
	}

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("deposits", u.Path) || abc.FindActivityDisableOne("CloseUSDT", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10062]

		c.JSON(200, r.Response())

		return
	}

	//如果存在未支付订单
	p := abc.QueryUnpaidOrders(uid)
	if p.Id > 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10058]

		c.JSON(200, r.Response())

		return
	}

	var uc abc.UserVipCash
	if cashId != 0 {
		uc = abc.CheckCouponExist(uid, cashId)

		if uc.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10060]

			c.JSON(200, r.Response())

			return
		}

		if amount < abc.ToFloat64(uc.PayAmount) {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10061]

			c.JSON(200, r.Response())

			return
		}
	}

	if !strings.Contains(u.UserType, "Level") {
		b := abc.GetAuditedBankOne(uid)

		if b.Id == 0 {
			pa := abc.UserIsSuccessDeposited(uid)

			if pa.Id == 0 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10063]

				c.JSON(200, r.Response())

				return
			}
		}
	}

	orderNo := fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid)
	t := abc.FormatNow()
	pId := abc.CreatePayment(orderNo, t, payName, "", "", u.Path, "", "", amount-float64(uc.DeductionAmount), abc.ToFloat64(uc.DeductionAmount), 1, uid, 1, 0)

	if pId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v创建订单失败", "usdt存款", uid))
		return
	}

	if uc.Id > 0 {
		abc.UpdateSql("user_vip_cash", fmt.Sprintf("id = %v", uc.Id), map[string]interface{}{
			"status": 1,
			"pay_id": pId,
		})
	}

	//go func() {
	//	tickerC := time.After(time.Hour * 1)
	//
	//	select {
	//	case <-tickerC:
	//		tron.Do(pId)
	//	}
	//}()
	abc.AddUserLog(uid, "Create Deposit", u.Email, abc.FormatNow(), c.ClientIP(), "USDT")
	telegram.SendMsg(telegram.TEXT, 3, fmt.Sprintf("%s转账待审核,用户：%s, 金额：%.2f", "USDT", u.Username, amount))
	r.Status = 1
	r.Msg = ""
	r.Data = pId
	c.JSON(200, r.Response())
}

func TeleportCallBack(c *gin.Context) {
	var teleportBack TeleportBack
	//if err := c.ShouldBind(&chipBack); err != nil {
	//	c.String(200, "false")
	//	return
	//}
	teleportBack.SysNo = c.PostForm("sys_no")
	teleportBack.Amount = c.PostForm("amount")
	teleportBack.Sign = c.PostForm("sign")
	teleportBack.BillStatus = abc.ToInt(c.PostForm("bill_status"))
	teleportBack.BillNo = c.PostForm("bill_no")

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v接收到的参数:%v", "TeleportCallBack", teleportBack))
	pConfig := abc.GetPaymentConfigOne("Teleport")
	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)
	if callback_key == "" {
		c.String(200, "false")
		return
	}
	//验证签名
	if sign := abc.Md5(fmt.Sprintf("%s%s", teleportBack.BillNo, callback_key)); sign != teleportBack.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TeleportCallBack", teleportBack.BillNo, "签名错误"))
		c.String(200, "false")

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf(`order_no = '%v'`, teleportBack.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TeleportCallBack", teleportBack.BillNo, "订单不存在"))
		c.String(200, "false")

		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	amount := abc.ToFloat64(teleportBack.Amount)

	if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
		abc.RefundCoupon(p.Id, 1)
	}

	state, _ := payNotice.SuccessPay(abc.ToString(teleportBack.BillNo), amount/p.ExchangeRate)

	//u := abc.GetUserById(p.UserId)
	//u1 := abc.GetUserById(u.ParentId)
	//
	////发送邮件
	//mail := abc.MailContent(6)
	//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//
	//brevo.Send(mail.Title, content, u.Email)
	//
	////用户存款给代理发邮件
	//mail2 := abc.MailContent(71)
	//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//brevo.Send(mail.Content, content2, u1.Email)
	//
	////发送站内信
	//message := abc.GetMessageConfig(19)
	//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
	//
	//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))
	go func() {
		payNotice.SuccessPayNotice(p.Id, state, amount)
	}()

	if state == 0 {
		c.String(200, "false")

		return
	}

	c.String(200, "true")
}

func GetPaymentAmount(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	t := c.PostForm("type")
	payName := c.PostForm("pay_name")
	start := c.PostForm("start_time")
	end := c.PostForm("end_time")
	status := c.PostForm("status")

	where := fmt.Sprintf("user_id=%d", uid)
	if t == "deposit" || t == "withdraw" || t == "transfer" || t == "commission" {
		where += fmt.Sprintf(" and type='%s'", t)
	}
	if payName != "" && !strings.Contains(payName, "'") {
		where += fmt.Sprintf(" and pay_name='%s'", payName)
	}
	if start != "" && end != "" {
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", start, end)
		}
	}
	if status != "" {
		if status == "-2" {
			where += fmt.Sprintf(" and status in (-2,-3)")
		} else {
			where += fmt.Sprintf(" and status=%d", abc.ToInt(status))
		}
	}
	r := R{}
	r.Status, r.Data = 1, abc.GetPaymentAmount(where)
	c.JSON(200, r.Response())
}

func ChipPayCallBack(c *gin.Context) {
	var tel ChipBack

	if err := c.BindJSON(&tel); err != nil {
		c.JSON(200, map[string]interface{}{
			"code": 200,
			"msg":  "success",
			"data": map[string]interface{}{
				"otcOrderNum":     tel.IntentOrderNo,
				"companyOrderNum": tel.CompanyOrderNum,
			},
			"success": true,
		})

		return
	}

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "ChipPayCallBack", tel))

	//buildUrl := fmt.Sprintf("coinAmount=%s&coinSign=%s&companyOrderNum=%s&intentOrderNo=%s&successAmount=%s&total=%s&tradeOrderTime=%s&tradeStatus=%s&unitPrice=%s",
	//	tel.CoinAmount, tel.CoinSign, tel.CompanyOrderNum, tel.IntentOrderNo, tel.SuccessAmount, tel.Total, tel.TradeOrderTime, tel.TradeStatus, tel.UnitPrice,
	//)

	buildUrl := ""

	if len(tel.CoinAmount) > 0 {
		buildUrl += "coinAmount=" + tel.CoinAmount + "&"
	}

	buildUrl += fmt.Sprintf("coinSign=%s&companyOrderNum=%s&intentOrderNo=%s&successAmount=%s&",
		tel.CoinSign, tel.CompanyOrderNum, tel.IntentOrderNo, tel.SuccessAmount,
	)

	if len(tel.Total) > 0 {
		buildUrl += "total=" + tel.Total + "&"
	}

	buildUrl += fmt.Sprintf("tradeOrderTime=%s&tradeStatus=%s", tel.TradeOrderTime, tel.TradeStatus)

	if len(tel.UnitPrice) > 0 {
		buildUrl += "&unitPrice=" + tel.UnitPrice
	}

	pConfig := abc.GetPaymentConfigOne("ChipPay")
	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		c.JSON(200, map[string]interface{}{
			"code": 200,
			"msg":  "success",
			"data": map[string]interface{}{
				"otcOrderNum":     tel.IntentOrderNo,
				"companyOrderNum": tel.CompanyOrderNum,
			},
			"success": true,
		})
		return
	}

	if err := abc.VerySignWithSha256RSA(buildUrl, tel.Sign, callback_key); err != nil {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "ChipPayCallBack", tel.CompanyOrderNum, "签名错误"))
		c.JSON(200, map[string]interface{}{
			"code": 200,
			"msg":  "success",
			"data": map[string]interface{}{
				"otcOrderNum":     tel.IntentOrderNo,
				"companyOrderNum": tel.CompanyOrderNum,
			},
			"success": true,
		})

		return
	}

	amount := abc.ToFloat64(tel.CoinAmount)

	if amount < 190 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "ChipPayCallBack", tel.CompanyOrderNum, "存款余额小于190"))
		c.JSON(200, map[string]interface{}{
			"code": 200,
			"msg":  "success",
			"data": map[string]interface{}{
				"otcOrderNum":     tel.IntentOrderNo,
				"companyOrderNum": tel.CompanyOrderNum,
			},
			"success": true,
		})

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", tel.CompanyOrderNum))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "ChipPayCallBack", tel.CompanyOrderNum, "订单不存在"))
		c.JSON(200, map[string]interface{}{
			"code":      200,
			"error_msg": "未查询到订单",
			"msg":       "未查询到订单",
			"success":   false,
			"data":      "",
		})

		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	if tel.TradeStatus == "1" {
		//u := abc.GetUserById(p.UserId)
		//u1 := abc.GetUserById(u.ParentId)
		if amount < p.Amount-1 {
			abc.RefundCoupon(p.Id, 1)
		}

		state, _ := payNotice.SuccessPay(tel.CompanyOrderNum, amount)

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()
		//发送邮件
		//mail := abc.MailContent(6)
		//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//
		//brevo.Send(mail.Title, content, u.Email)
		//
		////用户存款给代理发邮件
		//mail2 := abc.MailContent(71)
		//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//brevo.Send(mail.Content, content2, u1.Email)
		//
		////发送站内信
		//message := abc.GetMessageConfig(19)
		//
		//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, abc.ToString(amount/p.ExchangeRate)), fmt.Sprintf(message.ContentHk, abc.ToString(amount/p.ExchangeRate)), fmt.Sprintf(message.ContentEn, abc.ToString(amount/p.ExchangeRate)))
		//
		//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

		if state == 0 {
			c.JSON(200, map[string]interface{}{
				"code": 200,
				"msg":  "success",
				"data": map[string]interface{}{
					"otcOrderNum":     tel.IntentOrderNo,
					"companyOrderNum": tel.CompanyOrderNum,
				},
				"success": true,
			})

			return
		}
	} else {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)
	}

	c.JSON(200, map[string]interface{}{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"otcOrderNum":     tel.IntentOrderNo,
			"companyOrderNum": tel.CompanyOrderNum,
		},
		"success": true,
	})
}

func AlpapayCallBack(c *gin.Context) {
	var pay Pay7Back
	s, _ := c.GetRawData()
	json.Unmarshal(s, &pay)

	var star Star7Back
	json.Unmarshal(s, &star)

	str, _ := json.Marshal(&star)
	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "AlpapayCallBack", string(s)))
	pConfig := abc.GetPaymentConfigOne("7Star")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("key", pa)

	if callback_key == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,失败原因:%v", "AlpapayCallBack", "回调密钥错误"))
		c.String(200, "回调密钥错误")
		return
	}

	sign := abc.HmacSha256(string(str), callback_key)

	if strings.ToUpper(sign) != pay.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "AlpapayCallBack", pay.OrderNo, "签名错误"))
		c.String(200, "签名错误")
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", pay.OrderNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "AlpapayCallBack", pay.OrderNo, "订单不存在"))
		c.String(200, "订单不存在")
		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "失败")

		return
	}

	defer done()

	if pay.PayStatus == "0" {
		amount := abc.ToFloat64(pay.PayAmt)

		if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
			abc.RefundCoupon(p.Id, 1)
		}

		state, _ := payNotice.SuccessPay(p.OrderNo, amount/p.ExchangeRate)

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()

		c.String(200, "成功")

		return
	}

	if pay.PayStatus == "5" {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)

		c.String(200, "成功")
		return
	}

	c.String(200, "失败")
}

func BftCallback(c *gin.Context) {
	var b BftBack
	s, _ := c.GetRawData()

	json.Unmarshal(s, &b)
	//b.Money = c.PostForm("money")
	//b.TradeID = c.PostForm("tradeId")
	//b.APIOrderNo = c.PostForm("apiOrderNo")
	//b.Signature = c.PostForm("signature")
	//b.UniqueCode = c.PostForm("uniqueCode")
	//b.TradeStatus = c.PostForm("tradeStatus")
	//if err := c.ShouldBindJSON(&b); err != nil {
	//	fmt.Println("===========", err)
	//	c.JSON(200, map[string]interface{}{
	//		"code":    1,
	//		"message": "成功",
	//		"data":    "参数错误",
	//		"success": false,
	//	})
	//	return
	//}

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "BftCallback", b))
	pConfig := abc.GetPaymentConfigOne("BFT")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "参数错误",
			"success": false,
		})
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", b.APIOrderNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "BftCallback", b.APIOrderNo, "订单不存在"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "订单不存在",
			"success": false,
		})
		return
	}

	data := fmt.Sprintf("apiOrderNo=%s&money=%s&tradeId=%s&tradeStatus=1&uniqueCode=%d&key=%s", b.APIOrderNo, b.Money, b.TradeID, p.UserId, callback_key)

	if abc.Md5(data) != b.Signature {
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "签名验证不通过",
			"success": false,
		})

		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	sign := fmt.Sprintf("apiOrderNo=%s&money=%s&tradeId=%s&tradeStatus=1&uniqueCode=%d&key=%s", b.APIOrderNo, b.Money, b.TradeID, p.UserId, callback_key)

	if b.Signature != abc.Md5(sign) {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "BftCallback", b.APIOrderNo, "签名错误"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "签名验证不通过",
			"success": false,
		})
		return
	}

	if b.TradeStatus == "1" {
		amount := abc.ToFloat64(b.Money)

		if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
			abc.RefundCoupon(p.Id, 1)
		}

		state, _ := payNotice.SuccessPay(b.APIOrderNo, amount/p.ExchangeRate)

		//u := abc.GetUserById(p.UserId)
		//u1 := abc.GetUserById(u.ParentId)
		//
		////发送邮件
		//mail := abc.MailContent(6)
		//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//
		//brevo.Send(mail.Title, content, u.Email)
		//
		////用户存款给代理发邮件
		//mail2 := abc.MailContent(71)
		//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//brevo.Send(mail.Content, content2, u1.Email)
		//
		////发送站内信
		//message := abc.GetMessageConfig(19)
		//
		//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
		//
		//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()

		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "成功",
			"success": true,
		})

		return
	} else {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)

		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "成功",
			"success": true,
		})

		return
	}
}

func NeptuneCancel(c *gin.Context) {
	var n NepCancelBack

	n.BillNo = c.PostForm("bill_no")
	n.Sign = c.PostForm("sign")
	n.SysNo = c.PostForm("sys_no")
	n.BillStatus = abc.ToInt(c.PostForm("bill_status"))
	//data, _ := c.GetRawData()
	//telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf(`%v-%v`, "NeptuneCancel接收到的参数", string(data)))

	//json.Unmarshal(data, &n)

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf(`%v-%v-%v-%v-%v`, "NeptuneCancel", n.BillNo, n.BillStatus, n.Sign, n.SysNo))
	pConfig := abc.GetPaymentConfigOne("Neptune")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneCancel", n.BillNo, "回调密钥错误"))
		c.String(200, "false")
		return
	}

	sign := nep.Sign(map[string]interface{}{
		"bill_no":     n.BillNo,
		"bill_status": n.BillStatus,
		"sys_no":      n.SysNo,
	}, callback_key)

	if sign != n.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneCancel", n.BillNo, "签名错误"))
		c.String(200, "false")
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", n.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneCancel", n.BillNo, "订单不存在"))
		c.String(200, "false")
		return
	}

	abc.UpdateSql("payment", fmt.Sprintf("order_no = '%v'", n.BillNo), map[string]interface{}{
		"status":   -2,
		"a_status": -2,
	})

	abc.RefundCoupon(p.Id, 0)

	c.String(200, "success")
}

func NeptuneSuccess(c *gin.Context) {
	var n NepSuccessBack

	n.Sign = c.PostForm("sign")
	n.SysNo = c.PostForm("sys_no")
	n.BillNo = c.PostForm("bill_no")
	n.Amount = c.PostForm("amount")

	//data, _ := c.GetRawData()
	//json.Unmarshal(data, &n)

	pConfig := abc.GetPaymentConfigOne("Neptune")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)
	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("%v-%v-%v-%v-%v", "NeptuneSuccess", n.Sign, n.SysNo, n.BillNo, n.Amount))
	if callback_key == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneSuccess", n.BillNo, "回调密钥错误"))
		c.String(200, "false")
		return
	}

	if abc.Md5(n.BillNo+callback_key) != n.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneSuccess", n.BillNo, "签名错误"))
		c.String(200, "false")
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", n.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "NeptuneSuccess", n.BillNo, "订单不存在"))
		c.String(200, "false")
		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	amount := abc.ToFloat64(n.Amount)

	if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
		abc.RefundCoupon(p.Id, 1)
	}

	state, _ := payNotice.SuccessPay(n.BillNo, amount/p.ExchangeRate)

	//u := abc.GetUserById(p.UserId)
	//u1 := abc.GetUserById(u.ParentId)
	//
	////发送邮件
	//mail := abc.MailContent(6)
	//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//
	//brevo.Send(mail.Title, content, u.Email)
	//
	////用户存款给代理发邮件
	//mail2 := abc.MailContent(71)
	//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//brevo.Send(mail.Content, content2, u1.Email)
	//
	////发送站内信
	//message := abc.GetMessageConfig(19)
	//
	//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
	//
	//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

	go func() {
		payNotice.SuccessPayNotice(p.Id, state, amount)
	}()

	if state == 0 {
		c.JSON(200, map[string]interface{}{
			"code":    200,
			"message": "success",
			"data":    "更新失败",
			"success": true,
		})
		return
	}

	c.String(200, "success")
}

func Help2PayDeposit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	payName := c.PostForm("pay_name")
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	cashId := abc.ToInt(c.PostForm("cashId"))
	currency := c.PostForm("currency")
	bank := c.PostForm("bank")

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if currency == "" || bank == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	comment := currency + " - " + bank
	//if amount < 200 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10059]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	u := abc.GetUserById(uid)

	if payName != "help2pay" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10053]

		c.JSON(200, r.Response())

		return
	}

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("deposits", u.Path) || abc.FindActivityDisableOne("Closehelp2pay", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10062]

		c.JSON(200, r.Response())

		return
	}

	//if u.CoinStatus == 0 {
	//	pa := abc.UserIsSuccessDeposited(uid)
	//
	//	if pa.Id == 0 {
	//		r.Status = 0
	//		r.Msg = golbal.Wrong[language][10063]
	//
	//		c.JSON(200, r.Response())
	//
	//		return
	//	}
	//}

	//如果存在未支付订单
	//p := abc.QueryUnpaidOrders(uid)
	//if p.Id > 0 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10058]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	var uc abc.UserVipCash
	if cashId != 0 {
		uc = abc.CheckCouponExist(uid, cashId)

		if uc.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10060]

			c.JSON(200, r.Response())

			return
		}

		if amount < abc.ToFloat64(uc.PayAmount) {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10061]

			c.JSON(200, r.Response())

			return
		}
	}

	orderNo := fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid)
	t := abc.FormatNow()
	rate := 0.0
	r.Status, r.Msg, r.Data, rate = h2pay.H2Pay(currency, bank, orderNo, t, uid, amount-float64(uc.DeductionAmount), c.ClientIP())

	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10078]
		r.Data = nil

		c.JSON(200, r.Response())
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v下单访问第三方失败", "ChipPay", uid))
		return
	}

	pId := abc.CreatePayment(orderNo, t, payName, comment, "", u.Path, "", "", amount-float64(uc.DeductionAmount), float64(uc.DeductionAmount), rate, uid, 1, 1)

	if pId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v创建订单失败", "Help2Pay", uid))
		return
	}

	if uc.Id > 0 {
		abc.UpdateSql("user_vip_cash", fmt.Sprintf("id = %v", uc.Id), map[string]interface{}{
			"status": 1,
			"pay_id": pId,
		})
	}

	c.JSON(200, r.Response())
}

func Help2payCallBack(c *gin.Context) {
	var h Help2PayBack

	if err := c.ShouldBind(&h); err != nil {
		c.JSON(200, map[string]interface{}{
			"code":      1,
			"error_msg": "参数错误",
			"msg":       "参数错误",
			"data":      "",
			"success":   false,
		})
		return
	}

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "Help2payCallBack", h))

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", h.Reference))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	if h.Status == "000" || h.Status == "006" {
		amount := abc.ToFloat64(h.Amount)
		rate := 1.0

		if p.Status == 1 {
			c.JSON(200, map[string]interface{}{
				"code":      1,
				"error_msg": "",
				"success":   true,
				"msg":       "成功",
				"data":      nil,
			})

			return
		}

		if p.Id == 0 {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "Help2payCallBack", h.Reference, "订单不存在"))
			c.JSON(200, map[string]interface{}{
				"code":      1,
				"error_msg": "订单不存在",
				"msg":       "订单不存在",
				"data":      "",
				"success":   false,
			})
			return
		}

		if p.ExchangeRate > 0 {
			rate = p.ExchangeRate
		}

		if amount/rate < p.Amount-1 {
			abc.RefundCoupon(p.Id, 1)
		}

		state, _ := payNotice.SuccessPay(h.Reference, amount/rate)

		//u := abc.GetUserById(p.UserId)
		//u1 := abc.GetUserById(u.ParentId)
		//
		////发送邮件
		//mail := abc.MailContent(6)
		//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//
		//brevo.Send(mail.Title, content, u.Email)
		//
		////用户存款给代理发邮件
		//mail2 := abc.MailContent(71)
		//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//brevo.Send(mail.Content, content2, u1.Email)
		//
		////发送站内信
		//message := abc.GetMessageConfig(19)
		//
		//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
		//
		//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()
	} else {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)
	}

	c.JSON(200, map[string]interface{}{
		"code":      1,
		"error_msg": "",
		"success":   true,
		"msg":       "成功",
		"data":      nil,
	})
}

func Help2PayRate(c *gin.Context) {
	r := &R{}
	currency := c.PostForm("currency")
	language := abc.ToString(c.MustGet("language"))

	arr := []string{"MYR", "THB", "VND", "IDR"}

	flag := false

	for _, v := range arr {
		if v == currency {
			flag = true
			break
		}
	}

	if !flag {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	r.Status = 1
	r.Msg = ""
	r.Data = h2pay.GetRate(currency)

	c.JSON(200, r.Response())
}

func Help2Bank(c *gin.Context) {
	r := &R{}

	r.Status = 1
	r.Msg = ""
	r.Data = h2pay.GetBank()

	c.JSON(200, r.Response())
}

func GetWallet(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))

	u := abc.GetUserById(uid)

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())
	}

	if !strings.Contains(u.UserType, "Level") {
		b := abc.GetAuditedBankOne(uid)

		if b.Id == 0 {
			pa := abc.UserIsSuccessDeposited(uid)

			if pa.Id == 0 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10063]

				c.JSON(200, r.Response())

				return
			}
		}
	}

	address := ""
	wallet, b := abc.GetWallet(uid)

	if b {
		address = wallet.Address

		r.Status = 1
		r.Msg = ""
		r.Data = address

		c.JSON(200, r.Response())

		return
	}

	count := abc.GetResidueWalletCount()

	if count >= 15 {
		w, _ := abc.GetWallet(0)
		if w.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10502]

			c.JSON(200, r.Response())

			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("数据库usdt钱包个数不足"))
			return
		}

		abc.UpdateSql("user_wallet", fmt.Sprintf("id = %v", w.Id), map[string]interface{}{
			"user_id": uid,
		})

		address = w.Address

		r.Status = 1
		r.Msg = ""
		r.Data = address

		c.JSON(200, r.Response())

		return
	}

	go tron.CreateUserWallet(abc.LimiterPer(wallet2.Queue, "user"))

	done := abc.LimiterPer(wallet2.Queue, "user")

	defer done()
	w, _ := abc.GetWallet(0)
	if w.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10502]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("数据库usdt钱包个数不足"))
		return
	}

	abc.UpdateSql("user_wallet", fmt.Sprintf("id = %v", w.Id), map[string]interface{}{
		"user_id": uid,
	})

	address = w.Address

	r.Status = 1
	r.Msg = ""
	r.Data = address

	c.JSON(200, r.Response())
}

func GetMyBank(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))

	r.Status = 1
	r.Msg = ""
	r.Data = abc.GetMyBank(uid)

	c.JSON(200, r.Response())
}

func WithdrawalTimes(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	count := 0
	u := abc.GetUserById(uid)
	if abc.IsExperienceAccount(uid) {
		uv := abc.GetUserVip(uid)
		times := map[int]int{0: 1, 1: 1, 2: 2, 3: 3, 4: 5, 5: 99}
		withdrawCount := abc.NumberOfWithdrawals(uid)

		if uv.Grade >= 5 {
			count = times[uv.Grade]
		} else {
			count = times[uv.Grade] - int(withdrawCount)
		}

		if count < 0 {
			count = 0
		}
	}

	if u.UserType == "sales" {
		count = 1
	}

	r.Status = 1
	r.Msg = ""
	r.Data = count

	c.JSON(200, r.Response())
}

func Withdrawal(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	code := c.PostForm("code")
	bankId := abc.ToInt(c.PostForm("bank_id"))
	payName := c.PostForm("pay_name")
	walletId := abc.ToInt(c.PostForm("wallet_id"))
	currency := c.PostForm("currency")
	bank := c.PostForm("bank")
	phoneCode := c.PostForm("phone_code")

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	u := abc.GetUserById(uid)

	if amount < 100 && u.UserType != "sales" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10080]

		c.JSON(200, r.Response())

		return
	}

	if payName != "Wire" && payName != "USDT" && payName != "UnionPay" && payName != "Help2pay" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	//white := abc.FindActivityDisableOne("OpenUSDTWithdraw", u.Path)

	payemnt := abc.GetPaymentOne(fmt.Sprintf("type = 'deposit' AND `status` = 1 AND user_id = %v", uid))

	if u.UserType != "sales" && u.SalesType != "admin" {
		if payemnt.Id != 0 {
			if abc.IsUserWhiteHouse(uid) {
				if !abc.DepositTimeJudgment(uid, u.Path) {
					r.Status = 0
					r.Msg = golbal.Wrong[language][10528]

					c.JSON(200, r.Response())

					return
				}
			}

			if payName == "USDT" && abc.FinalDepositUsdt(uid, !abc.IsUserWhiteHouse(uid)) {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10513]

				c.JSON(200, r.Response())

				return
			}
		}
	}

	if payName == "USDT" && walletId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10080]

		c.JSON(200, r.Response())

		return
	}

	if payName == "Wire" && amount < 200 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10086]

		c.JSON(200, r.Response())

		return
	}

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("withdraw"+payName, u.Path) || abc.FindActivityDisableOne("withdraws", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10081]

		c.JSON(200, r.Response())

		return
	}

	var capture abc.Captcha
	if payName != "USDT" {
		capture = abc.VerifySmsCode(u.Mobile, code)

		if capture.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10033]

			c.JSON(200, r.Response())

			return
		}
	} else {
		capture = abc.VerifySmsCode(u.Mobile, phoneCode)

		if capture.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10033]

			c.JSON(200, r.Response())

			return
		}

		g := abc.GetUserG2(uid)

		if g.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10097]

			c.JSON(200, r.Response())

			return
		}

		myCode, _ := validator.GetCode(g.Secret)

		if myCode != code {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10098]

			c.JSON(200, r.Response())

			return
		}
	}

	accountList := abc.GetMyAccount(uid)

	for _, v := range accountList {
		var data abc.AccountInfoData

		m := map[string]interface{}{
			"login": fmt.Sprintf("%d", v.Login),
		}
		res := mt4.CrmPost("api/account_info", m)
		str := abc.ToJSON(res)
		json.Unmarshal(str, &data)
		if data.Code == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10095]

			c.JSON(200, r.Response())

			return
		}
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Credit = fmt.Sprintf("%.2f", data.Credit)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.MarginLevel = fmt.Sprintf("%.2f", data.MarginLevel)
		data.Margin = fmt.Sprintf("%.2f", abc.ToFloat64(data.Margin))
		data.MarginFree = fmt.Sprintf("%.2f", data.MarginFree)
		data.Volume = fmt.Sprintf("%.2f", data.Volume)

		if abc.ToFloat64(data.Balance) < 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10520]

			c.JSON(200, r.Response())

			return
		}
	}
	//if u.CoinStatus == 0 && u.UserType == "user" {
	//	p := abc.UserIsSuccessWithdraw(uid)
	//
	//	if p.Id == 0 && payName == "USDT" {
	//		r.Status = 0
	//		r.Msg = golbal.Wrong[language][10082]
	//
	//		c.JSON(200, r.Response())
	//
	//		return
	//	}
	//}

	paymentAddr := ""

	if payName == "USDT" {
		w := abc.GetAuditedWallet(uid, walletId)
		paymentAddr = w.AddressType + " - " + w.Address
	} else {
		b := abc.GetAuditedBank(bankId, uid)

		if b.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10087]

			c.JSON(200, r.Response())

			return
		}

		paymentAddr = fmt.Sprintf("%s %s %s %s", b.BankNo, b.BankName, b.BankAddress, b.TrueName)

		if len(b.Swift) > 0 {
			paymentAddr += " swift:" + b.Swift
		}
		if len(b.Iban) > 0 {
			paymentAddr += " iban:" + b.Iban
		}
	}

	comment := ""
	if payName == "Help2pay" {
		comment = currency + " - " + bank
	}

	flag := abc.IsExperienceAccount(uid)

	if !flag {
		if amount > 100 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10084]

			c.JSON(200, r.Response())

			return
		}

		p := abc.GetPaymentOne(fmt.Sprintf("status >= 0 and create_time >= '%v' and type = 'withdraw' and user_id = %d", time.Now().Format("2006-01"), uid))

		if p.Id > 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10084]

			c.JSON(200, r.Response())

			return
		}
	}

	if amount > u.WalletBalance {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10085]

		c.JSON(200, r.Response())

		return
	}

	fee := 0.0
	isFee := false

	if !flag || payName == "USDT" {
		isFee = true
	}

	if payName == "UnionPay" || payName == "Help2pay" {
		vip := map[int]int{
			1: 1,
			2: 2,
			3: 3,
			4: 5,
			5: 999,
		}

		uv := abc.GetUserVip(uid)
		count := abc.NumberWithdrawalsMonth(uid)

		if uv.Grade == 5 {
			isFee = false
		} else {
			if int(count) >= vip[uv.Grade] {
				isFee = true
			}
		}
	}

	if isFee {
		res := abc.WithdrawalConfiguration()
		withdraw_fee := 0.0
		withdraw_max := 0.0
		withdraw_min := 0.0
		withdraw_min_usdt := 0.0
		if res != nil {
			arr := strings.Split(abc.PtoString(res, "temp"), ",")
			withdraw_fee = abc.ToFloat64(arr[1])
			withdraw_max = abc.ToFloat64(arr[2])
			withdraw_min = abc.ToFloat64(arr[3])
			withdraw_min_usdt = abc.ToFloat64(arr[4])
		}
		fee = amount * withdraw_fee

		if fee < withdraw_min {
			fee = withdraw_min
		}

		if fee > withdraw_max {
			fee = withdraw_max
		}

		if payName == "USDT" {
			fee = withdraw_min_usdt
		}
	}

	if payName == "Wire" {
		//p := abc.GetPaymentOne(fmt.Sprintf("status >= 0 and user_id = %d and create_time >='%s' and type = 'withdraw' and pay_name='Wire'", uid, time.Now().Format("2006-01")))
		//
		//if p.Id > 0 {
		//	fee = 50
		//}
		fee = abc.ToFloat64(fmt.Sprintf("%.2f", amount*0.01))

		if !abc.IsUserWhiteHouse(uid) {
			fee = 0
		}
	}

	//if payName == "USDT" && white {
	//	//p := abc.GetPaymentOne(fmt.Sprintf("status >= 0 and user_id = %d and create_time >='%s' and type = 'withdraw' and pay_name='USDT'", uid, time.Now().Format("2006-01-02")))
	//	//
	//	//if p.Id == 0 || white {
	//	//	fee = 0
	//	//}
	//	fee = 0
	//}

	p2 := abc.GetPaymentOne(fmt.Sprintf("status >= 0 and user_id = %d and create_time >='%s' and type = 'withdraw' and pay_name='USDT'", uid, time.Now().Format("2006-01")))

	if payName == "USDT" && p2.Id == 0 {
		fee = 0
	}

	transferLogin := 0
	if u.UserType == "sales" {
		fee = 0
		transferLogin = -2
	}

	newAmount := amount - fee

	orderNo := fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid)

	n, p := abc.CreateWithdrawPayment(orderNo, abc.FormatNow(), payName, "withdraw", paymentAddr, u.Path, comment, uid, 1, 0, transferLogin, newAmount*-1, 0, fee*-1)
	if n == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())

		return
	}

	abc.DeleteSmsOrMail(capture.Id)

	message := abc.GetMessageConfig(20)

	abc.SendMessage(uid, 20, fmt.Sprintf(message.ContentZh), fmt.Sprintf(message.ContentHk), fmt.Sprintf(message.ContentEn))

	mail := abc.MailContent(9)

	brevo.Send(mail.Title, fmt.Sprintf(mail.Content, u.TrueName, abc.ToString(amount)), u.Email)

	//m := make(map[string][]string)
	//m["Content-Type"] = []string{"application/json"}
	//go func() {
	//abc.SendRequest("GET", fmt.Sprintf("%v/auto_withdraw/action?id=%v", conf.WithdrawalAddress, p.Id), strings.NewReader(""), m)
	//}()

	abc.AddUserLog(uid, "Create Withdraw", u.Email, abc.FormatNow(), c.ClientIP(), payName)
	r.Status = 1
	r.Msg = ""
	r.Data = p.Id

	go WithdrawalRestrictions2(p.Id, uid)

	c.JSON(200, r.Response())
}

func Transfer(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	account := abc.ToInt(c.PostForm("account"))
	transferAccount := abc.ToInt(c.PostForm("transfer_account"))
	cType := abc.ToInt(c.PostForm("type"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if account == 0 || transferAccount == 0 || amount == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if account == transferAccount {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if account != 1 && transferAccount != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if account == 1 && amount < 10 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10088]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("transfer", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10089]

		c.JSON(200, r.Response())

		return
	}

	if account > 1 {
		a := abc.GetAccountOne(fmt.Sprintf("user_id = %v and login = %v", uid, account))

		if a.Enable != 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10526]

			c.JSON(200, r.Response())

			return
		}

		if a.Login <= 0 || a.IsMam == 1 || a.Enable != 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10090]

			c.JSON(200, r.Response())

			return
		}
		//体验账户一个月转账金额不能大于100美金
		if a.Experience == 1 {
			if amount > 100 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10084]

				c.JSON(200, r.Response())

				return
			}

			money := abc.TransferTotalAmount(uid)
			if money >= 100 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10084]

				c.JSON(200, r.Response())

				return
			}
		}
	}

	if transferAccount > 1 {
		a := abc.GetAccountOne(fmt.Sprintf("user_id = %v and login = %v", uid, transferAccount))

		if a.Enable != 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10526]

			c.JSON(200, r.Response())

			return
		}

		if a.Login <= 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10090]

			c.JSON(200, r.Response())

			return
		}
	}

	equity := 0.0
	credit := 0.0

	if account > 1 {
		res := mt4.CrmPost("api/account_info", map[string]interface{}{
			"login": account,
		})

		if res == nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10095]

			c.JSON(200, r.Response())

			return
		}

		code, ok := res["code"]

		if !ok {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10095]

			c.JSON(200, r.Response())

			return
		}

		if abc.ToInt(code) == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10095]

			c.JSON(200, r.Response())

			return
		}

		equity = res["equity"].(float64)
		balance := res["balance"].(float64)
		credit = res["credit"].(float64)
		margin := res["margin"].(float64)
		data := GetMt4AccountPosition(account)

		if len(data.List) == 0 {
			equity = balance
		} else {
			equity = equity - credit
			if equity < 0.01 || balance < 0.01 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10085]

				c.JSON(200, r.Response())

				return
			}

			if equity > balance {
				equity = balance
			}
			equity = (equity - margin) * 0.9
		}
	} else {
		equity = abc.ToFloat64(u.WalletBalance)
	}

	if amount > equity {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10085]

		c.JSON(200, r.Response())

		return
	}

	transferor := transferAccount
	payee := account
	newAmount := amount
	if account == 1 {
		newAmount = newAmount * -1
		transferor = account
		payee = transferAccount
	}

	p := abc.Payment{
		OrderNo:       fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid),
		UserId:        uid,
		Login:         transferor,
		CreateTime:    abc.FormatNow(),
		Amount:        newAmount,
		PayName:       "",
		Status:        -2,
		Type:          "transfer",
		TransferLogin: payee,
		UserPath:      u.Path,
	}

	payment := abc.CreatePaymentTransfer(p)

	if payment.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())

		return
	}

	//开始转账

	if account == 1 && transferAccount > 1 {
		if err := abc.UpdateUserWallet(amount*-1, uid); err != nil {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v钱包转账户修改用户wallet失败", uid))
			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		res := mt4.CrmPost("api/deposit", map[string]interface{}{
			"xlogin":   transferAccount,
			"xbalance": amount,
			"xcomment": "transfer",
		})

		if res == nil {
			if err := abc.UpdateUserWallet(amount, uid); err != nil {
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v钱包转账户mt4增加金额失败返回wallet失败", uid))
			}

			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		code, ok := res["code"]

		if !ok {
			if err := abc.UpdateUserWallet(amount, uid); err != nil {
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v钱包转账户mt4增加金额失败返回wallet失败", uid))
			}

			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		if abc.ToInt(code) != 1 {
			if err := abc.UpdateUserWallet(amount, uid); err != nil {
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v钱包转账户mt4增加金额失败返回wallet失败", uid))
			}

			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		if err := abc.UpdatePaymentStatus(payment.Id); err != nil {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单:%v修改订单状态失败", p.Id))

			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}
	} else {
		res := mt4.CrmPost("api/deposit", map[string]interface{}{
			"xlogin":   payee,
			"xbalance": amount * -1,
			"xcomment": "transfer",
		})

		if res == nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		code, ok := res["code"]

		if !ok {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		if abc.ToInt(code) != 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		if err := abc.UpdateUserWallet(amount, uid); err != nil {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("账户:%v钱包转账户修改用户wallet失败", uid))
			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}

		if err := abc.UpdatePaymentStatus(payment.Id); err != nil {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单:%v修改订单状态失败", p.Id))

			r.Status = 0
			r.Msg = golbal.Wrong[language][10093]

			c.JSON(200, r.Response())

			return
		}
	}

	//res := mt4.CrmPost("api/deposit", map[string]interface{}{
	//	"xlogin":   payee,
	//	"xbalance": amount * -1,
	//	"xcomment": "transfer",
	//})
	//
	//if res == nil {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10093]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}
	//
	//code, ok := res["code"]
	//
	//if !ok {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10093]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	//tx := abc.Tx()
	//err := abc.UpdatePaymentStatus(tx, payment.Id)
	//err = abc.UpdateUserWallet(tx, amount, payment.UserId)
	//
	//fmt.Println("======err====", err)
	//if err != nil {
	//	tx.Rollback()
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10094]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}
	//
	//tx.Commit()

	//入金到MT4，判断 如果是DMA账户 且账户余额大于 10000 需要激活
	a := abc.GetAccountOne(fmt.Sprintf("login = %v", transferAccount))

	if a.Login > 0 && (a.RebateId > 4 && a.RebateId < 9) && a.ReadOnly == 1 {
		//统计总的转账金额
		totalAmount := abc.StatisticalTransferAmount(transferAccount)
		config := abc.GetConf(7)

		DmaTime := abc.GetConf(10)
		value := abc.ToFloat64(config.Value)

		startTime, endTime := "", ""

		arr := strings.Split(DmaTime.Value, " ")
		if len(arr) == 2 {
			startTime = arr[0]
			endTime = arr[1]
		}

		t, _ := time.Parse("2006-01-02 15:04:05", u.CreateTime)
		t1 := t.Format("2006-01-02")

		if totalAmount*-1 >= 9999 || (u.InviteCode != "TlJwgx" && totalAmount*-1 >= value && t1 >= startTime && t1 <= endTime)  {
			res1 := mt4.CrmPost("api/read_only", map[string]interface{}{
				"xlogin": transferAccount,
				"xvalue": "0",
			})

			code1 := res1["code"].(float64)

			if code1 == 1 {
				abc.UpdateAccountStatus(fmt.Sprintf("login = %v", transferAccount), map[string]interface{}{
					"read_only": 0,
				})

				//发送站内信
				message := abc.GetMessageConfig(24)
				abc.SendMessage(a.UserId, 24, fmt.Sprintf(message.ContentZh, abc.ToString(transferAccount)), fmt.Sprintf(message.ContentHk, abc.ToString(transferAccount)), fmt.Sprintf(message.ContentEn, abc.ToString(transferAccount)))

				//发送邮件
				mail := abc.MailContent(18)
				content := fmt.Sprintf(mail.Content, u.TrueName, transferAccount)
				brevo.Send(mail.Title, content, u.Email)

				if u.InviteCode == "SIVYVT" {
					coupon := abc.GetCouponOne(fmt.Sprintf("user_id=%d and type=2", u.Id))
					if coupon.Id == 0 {
						couponNo := abc.RandStr(10)
						abc.Coupon{
							Type:          2,
							UserId:        uid,
							CouponNo:      couponNo,
							Amount:        0,
							Status:        1,
							Comment:       "DMA激活",
							CreateTime:    abc.FormatNow(),
							UsedStartTime: abc.FormatNow(),
							UsedEndTime:   "2024-04-21 23:59:59",
							Login:         abc.ToString(transferAccount),
							Credit:        0,
						}.CreateCoupon()
						message = abc.GetMessageConfig(114)
						abc.SendMessage(a.UserId, 114, message.ContentZh, message.ContentHk, message.ContentEn)
						telegram.SendMsg(telegram.TEXT, 20,
							fmt.Sprintf("DMA激活，用户：%s，劵号：%s，转账金额：%f，时间：%s", u.TrueName, couponNo, amount, abc.FormatNow()))
					}
				}
			}
		}
	}
	message := abc.GetMessageConfig(23)
	abc.SendMessage(uid, 23, message.ContentZh, message.ContentHk, message.ContentEn)

	go func() {
		mail := abc.MailContent(12)
		conetnt := fmt.Sprintf(mail.Content, abc.ToString(u.TrueName), abc.ToString(payee), abc.ToString(newAmount))
		brevo.Send(mail.Title, conetnt, u.Email)
	}()

	if cType == 1 {
		messageConfig := abc.GetMessageConfig(117)
		abc.SendMessage(uid, 117, fmt.Sprintf(messageConfig.ContentZh, abc.FormatNow(), amount, transferAccount),
			fmt.Sprintf(messageConfig.ContentHk, abc.FormatNow(), amount, transferAccount), fmt.Sprintf(messageConfig.ContentEn, abc.FormatNow(), amount, transferAccount))
	}

	abc.AddUserLog(uid, "Transfer", u.Email, abc.FormatNow(), c.ClientIP(), fmt.Sprintf("%v ==> %v", account, transferAccount))
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func WireDeposit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	file := HandleFilesAllFiles(c, uid, "upload", "file")
	waterNumber := c.PostForm("water_number")
	cashId := abc.ToInt(c.PostForm("cashId"))

	newFile := ""
	if len(file) != 0 {
		newFile = file[0]
	}
	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if amount < 200 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10059]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	if u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	if abc.FindActivityDisableOne("deposits", u.Path) || abc.FindActivityDisableOne("CloseWire", u.Path) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10062]

		c.JSON(200, r.Response())

		return
	}

	//p := abc.QueryUnpaidOrders(uid)
	//if p.Id > 0 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10058]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	var uc abc.UserVipCash
	if cashId != 0 {
		uc = abc.CheckCouponExist(uid, cashId)

		if uc.Id == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10060]

			c.JSON(200, r.Response())

			return
		}

		if amount < abc.ToFloat64(uc.PayAmount) {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10061]

			c.JSON(200, r.Response())

			return
		}
	}

	orderNo := fmt.Sprintf("%d_%d", time.Now().UnixMicro(), uid)

	pId := abc.CreatePayment(orderNo, abc.FormatNow(), "Wire", "", newFile, u.Path, "", waterNumber, amount-float64(uc.DeductionAmount), float64(uc.DeductionAmount), 1, uid, 1, 1)

	if pId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10078]

		c.JSON(200, r.Response())
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("存款渠道名称:%v,用户id:%v创建订单失败", "wire", uid))
		return
	}

	if uc.Id > 0 {
		abc.UpdateSql("user_vip_cash", fmt.Sprintf("id = %v", uc.Id), map[string]interface{}{
			"status": 1,
			"pay_id": pId,
		})
	}

	telegram.SendMsg(telegram.TEXT, 3, fmt.Sprintf("%s转账待审核,用户：%s, 金额：%.2f", "Wire", u.Username, amount))
	telegram.SendMsg(telegram.TEXT, 18, fmt.Sprintf("%s转账待审核,用户：%s, 金额：%.2f", "Wire", u.Username, amount))
	r.Status = 1
	r.Msg = ""
	r.Data = pId
	abc.AddUserLog(uid, "Create Deposit", u.Email, abc.FormatNow(), c.ClientIP(), "Wire")
	c.JSON(200, r.Response())
}

func CancelPayment(c *gin.Context) {
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

	pay := abc.GetPaymentOne(fmt.Sprintf("id=%d and user_id=%d", id, uid))
	if pay.Id == 0 || pay.Type != "withdraw" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	} else if pay.Status != 0 || (pay.AStatus == 1 && pay.BStatus == 1) {
		r.Msg = golbal.Wrong[language][10120]
		c.JSON(200, r.Response())
		return
	}
	//err := abc.UpdateSql("payment", fmt.Sprintf("id=%d", id), map[string]interface{}{
	//	"status": -1,
	//})
	tx := abc.Tx()
	err := tx.Debug().Model(abc.Payment{}).Where(fmt.Sprintf("id=%d", pay.Id)).Updates(map[string]interface{}{
		"status": -1,
	}).Error
	if err != nil {
		tx.Rollback()
		log.Println(" abc CancelPayment1 ", err)
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(200, r.Response())
		return
	}
	if err = tx.Debug().Model(abc.User{}).Where("id=?", uid).Updates(map[string]any{
		"wallet_balance": gorm.Expr(fmt.Sprintf("wallet_balance + (%.2f)", math.Abs(pay.Amount+pay.PayFee))),
	}).Error; err != nil {
		log.Println(" abc CancelPayment2 ", err)
		//telegram.SendMsg(telegram.TEXT, telegram.TEXT,
		//	fmt.Sprintf("用户钱包余额更新失败,UID:%d,Amount:%2f\n", interest.UserId, interest.Fee))
		tx.Rollback()
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(200, r.Response())
		return
	}
	tx.Debug().Create(&abc.PaymentLog{
		PaymentId:  pay.Id,
		CreateTime: abc.FormatNow(),
		Comment:    "用户取消取款",
	})
	tx.Commit()
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}
func ToPay(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	id := abc.ToInt(c.PostForm("id"))

	if id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("user_id = %v and id = %v", uid, id))

	if p.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10092]

		c.JSON(200, r.Response())

		return
	}

	r.Status = 1
	r.Msg = ""
	r.Data = p.PayUrl

	c.JSON(200, r.Response())
}

func UploadWireFile(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	wireDoc := HandleFilesAllFiles(c, uid, "upload", "wire_doc")
	waterNumber := c.PostForm("water_number")
	id := abc.ToInt(c.PostForm("id"))

	if id == 0 || len(wireDoc) == 0 || waterNumber == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("user_id = %v and id = %v", uid, id))

	if p.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10092]

		c.JSON(200, r.Response())

		return
	}

	if err := abc.UpdateSql("payment", fmt.Sprintf("id = %v", id), map[string]interface{}{
		"wire_doc":     wireDoc[0],
		"water_number": waterNumber,
	}); err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10014]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func Remind(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	id := abc.ToInt(c.PostForm("id"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("user_id = %v and id = %v", uid, id))

	if p.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10092]

		c.JSON(200, r.Response())

		return
	}

	if p.PayName == "USDT" {
		tron.Do(p.Id)
	}

	abc.UpdateSql("payment", fmt.Sprintf("id = %v", id), map[string]interface{}{
		"c_status": 1,
	})

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func TeleportCanceled(c *gin.Context) {
	var teleportBack TeleportBack
	//if err := c.ShouldBind(&chipBack); err != nil {
	//	c.JSON(200, map[string]interface{}{
	//		"code":    200,
	//		"err_msg": "数据错误",
	//		"msg":     "数据错误",
	//		"success": false,
	//		"data":    "",
	//	})
	//	return
	//}

	teleportBack.SysNo = c.PostForm("sys_no")
	teleportBack.Amount = c.PostForm("amount")
	teleportBack.Sign = c.PostForm("sign")
	teleportBack.BillStatus = abc.ToInt(c.PostForm("bill_status"))
	teleportBack.BillNo = c.PostForm("bill_no")

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "TeleportCanceled", teleportBack))
	pConfig := abc.GetPaymentConfigOne("Teleport")
	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		c.String(200, "false")
		return
	}

	sign := abc.Md5(fmt.Sprintf("bill_no=%s&bill_status=%s&sys_no=%s%s", teleportBack.BillNo, abc.ToString(teleportBack.BillStatus), teleportBack.SysNo, callback_key))

	if sign != teleportBack.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "ChipPayCanceled", teleportBack.BillNo, "签名错误"))
		c.String(200, "false")

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", teleportBack.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "ChipPayCanceled", teleportBack.BillNo, "订单不存在"))
		c.String(200, "false")

		return
	}

	if err := abc.UpdateSql("payment", fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
		"status":   -2,
		"a_status": -2,
	}); err != nil {
		c.String(200, "false")

		return
	}

	abc.RefundCoupon(p.Id, 0)

	c.String(200, "success")
}

func TodayDeposit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	p := abc.GetPaymentOne(fmt.Sprintf("user_id = %d and status = 1 and pay_time > '%v'", uid, time.Now().Format("2006-01-02")))

	flag := false

	if p.Id > 0 {
		flag = true
	}

	r.Status = 1
	r.Msg = ""
	r.Data = flag

	c.JSON(200, r.Response())
}

func UnipayRate(c *gin.Context) {
	r := &R{}

	m := make(map[string]interface{}, 0)
	pSlice := abc.GetPaymentUnipayRate()
	rate := teleport.GetBaseExchangeRate()

	for _, v := range pSlice {
		m[v.Name] = v.ExchangeRate
		if v.ExchangeRate < 2 {
			m[v.Name], _ = decimal.NewFromFloat(rate * v.ExchangeRate).Round(2).Float64()
		}
	}

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func PaymentAllUserList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	trueName := c.PostForm("true_name")
	email := c.PostForm("email")
	grade := abc.ToInt(c.PostForm("grade"))
	t := c.PostForm("type")
	start := c.PostForm("start")
	end := c.PostForm("end")
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if !(t == "deposit" || t == "withdraw" || t == "transfer" || t == "commission") {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	where := fmt.Sprintf("p.user_path like '%%,%d,%%' and user_id!=%d and type='%s' ", uid, uid, t)
	where += " and p.status=1"
	if trueName != "" && !strings.Contains(trueName, "'") {
		where += fmt.Sprintf(" and u.true_name like '%s%%'", trueName)
	}
	if email != "" && !strings.Contains(email, "'") {
		where += fmt.Sprintf(" and u.email like '%s%%'", email)
	}
	if grade != 0 {
		switch grade {
		case 1:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅰ")
		case 2:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅱ")
		case 3:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅲ")
		case 10:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "user")
		default:

		}
	}
	if start != "" && end != "" {
		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and p.pay_time >= '%s' and p.pay_time<='%s'", start, end)
		}
	}
	r.Status = 1
	m := make(map[string]interface{})
	re := false
	if page == 1 && grade == 0 && start == "" && end == "" {
		re = true
	}
	r.Count, m["list"], m["total"] = abc.GetPaymentSimple(re, uid, page, size, t, where)
	r.Data = m
	//r.Count, r.Data = abc.GetPaymentSimple(page, size, where)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func GetCashList(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res := abc.GetUserVipCash(fmt.Sprintf(`user_id = %v and pay_amount != 0 AND status = 0 AND create_time >= '%v'`, uid, time.Now().AddDate(0, 0, -90).Format("2006-01-02 15:04:05")))

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())

}

func GetWalletList(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	w := abc.GetWalletAddress(uid)

	r.Status = 1
	r.Msg = ""
	r.Data = w

	c.JSON(200, r.Response())
}

func GetAccountBalance(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	account := c.PostForm("account")

	if account == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}
	res := mt4.CrmPost("api/account_info", map[string]interface{}{
		"login": account,
	})

	if res == nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10095]

		c.JSON(200, r.Response())

		return
	}

	code, ok := res["code"]

	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10095]

		c.JSON(200, r.Response())

		return
	}

	if abc.ToInt(code) == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10095]

		c.JSON(200, r.Response())

		return
	}

	balance := res["balance"].(float64)
	equity := res["equity"].(float64)
	credit := res["credit"].(float64)
	margin := res["margin"].(float64)
	data := GetMt4AccountPosition(abc.ToInt(account))

	if len(data.List) == 0 {
		equity = balance
	} else {
		equity = equity - credit
		if equity < 0.01 || balance < 0.01 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10085]

			c.JSON(200, r.Response())

			return
		}

		if equity > balance {
			equity = balance
		}
		equity = (equity - margin) * 0.9
	}
	money := equity

	r.Status = 1
	r.Msg = ""
	r.Data = money

	c.JSON(200, r.Response())
}

func GetPaymentLimit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	u := abc.GetUserById(uid)
	res := abc.PaymentLimit(0, u.Path)

	for _, v := range res {
		if abc.ToInt(abc.PtoString(v, "quick_pay")) == 1 {
			v.(map[string]interface{})["quick_pay"] = abc.PtoString(v, "weight")
		}
		if abc.PtoString(v, "name") == "7Star" {
			v.(map[string]interface{})["flag"] = !abc.FindActivityDisableOne("open"+abc.PtoString(v, "name"), u.Path)
		} else {
			v.(map[string]interface{})["flag"] = abc.FindActivityDisableOne("Close"+abc.PtoString(v, "name"), u.Path)
		}
	}

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func DisablePayChannels(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))

	u := abc.GetUserById(uid)

	m := make(map[string]interface{})
	m["limit"] = abc.PaymentLimit(1, u.Path)
	m["USDT"] = abc.FindActivityDisableOne("CloseUSDT", u.Path)
	m["UnionPay"] = abc.FindActivityDisableOne("CloseUnionPay", u.Path)
	m["Wire"] = abc.FindActivityDisableOne("CloseWire", u.Path)
	m["Help2Pay"] = abc.FindActivityDisableOne("Closehelp2pay", u.Path)

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func GetPathFull(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	userId := abc.ToInt(c.PostForm("user_id"))

	r := &R{}

	user := abc.GetUser(fmt.Sprintf("id=%d", userId))

	if !strings.Contains(user.Path, fmt.Sprintf(",%d,", uid)) {

		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}

	list := abc.GetUsersSimple(fmt.Sprintf("find_in_set(id,'%s')", user.Path),
		fmt.Sprintf("field(id, %s) desc", strings.Trim(user.Path, ",")))

	for i, simple := range list {
		if simple.SalesType == "admin" {
			list[i].SalesType = ""
		}
		if simple.Id == uid {
			list = list[:i+1]
			break
		}
	}

	r.Data = list
	r.Status = 1
	c.JSON(200, r.Response())
}

func EXlinkCallback(c *gin.Context) {
	var e EXlinkBack
	s, _ := c.GetRawData()

	json.Unmarshal(s, &e)
	//b.Money = c.PostForm("money")
	//b.TradeID = c.PostForm("tradeId")
	//b.APIOrderNo = c.PostForm("apiOrderNo")
	//b.Signature = c.PostForm("signature")
	//b.UniqueCode = c.PostForm("uniqueCode")
	//b.TradeStatus = c.PostForm("tradeStatus")
	//if err := c.ShouldBindJSON(&b); err != nil {
	//	fmt.Println("===========", err)
	//	c.JSON(200, map[string]interface{}{
	//		"code":    1,
	//		"message": "成功",
	//		"data":    "参数错误",
	//		"success": false,
	//	})
	//	return
	//}

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v,接收到的参数:%v", "EXlinkCallback", e))
	pConfig := abc.GetPaymentConfigOne("EXlink")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "参数错误",
			"success": false,
		})
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", e.APIOrderNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "EXlinkCallback", e.APIOrderNo, "订单不存在"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "订单不存在",
			"success": false,
		})
		return
	}

	data := fmt.Sprintf("apiOrderNo=%s&money=%s&tradeId=%s&tradeStatus=1&uniqueCode=%d&key=%s", e.APIOrderNo, e.Money, e.TradeID, p.UserId, callback_key)

	if abc.Md5(data) != e.Signature {
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "签名验证不通过",
			"success": false,
		})

		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	sign := fmt.Sprintf("apiOrderNo=%s&money=%s&tradeId=%s&tradeStatus=1&uniqueCode=%d&key=%s", e.APIOrderNo, e.Money, e.TradeID, p.UserId, callback_key)

	if e.Signature != abc.Md5(sign) {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "EXlinkCallback", e.APIOrderNo, "签名错误"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "签名验证不通过",
			"success": false,
		})
		return
	}

	if e.TradeStatus == "1" {
		amount := abc.ToFloat64(e.Money)

		if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
			abc.RefundCoupon(p.Id, 1)
		}

		state, _ := payNotice.SuccessPay(e.APIOrderNo, amount/p.ExchangeRate)

		//u := abc.GetUserById(p.UserId)
		//u1 := abc.GetUserById(u.ParentId)
		//
		////发送邮件
		//mail := abc.MailContent(6)
		//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//
		//brevo.Send(mail.Title, content, u.Email)
		//
		////用户存款给代理发邮件
		//mail2 := abc.MailContent(71)
		//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
		//brevo.Send(mail.Content, content2, u1.Email)
		//
		////发送站内信
		//message := abc.GetMessageConfig(19)
		//
		//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
		//
		//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()

		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "成功",
			"success": true,
		})

		return
	} else {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)

		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "成功",
			"data":    "成功",
			"success": true,
		})

		return
	}
}

func FastportCancel(c *gin.Context) {
	var n FastCancelBack

	n.BillNo = c.PostForm("bill_no")
	n.Sign = c.PostForm("sign")
	n.SysNo = c.PostForm("sys_no")
	n.BillStatus = abc.ToInt(c.PostForm("bill_status"))
	//data, _ := c.GetRawData()
	//telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf(`%v-%v`, "NeptuneCancel接收到的参数", string(data)))

	//json.Unmarshal(data, &n)

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf(`%v-%v-%v-%v-%v`, "FastportCancel", n.BillNo, n.BillStatus, n.Sign, n.SysNo))
	pConfig := abc.GetPaymentConfigOne("Fastport")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)

	if callback_key == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportCancel", n.BillNo, "回调密钥错误"))
		c.String(200, "false")
		return
	}

	sign := yinlian_fastPort.Sign(map[string]interface{}{
		"bill_no":     n.BillNo,
		"bill_status": n.BillStatus,
		"sys_no":      n.SysNo,
	}, callback_key)

	if sign != n.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportCancel", n.BillNo, "签名错误"))
		c.String(200, "false")
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", n.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportCancel", n.BillNo, "订单不存在"))
		c.String(200, "false")
		return
	}

	abc.UpdateSql("payment", fmt.Sprintf("order_no = '%v'", n.BillNo), map[string]interface{}{
		"status":   -2,
		"a_status": -2,
	})

	abc.RefundCoupon(p.Id, 0)

	c.String(200, "success")
}

func FastportSuccess(c *gin.Context) {
	var n FastSuccessBack

	n.Sign = c.PostForm("sign")
	n.SysNo = c.PostForm("sys_no")
	n.BillNo = c.PostForm("bill_no")
	n.Amount = c.PostForm("amount")

	//data, _ := c.GetRawData()
	//json.Unmarshal(data, &n)

	pConfig := abc.GetPaymentConfigOne("Fastport")

	var pa []chip.Parameter
	json.Unmarshal([]byte(pConfig.KeySecret), &pa)

	callback_key := chip.GetValue("callback_key", pa)
	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("%v-%v-%v-%v-%v", "FastportSuccess", n.Sign, n.SysNo, n.BillNo, n.Amount))
	if callback_key == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportSuccess", n.BillNo, "回调密钥错误"))
		c.String(200, "false")
		return
	}

	if abc.Md5(n.BillNo+callback_key) != n.Sign {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportSuccess", n.BillNo, "签名错误"))
		c.String(200, "false")
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("order_no = '%v'", n.BillNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "FastportSuccess", n.BillNo, "订单不存在"))
		c.String(200, "false")
		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		c.String(200, "false")

		return
	}

	defer done()

	amount := abc.ToFloat64(n.Amount)

	if abc.ToFloat64(fmt.Sprintf("%.2f", amount/p.ExchangeRate)) < p.Amount-1 {
		abc.RefundCoupon(p.Id, 1)
	}

	state, msg := payNotice.SuccessPay(n.BillNo, amount/p.ExchangeRate)
	fmt.Println("FastportSuccess state=", state, " msg=", msg)

	//u := abc.GetUserById(p.UserId)
	//u1 := abc.GetUserById(u.ParentId)
	//
	////发送邮件
	//mail := abc.MailContent(6)
	//content := fmt.Sprintf(mail.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//
	//brevo.Send(mail.Title, content, u.Email)
	//
	////用户存款给代理发邮件
	//mail2 := abc.MailContent(71)
	//content2 := fmt.Sprintf(mail2.Content, strconv.Itoa(p.UserId), fmt.Sprintf("%.2f", p.Amount))
	//brevo.Send(mail.Content, content2, u1.Email)
	//
	////发送站内信
	//message := abc.GetMessageConfig(19)
	//
	//abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, amount/p.ExchangeRate), fmt.Sprintf(message.ContentHk, amount/p.ExchangeRate), fmt.Sprintf(message.ContentEn, amount/p.ExchangeRate))
	//
	//telegram.SendMsg(telegram.TEXT, 10, fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f)", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate))

	go func() {
		payNotice.SuccessPayNotice(p.Id, state, amount)
	}()

	if state == 0 {
		c.String(200, "false");
		return
		c.JSON(200, map[string]interface{}{
			"code":    200,
			"message": "success",
			"data":    "更新失败",
			"success": true,
		})
		return
	}

	c.String(200, "success")
}

func CancelDeposit(c *gin.Context) {
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

	pay := abc.GetPaymentOne(fmt.Sprintf("id=%d and user_id=%d and pay_name = 'USDT'", id, uid))
	if pay.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}

	if pay.Status != 0 {
		r.Msg = golbal.Wrong[language][10120]
		c.JSON(200, r.Response())
		return
	}

	abc.CancelDeposit(pay.Id)
	abc.AddUserLog(uid, "Cancel Pay","",abc.FormatNow(),c.ClientIP(),"USDT")
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func TransferJudgment(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	amount := math.Abs(abc.ToFloat64(c.PostForm("amount")))
	account := abc.ToInt(c.PostForm("account"))
	transferAccount := abc.ToInt(c.PostForm("transfer_account"))

	if account == 0 || transferAccount == 0 || amount == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if account == transferAccount {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if account != 1 && transferAccount != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	flag := true
	if account == 1 && transferAccount > 1 {
		a := abc.GetAccountOne(fmt.Sprintf("user_id = %v and login = %v", uid, transferAccount))

		if a.Login <= 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10090]

			c.JSON(200, r.Response())

			return
		}

		var data abc.AccountInfoData

		m := map[string]interface{}{
			"login": fmt.Sprintf("%d", transferAccount),
		}
		res := mt4.CrmPost("api/account_info", m)
		str := abc.ToJSON(res)
		json.Unmarshal(str, &data)
		if data.Code == 0 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10095]

			c.JSON(200, r.Response())

			return
		}
		data.Balance = fmt.Sprintf("%.2f", data.Balance)
		data.Credit = fmt.Sprintf("%.2f", data.Credit)
		data.Equity = fmt.Sprintf("%.2f", data.Equity)
		data.Margin = fmt.Sprintf("%.2f", data.Margin)
		data.MarginLevel = fmt.Sprintf("%.2f", data.MarginLevel)
		data.Margin = fmt.Sprintf("%.2f", abc.ToFloat64(data.Margin))
		data.MarginFree = fmt.Sprintf("%.2f", data.MarginFree)
		data.Volume = fmt.Sprintf("%.2f", data.Volume)

		newAmount := amount + abc.ToFloat64(data.Equity)

		if strings.Contains(a.GroupName, "STD") && newAmount < 50000 {
			if abc.ToInt(data.Leverage) > 800 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "STD") && (newAmount >= 50000 && newAmount < 100000) {
			if abc.ToInt(data.Leverage) > 500 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "STD") && (newAmount >= 100000 && newAmount < 1500000) {
			if abc.ToInt(data.Leverage) > 200 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "STD") && (newAmount >= 1500000 && newAmount < 2000000) {
			if abc.ToInt(data.Leverage) > 100 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "STD") && newAmount >= 200000 {
			if abc.ToInt(data.Leverage) > 50 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "DMA") && newAmount < 50000 {
			if abc.ToInt(data.Leverage) > 400 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "DMA") && (newAmount >= 50000 && newAmount < 100000) {
			if abc.ToInt(data.Leverage) > 200 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "DMA") && (newAmount >= 100000 && newAmount < 1500000) {
			if abc.ToInt(data.Leverage) > 100 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "DMA") && (newAmount >= 1500000 && newAmount < 2000000) {
			if abc.ToInt(data.Leverage) > 50 {
				flag = false
			}
		}

		if strings.Contains(a.GroupName, "DMA") && newAmount >= 200000 {
			if abc.ToInt(data.Leverage) > 25 {
				flag = false
			}
		}
	}

	r.Status = 1
	r.Data = flag
	r.Msg = ""

	c.JSON(200, r.Response())
}



func UserCancelDeposit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	language := abc.ToString(c.MustGet("language"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	if id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("id = %v and user_id = %v", id, uid))

	if p.Id == 0 {
		r.Msg = golbal.Wrong[language][10092]

		c.JSON(200, r.Response())

		return
	}

	if p.Status != 0 {
		r.Msg = golbal.Wrong[language][10120]
		c.JSON(200, r.Response())
		return
	}

	if p.PayFee != 0 {
		abc.ReturnDepositCoupon(p.Id)
	}

	abc.UpdatePayment(fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
		"pay_fee": 0,
		"status":  -1,
	})

	abc.AddUserLog(uid, "Cancel Pay", "", abc.FormatNow(), c.ClientIP(), p.PayName)

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func TopsCallback(c *gin.Context) {
	var tops yinlian_tops.TopsReq

	s, _ := c.GetRawData()

	telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v接收到的参数:%v", "TopsCallback", string(s)))

	if err := json.Unmarshal(s, &tops); err != nil {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TopsCallback", tops.MerchantOrderNo, "解析参数失败"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "success",
			"data":    "解析参数失败",
			"success": true,
		})
		return
	}

	if c.Request.Header.Get("sign") != yinlian_tops.TopsBackSign(s) {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v签名失败", "TopsCallback"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "success",
			"data":    "签名验证不通过",
			"success": true,
		})
		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf(`order_no = '%v'`, tops.MerchantOrderNo))

	if p.Id == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TopsCallback", tops.MerchantOrderNo, "订单不存在"))
		c.JSON(200, map[string]interface{}{
			"code":    1,
			"message": "success",
			"data":    "订单不存在",
			"success": true,
		})
		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, p.UserId)

	if !ok {
		return
	}

	defer done()

	amount := float64(tops.PayAmount)

	if tops.Status == 1 {

		if amount < p.Amount * p.ExchangeRate-1 {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TopsCallback", tops.MerchantOrderNo, "金额不匹配"))
			abc.RefundCoupon(p.Id, 1)

			c.JSON(200, map[string]interface{}{
				"code":    1,
				"message": "success",
				"data":    "金额不匹配",
				"success": true,
			})
			return
		}

		state, _ := payNotice.SuccessPay(abc.ToString(tops.MerchantOrderNo), amount/p.ExchangeRate)

		if state == 0 {
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("回调名:%v订单号:%v失败原因:%v", "TopsCallback", tops.MerchantOrderNo, "更新金额失败"))
			c.JSON(200, map[string]interface{}{
				"code":    1,
				"message": "success",
				"data":    "更新金额失败",
				"success": true,
			})
			return
		}

		go func() {
			payNotice.SuccessPayNotice(p.Id, state, amount)
		}()

	} else {
		abc.UpdateSql("payment", fmt.Sprintf("id = %d", p.Id), map[string]interface{}{
			"status":   -2,
			"a_status": -2,
		})

		abc.RefundCoupon(p.Id, 0)
	}


	c.JSON(200, map[string]interface{}{
		"code":    1,
		"message": "success",
		"data":    tops.MerchantOrderNo,
		"success": true,
	})
}

func WithdrawalRestrictions(c *gin.Context) {
	r := &R{}
	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
	return

	id := abc.ToInt(c.PostForm("id"))
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	u := abc.GetUserById(uid)

	if id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("id = %v and type = 'withdraw' and user_id = %v", id, uid))

	if p.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if u.UserType != "sales" && u.SalesType != "admin" {
		payemnt := abc.GetPaymentOne(fmt.Sprintf("type = 'deposit' AND `status` = 1 AND user_id = %v", uid))

		if abc.IsUserWhiteHouse(uid) {
			if payemnt.Id != 0 {
				if p.PayName != "USDT" {
					flag, need, available := abc.DetermineWithdrawalVolume(uid, math.Abs(p.Amount)+math.Abs(p.PayFee), 1)
					fmt.Println("---")
					fmt.Println(need)
					fmt.Println(available)
					fmt.Println("---")
					if !flag {
						abc.UpdatePayment(fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
							"status":  -3,
							"comment": "Insufficient trading volume",
						})

						abc.UpdateUserWallet(math.Abs(p.Amount+p.PayFee), uid)
						abc.AddPaymentLog(p.Id, "交易量不足")
						r.Status = 0
						r.Msg = fmt.Sprintf(golbal.Wrong[language][10529], need, available)

						c.JSON(200, r.Response())

						return
					}
				} else {
					flag, need, available := abc.DetermineWithdrawalVolume(uid, math.Abs(p.Amount)+math.Abs(p.PayFee), 0)
					fmt.Println(fmt.Sprintf("need=%vava=%v", need, available))
					fmt.Println("===flag=", flag)

					if !flag {
						abc.UpdatePayment(fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
							"status":  -3,
							"comment": "Insufficient trading volume",
						})

						abc.UpdateUserWallet(math.Abs(p.Amount+p.PayFee), uid)
						abc.AddPaymentLog(p.Id, "交易量不足")

						r.Status = 0
						r.Msg = fmt.Sprintf(golbal.Wrong[language][10529], fmt.Sprintf("%.2f", need), fmt.Sprintf("%.2f", available))

						c.JSON(200, r.Response())

						return
					}
				}
			}
		}
	}

	go func() {
		m := make(map[string][]string)
		m["Content-Type"] = []string{"application/json"}
		abc.SendRequest("GET", fmt.Sprintf("%v/auto_withdraw/action?id=%v", conf.WithdrawalAddress, p.Id), strings.NewReader(""), m)
	}()

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func WithdrawalRestrictions2(id, uid int) {
	r := &R{}
	//id := abc.ToInt(c.PostForm("id"))
	//uid := abc.ToInt(c.MustGet("uid"))
	//language := abc.ToString(c.MustGet("language"))
	fmt.Println("WithdrawalRestrictions: id",id," uid=",uid)
	//ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	//if !ok {
	//	//r.Status = 0
	//	//r.Msg = golbal.Wrong[language][10119]
	//	//
	//	//c.JSON(200, r.Response())
	//	return
	//}
	//defer done()

	u := abc.GetUserById(uid)

	if id == 0 {
		//r.Status = 0
		//r.Msg = golbal.Wrong[language][10000]
		//
		//c.JSON(200, r.Response())

		return
	}

	p := abc.GetPaymentOne(fmt.Sprintf("id = %v and type = 'withdraw' and user_id = %v", id, uid))

	if p.Id == 0 {
		//r.Status = 0
		//r.Msg = golbal.Wrong[language][10000]
		//
		//c.JSON(200, r.Response())

		return
	}

	if u.UserType != "sales" && u.SalesType != "admin" {
		payemnt := abc.GetPaymentOne(fmt.Sprintf("type = 'deposit' AND `status` = 1 AND user_id = %v", uid))

		if abc.IsUserWhiteHouse(uid) {
			if payemnt.Id != 0 {
				if p.PayName != "USDT" {
					flag, need, available := abc.DetermineWithdrawalVolume(uid, math.Abs(p.Amount)+math.Abs(p.PayFee), 1)
					fmt.Println("---")
					fmt.Println(need)
					fmt.Println(available)
					fmt.Println("---")
					if !flag {
						abc.UpdatePayment(fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
							"status":  -3,
							"comment": "Insufficient trading volume",
						})

						abc.UpdateUserWallet(math.Abs(p.Amount+p.PayFee), uid)
						abc.AddPaymentLog(p.Id, "交易量不足")
						//r.Status = 0
						//r.Msg = fmt.Sprintf(golbal.Wrong[language][10529], need, available)
						//
						//c.JSON(200, r.Response())
						return
					}
				} else {
					flag, need, available := abc.DetermineWithdrawalVolume(uid, math.Abs(p.Amount)+math.Abs(p.PayFee), 0)
					fmt.Println(fmt.Sprintf("need=%vava=%v", need, available))
					fmt.Println("===flag=", flag)

					if !flag {
						abc.UpdatePayment(fmt.Sprintf("id = %v", p.Id), map[string]interface{}{
							"status":  -3,
							"comment": "Insufficient trading volume",
						})

						abc.UpdateUserWallet(math.Abs(p.Amount+p.PayFee), uid)
						abc.AddPaymentLog(p.Id, fmt.Sprintf("提款失敗，可用交易量不足。提款所需交易%v，可用交易量%v。", fmt.Sprintf("%.2f", need), fmt.Sprintf("%.2f", available)))

						//r.Status = 0
						//r.Msg = fmt.Sprintf(golbal.Wrong[language][10529], fmt.Sprintf("%.2f", need), fmt.Sprintf("%.2f", available))
						//
						//c.JSON(200, r.Response())
						return
					}
				}
			}
		}
	}

	go func() {
		m := make(map[string][]string)
		m["Content-Type"] = []string{"application/json"}
		abc.SendRequest("GET", fmt.Sprintf("%v/auto_withdraw/action?id=%v", conf.WithdrawalAddress, p.Id), strings.NewReader(""), m)
	}()

	r.Status = 1
	r.Msg = ""
	//c.JSON(200, r.Response())
}
