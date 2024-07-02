package teleport

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	"github.com/chenqgp/abc/third/telegram"
	"net/url"
	"sort"
	"strings"
)

type Result struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		OrderNo string `json:"order_no"`
		SendURL string `json:"send_url"`
		UserID  string `json:"user_id"`
	} `json:"data"`
	Code int `json:"code"`
}

func TeleportPay(orderNo, createTime, username, language string, amount float64, uid int, rate float64, ip string) (int, string, string) {

	if rate <= 6 && rate != 1 {
		return 0, "汇率异常", ""
	}

	p := abc.GetPaymentConfigOne("Teleport")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	sys_no := chip.GetValue("sys_no", pa)
	sign_key := chip.GetValue("sign_key", pa)
	urlValue := chip.GetValue("url", pa)

	if sys_no == "" || sign_key == "" || urlValue == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Teleport", orderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], ""
	}

	m := make(map[string]string, 0)
	m["order_id"] = orderNo
	m["order_amount"] = abc.ToString(int(amount * rate))
	m["sys_no"] = sys_no
	m["amount_type"] = ""
	m["pay_user_name"] = username
	m["user_id"] = abc.ToString(uid)
	m["order_ip"] = ip
	m["order_time"] = createTime
	m["rate"] = fmt.Sprintf("%.2f", rate)

	urlAddress := UrlBuild(m)

	sign := abc.Md5(urlAddress + sign_key)

	params := url.Values{}
	params.Set("order_id", m["order_id"])
	params.Set("order_amount", m["order_amount"])
	params.Set("amount_type", "")
	params.Set("sys_no", sys_no)
	params.Set("user_id", m["user_id"])
	params.Set("pay_user_name", m["pay_user_name"])
	params.Set("order_ip", m["order_ip"])
	params.Set("order_time", m["order_time"])
	params.Set("rate", m["rate"])
	params.Set("sign", sign)

	m1 := make(map[string][]string)
	m1["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	r := abc.DoRequest("POST", urlValue, strings.NewReader(params.Encode()), m1)

	res := Result{}
	json.Unmarshal(r, &res)

	if res.Status != "success" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Teleport", orderNo, string(r)))
		return 0, golbal.Wrong[language][10052], ""
	}

	return 1, "", res.Data.SendURL + "?in_order_id=" + res.Data.OrderNo + "&user_id=" + res.Data.UserID
}

func UrlBuild(m map[string]string) string {
	var keys []string
	var values []string
	for k, _ := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, v := range keys {
		values = append(values, v+"="+url.QueryEscape(m[v]))
	}

	return strings.Join(values, "&")
}

type otcRes struct {
	Code    int        `json:"code"`
	Msg     string     `json:"msg"`
	Data    otcResData `json:"data"`
	Success bool       `json:"success"`
}

type otcResData struct {
	Id        int     `json:"id"`
	CoinType  string  `json:"coinType"`
	BuyPrice  float64 `json:"buyPrice"`
	SellPrice float64 `json:"sellPrice"`
	Link      string  `json:"link"`
}

type ChippayRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Id                int64   `json:"id"`
		CoinType          string  `json:"coinType"`
		Status            int     `json:"status"`
		SellStatus        int     `json:"sellStatus"`
		BuyStatus         int     `json:"buyStatus"`
		Platform          string  `json:"platform"`
		RealBuyPrice      float64 `json:"realBuyPrice"`
		RealSellPrice     float64 `json:"realSellPrice"`
		OffsetMiddlePrice float64 `json:"offsetMiddlePrice"`
	} `json:"data"`
	Success bool `json:"success"`
}

func GetBaseExchangeRate() float64 {
	m := make(map[string][]string)
	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	res := abc.DoRequest("GET", "https://c2c.chippay.com/api/cola/quotePriceBusiness/priceConfig/getOpenPriceConfig?coinType=cnyusdt&platform=otc", strings.NewReader(""), m)
	t := ChippayRes{}
	json.Unmarshal(res, &t)
	return t.Data.RealBuyPrice
}
