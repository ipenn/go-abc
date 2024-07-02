package yinlian_tops

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	"github.com/chenqgp/abc/third/telegram"
	"strings"
	"time"
)

type TopsResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Link string
	} `json:"data"`
}

type TopsReq struct {
	MerchantOrderNo string  `json:"merchantOrderNo"`
	Status          int     `json:"status"`
	OrderAmount     float32 `json:"orderAmount"` //提交订单时候人民币的金额
	PayAmount       float32 `json:"payAmount"` //实际支付人民币的金额
	Timestamp       int64   `json:"timestamp"`
}

func TopsPay(orderNo, trueName, phone, idNo, language string, amount, rate float64) (int, string, string) {
	if rate <= 6 && rate != 1 {
		return 0, "汇率异常", ""
	}

	p := abc.GetPaymentConfigOne("TOPS")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	access := chip.GetValue("access", pa)
	access_secret := chip.GetValue("access_secret", pa)
	callback_url := chip.GetValue("callback_url", pa)
	sync_url := chip.GetValue("sync_url", pa)

	if access == "" || access_secret == "" || callback_url == "" || sync_url == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "TOPS", orderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], ""
	}

	m := make(map[string]interface{}, 0)
	m["MerchantOrderNo"] = orderNo
	m["AsyncUrl"] = callback_url
	m["SyncUrl"] = sync_url
	m["TrueName"] = trueName
	m["Phone"] = phone
	m["IdNo"] = idNo
	m["OrderAmount"] = amount
	m["PayAmount"] = int(amount * rate)
	m["UnitPrice"] = rate

	timestamp := abc.ToString(time.Now().Unix())
	m1 := make(map[string][]string, 0)
	m1["Content-Type"] = []string{"application/json"}
	m1["access"] = []string{access}
	m1["timestamp"] = []string{timestamp}

	s, _ := json.Marshal(&m)
	message := fmt.Sprintf("%s%s%s", "/v5/place_order", timestamp, string(s))
	mac := hmac.New(sha256.New, []byte(access_secret))
	mac.Write([]byte(message))

	m1["sign"] = []string{hex.EncodeToString(mac.Sum(nil))}

	r := abc.DoRequest("POST", "https://www.topsotc.com/v5/place_order", strings.NewReader(string(s)), m1)

	var topsResult TopsResp
	json.Unmarshal(r, &topsResult)

	if topsResult.Code != 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "TOPS", orderNo, string(r)))
		return 0, golbal.Wrong[language][10052], ""
	}

	return 1, "", topsResult.Data.Link
}

func TopsBackSign(data []byte) string {
	p := abc.GetPaymentConfigOne("TOPS")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	access_secret := chip.GetValue("access_secret", pa)

	reqData := TopsReq{}

	json.Unmarshal(data, &reqData)
	//加密
	message := fmt.Sprintf("%s%s", fmt.Sprintf("%d", reqData.Timestamp), string(data))
	mac := hmac.New(sha256.New, []byte(access_secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
