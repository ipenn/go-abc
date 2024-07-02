package alpapay

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	"github.com/chenqgp/abc/third/telegram"
	"net/url"
	"strings"
)

func AlpapayPay(orderNo, language string, amount, rate float64) (int, string, string) {
	p := abc.GetPaymentConfigOne("7Star")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	appId := chip.GetValue("app_id", pa)
	key := chip.GetValue("key", pa)
	returnUrl := chip.GetValue("return_url", pa)
	notifyUrl := chip.GetValue("notify_url", pa)
	urlValue := chip.GetValue("url", pa)

	if appId == "" || key == "" || returnUrl == "" || notifyUrl == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "BFT", orderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], ""
	}

	if rate <= 6 && rate != 1 {
		return 0, "汇率异常", ""
	}

	params := url.Values{}
	params.Set("appId", appId)
	params.Set("orderNo", orderNo)
	params.Set("txnAmt", fmt.Sprintf("%.2f", amount*rate))
	params.Set("returnUrl", returnUrl)
	params.Set("notifyUrl", notifyUrl)
	params.Set("sign", abc.Md5(fmt.Sprintf("appId=%s&notifyUrl=%v&orderNo=%v&returnUrl=%v&txnAmt=%v&key=%v", appId, notifyUrl, orderNo, returnUrl, fmt.Sprintf("%.2f", amount*rate), key)))

	m := make(map[string][]string)
	m["Content-Type"] = []string{"application/x-www-form-urlencoded;charset=utf-8"}

	r := abc.DoRequest("POST", urlValue, strings.NewReader(params.Encode()), m)

	return 1, "", string(r)
}
