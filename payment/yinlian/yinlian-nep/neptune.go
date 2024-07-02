package nep

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	teleport "github.com/chenqgp/abc/payment/yinlian/yinlian-teleport"
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

func NeptunePay(orderNo, username, ip, createTime, language string, amount, rate float64, uid int) (int, string, string) {
	if rate <= 6 && rate != 1 {
		return 0, "汇率异常", ""
	}

	p := abc.GetPaymentConfigOne("Neptune")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	sys_no := chip.GetValue("sys_no", pa)
	secret := chip.GetValue("secret", pa)
	urlValue := chip.GetValue("url", pa)

	if sys_no == "" || secret == "" || urlValue == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Nep", orderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], ""
	}

	m := make(map[string]string, 0)
	m["order_id"] = orderNo
	m["order_amount"] = abc.ToString(int(amount * rate))
	m["sys_no"] = sys_no
	m["user_id"] = abc.ToString(uid)
	m["order_ip"] = ip
	m["order_time"] = createTime
	m["pay_user_name"] = username

	sign := abc.Md5(teleport.UrlBuild(m) + secret)

	params := url.Values{}
	params.Set("order_id", m["order_id"])
	params.Set("order_amount", m["order_amount"])
	params.Set("sys_no", m["sys_no"])
	params.Set("user_id", m["user_id"])
	params.Set("order_ip", m["order_ip"])
	params.Set("order_time", m["order_time"])
	params.Set("pay_user_name", m["pay_user_name"])
	params.Set("sign", sign)

	m1 := make(map[string][]string)
	m1["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	r := abc.DoRequest("POST", urlValue, strings.NewReader(params.Encode()), m1)

	res := Result{}
	json.Unmarshal(r, &res)

	if res.Status != "success" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Nep", orderNo, string(r)))
		return 0, golbal.Wrong[language][10052], ""
	}

	return 1, "", res.Data.SendURL + "?in_order_id=" + res.Data.OrderNo + "&user_id=" + res.Data.UserID
}

func buildUrl(data map[string]interface{}) string {
	var (
		keys  []string
		query []string
	)
	for index, _ := range data {
		keys = append(keys, index)
	}
	sort.Strings(keys)
	for _, k := range keys {
		query = append(query, fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprintf("%v", data[k]))))
	}
	return strings.Join(query, `&`)
}

func Sign(data map[string]interface{}, key string) string {
	signTemp := buildUrl(data) + key
	fmt.Println(signTemp)
	signData := abc.Md5(signTemp)
	return strings.ToLower(signData)
}
