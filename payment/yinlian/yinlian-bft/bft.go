package bft

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	"github.com/chenqgp/abc/third/telegram"
	"strings"
)

type BftResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
	Success bool   `json:"success"`
}

func BftPay(orderNo, username, language string, amount, rate float64, uid int) (int, string, string) {

	p := abc.GetPaymentConfigOne("BFT")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	Mid := chip.GetValue("mid", pa)
	Secret := chip.GetValue("secret", pa)
	UrlValue := chip.GetValue("url", pa)

	if Mid == "" || Secret == "" || UrlValue == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "BFT", orderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], ""
	}
	if rate <= 6 && rate != 1 {
		return 0, "汇率异常", ""
	}

	mParams := make(map[string]string, 0)
	mParams["uid"] = Mid
	mParams["uniqueCode"] = abc.ToString(uid)
	mParams["money"] = abc.ToString(int(amount * rate))
	mParams["payType"] = "1"
	mParams["orderId"] = orderNo
	mParams["payerName"] = username
	mParams["signature"] = abc.Md5(fmt.Sprintf(`money=%v&orderId=%v&payType=1&payerName=%v&uid=%v&uniqueCode=%v&key=%v`, abc.ToString(int(amount*rate)), orderNo, username, Mid, uid, Secret))

	s, _ := json.Marshal(&mParams)

	m := make(map[string][]string)
	m["Content-Type"] = []string{"application/json"}

	r := abc.DoRequest("POST", UrlValue, strings.NewReader(string(s)), m)

	res := BftResult{}
	json.Unmarshal(r, &res)

	if !res.Success {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "BFT", orderNo, string(r)))
		return 0, golbal.Wrong[language][10052], ""
	}

	return 1, "", res.Data
}

//func main() {
//	status, msg, data := BftPay(fmt.Sprintf("%d_%d", time.Now().Unix(), 16672), "张三", 500.0, 16672)
//
//	fmt.Println("status===", status)
//	fmt.Println("msg====", msg)
//	fmt.Println("data===", data)
//	//t := time.Now().Format("2006-01-02 15:04:05")
//	//t1, _ := time.Parse("2006-01-02 15:04:05", t)
//	//t2 := t1.Format("20060102150405")
//	//fmt.Println(t)
//	//fmt.Println(t2)
//}
