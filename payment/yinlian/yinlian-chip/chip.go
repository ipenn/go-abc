package chip

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	"github.com/chenqgp/abc/third/telegram"
	"strings"
)

type ChipResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Link          string
		IntentOrderNo string
	} `json:"data"`
	Success bool `json:"success"`
}

type Req struct {
	CompanyId       string `json:"companyId"`
	CompanyOrderNum string `json:"companyOrderNum"`
	AreaCode        string `json:"areaCode"`
	Phone           string `json:"phone"`
	CoinQuantity    int    `json:"coinQuantity"`
	AsyncUrl        string `json:"asyncUrl"`
	SyncUrl         string `json:"syncUrl"`
	//TotalAmount				int		`json:"totalAmount"` //用户付款的法币总金额(只能传整数)
	Sign                 string `json:"sign"`
	Name                 string `json:"name"`
	IdentityType         int    `json:"identityType"`         //证件类型 1.身份证 2.护照 6.其他
	IdentityPictureFront string `json:"identityPictureFront"` //证件正面照URL
	IdentityPictureBack  string `json:"identityPictureBack"`  //证件反面照URL
	Number               string `json:"number"`               //证件号
	//coinQuantity				float64
}

type Parameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func GetValue(key string, temp []Parameter) string {
	for _, v := range temp {
		if v.Key == key {
			return v.Value
		}
	}

	return ""
}

func ChipPay(OrderNo, Phonectcode, Mobile, TrueName, Identity, idFront, idBack, language string, Amount float64) (int, string, string, float64) {
	p := abc.GetPaymentConfigOne("ChipPay")
	var pa []Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	mid := GetValue("mid", pa)
	secret := GetValue("secret", pa)
	urlValue := GetValue("url", pa)
	callbackUrl := GetValue("callback_url", pa)

	if mid == "" || secret == "" || urlValue == "" || callbackUrl == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Chip", OrderNo, "下单密钥参数错误"))
		return 0, golbal.Wrong[language][10052], "", 0.0
	}

	req := Req{
		CompanyId:       mid,
		CompanyOrderNum: OrderNo,
		AreaCode:        strings.ReplaceAll(Phonectcode, "+", ""),
		Phone:           Mobile,
		AsyncUrl:        callbackUrl,
		SyncUrl:         callbackUrl,
		//TotalAmount:int(param.Amount*u.Rate),
		CoinQuantity:         int(Amount),
		Name:                 TrueName,
		IdentityType:         1,
		Number:               Identity,
		IdentityPictureFront: idFront,
		IdentityPictureBack:  idBack,
	}

	signStr := fmt.Sprintf(`areaCode=%s&asyncUrl=%s&coinQuantity=%d&companyId=%s&companyOrderNum=%s&identityPictureBack=%s&identityPictureFront=%s&identityType=1&name=%s&number=%s&phone=%s&syncUrl=%s`,
		req.AreaCode, req.AsyncUrl, req.CoinQuantity, req.CompanyId, req.CompanyOrderNum, req.IdentityPictureBack, req.IdentityPictureFront, req.Name, req.Number, req.Phone, req.SyncUrl)

	req.Sign = HmacSha256RSA(signStr, secret)

	m := make(map[string]interface{})
	m["companyId"] = req.CompanyId
	m["companyOrderNum"] = req.CompanyOrderNum
	m["areaCode"] = req.AreaCode
	m["phone"] = req.Phone
	m["coinQuantity"] = req.CoinQuantity
	m["asyncUrl"] = req.AsyncUrl
	m["syncUrl"] = req.SyncUrl
	m["sign"] = req.Sign
	m["name"] = req.Name
	m["identityType"] = req.IdentityType
	m["identityPictureFront"] = req.IdentityPictureFront
	m["identityPictureBack"] = req.IdentityPictureBack
	m["number"] = req.Number

	s, _ := json.Marshal(&m)

	m1 := make(map[string][]string)
	m1["Content-Type"] = []string{"application/json"}

	r := abc.DoRequest("POST", urlValue, strings.NewReader(string(s)), m1)

	res := &ChipResult{}

	json.Unmarshal(r, &res)

	if !res.Success {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "Chip", OrderNo, string(r)))
		return 0, golbal.Wrong[language][10052], "", 0.0
	}

	return 1, "", res.Data.Link, 1.0
}

func HmacSha256RSA(message string, secret string) string {
	decodeString, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		fmt.Println("DecodeString err", err)
	}
	private, err := x509.ParsePKCS8PrivateKey(decodeString)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
	}

	h := crypto.Hash.New(crypto.SHA256)
	h.Write([]byte(message))
	hashed := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, private.(*rsa.PrivateKey),
		crypto.SHA256, hashed)
	if err != nil {
		fmt.Println("Error from signing:", err)
		return ""
	}

	signedString := base64.StdEncoding.EncodeToString(signature)
	//fmt.Printf("Encoded: %v\n", signedString)
	return signedString
}

//func main() {
//	status, msg, data := ChipPay(fmt.Sprintf("%d_%d", time.Now().Unix(), 16672), "+86", "18782660531", "刘峰", "16672", "", "", 500.0)
//
//	fmt.Println("status===", status)
//	fmt.Println("msg====", msg)
//	fmt.Println("data===", data)
//}
