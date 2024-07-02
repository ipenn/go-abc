package h2pay

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	"github.com/chenqgp/abc/conf"
	chip "github.com/chenqgp/abc/payment/yinlian/yinlian-chip"
	"github.com/chenqgp/abc/third/telegram"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func H2Pay(currency, bank, OrderNo, createTime string, uid int, amount float64, ip string) (int, string, string, float64) {
	rate := GetRate(currency)

	p := abc.GetPaymentConfigOne("Help2Pay")

	var pa []chip.Parameter
	json.Unmarshal([]byte(p.KeySecret), &pa)

	merchant := chip.GetValue("merchant", pa)
	secret := chip.GetValue("secret", pa)
	urlValue := chip.GetValue("url", pa)
	backUrl := chip.GetValue("back_url", pa)
	frontUrl := chip.GetValue("front_url", pa)

	if merchant == "" || secret == "" || urlValue == "" || backUrl == "" {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("支付方式%v订单%v原因%v", "H2Pay", OrderNo, "下单密钥参数错误"))
	}

	t, _ := time.Parse("2006-01-02 15:04:05", createTime)
	t1 := t.Format("20060102150405")
	key := abc.Md5(fmt.Sprintf(`%s%s%s%s%s%s%s%s`, merchant, OrderNo, fmt.Sprintf("%v%d", conf.WebName, uid), fmt.Sprintf("%.2f", amount*rate), currency, t1, secret, ip))

	params := url.Values{}
	params.Set("Merchant", merchant)
	params.Set("Currency", currency)
	params.Set("Customer", fmt.Sprintf("%v%d", conf.WebName, uid))
	params.Set("Reference", OrderNo)
	params.Set("Amount", fmt.Sprintf("%.2f", amount*rate))
	params.Set("Note", "")
	params.Set("DateTime", createTime)
	params.Set("FrontURI", frontUrl)
	params.Set("BackURI", backUrl)
	params.Set("Bank", bank)
	params.Set("Language", "en-us")
	params.Set("ClientIP", ip)
	params.Set("key", key)

	m := make(map[string][]string)
	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	r := abc.DoRequest("POST", urlValue, strings.NewReader(params.Encode()), m)

	s := string(r)
	s = strings.ReplaceAll(s, "<!DOCTYPE html>", "")
	s = strings.ReplaceAll(s, "<html>", "")
	s = strings.ReplaceAll(s, "<head>", "")
	s = strings.ReplaceAll(s, "<title></title>", "")
	s = strings.ReplaceAll(s, "</head>", "")
	s = strings.ReplaceAll(s, "<body>", "")
	s = strings.ReplaceAll(s, "</body>", "")
	s = strings.ReplaceAll(s, "</html>", "")

	return 1, "", s, rate
}

func GetRate(currency string) float64 {
	m := make(map[string][]string)
	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	r := abc.DoRequest("GET", fmt.Sprintf(`https://tw.exchange-rates.org/converter/USD/%s/1`, currency), strings.NewReader(""), m)

	htmlArr := strings.Split(string(r), "1 USD = ")

	if len(htmlArr) > 1 {
		slice := strings.Split(htmlArr[1], fmt.Sprintf(" %s", currency))
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.Trim(slice[0], " "), ",", ""), 64)

		if err != nil {
			log.Println("api GetRate 2", err)
		}

		return v
	}

	return 0
}

func GetBank() map[string]map[string]string {
	mSlice := make(map[string]map[string]string, 0)
	MYR := make(map[string]string, 0)
	MYR["AFF"] = "Affin Bank"
	MYR["ALB"] = "Alliance Bank Malaysia Berhad"
	MYR["AMB"] = "AmBank Group"
	MYR["BIMB"] = "Bank Islam Malaysia Berhad"
	MYR["BSN"] = "Bank Simpanan Nasional"
	MYR["CIMB"] = "CIMB Bank Berhad"
	MYR["HLB"] = "Hong Leong Bank Berhad"
	MYR["HSBC"] = "HSBC Bank (Malaysia) Berhad"
	MYR["MBB"] = "Maybank Berhad"
	MYR["OCBC"] = "OCBC Bank (Malaysia) Berhad"
	MYR["PBB"] = "Public Bank Berhad"
	MYR["RHB"] = "RHB Banking Group"
	MYR["UOB"] = "United Overseas Bank (Malaysia) Bhd"
	THB := make(map[string]string, 0)
	THB["BBL"] = "Bangkok Bank"
	THB["BOA"] = "Bank of Ayudhya (Krungsri)"
	THB["KKR"] = "Karsikorn Bank (K-Bank)"
	THB["KNK"] = "Kiatnakin Bank"
	THB["KTB"] = "Krung Thai Bank"
	THB["SCB"] = "Siam Commercial Bank"
	THB["TMB"] = "TMBThanachart Bank(TTB)"
	THB["PPTP"] = "Promptpay"
	VND := make(map[string]string, 0)
	VND["ACB"] = "Asia Commercial Bank"
	VND["AGB"] = "Agribank"
	VND["BIDV"] = "Bank for Investment and Development of Vietnam"
	VND["DAB"] = "DongA Bank"
	VND["EXIM"] = "Eximbank Vietnam"
	VND["HDB"] = "HD Bank"
	VND["MB"] = "Military Commercial Joint Stock Bank"
	VND["MTMB"] = "Maritime Bank"
	VND["OCB"] = "Orient Commercial Joint Stock Bank"
	VND["SACOM"] = "Sacombank"
	VND["TCB"] = "Techcombank"
	VND["TPB"] = "Tien Phong Bank"
	VND["VCB"] = "Vietcombank"
	VND["VIB"] = "Vietnam International Bank"
	VND["VPB"] = "VP Bank"
	VND["VTB"] = "Vietinbank"
	VND["VIETQR"] = "VietQRpay"
	VND["VIETQRMOMO"] = "VietQRpay MOMO"
	VND["VIETQRZALO"] = "VietQRpay Zalo Pay"
	VND["VIETQRVIETTEL"] = "VietQRpay Viettel Pay"
	IDR := make(map[string]string, 0)
	IDR["BCA"] = "Bank Central Asia"
	IDR["BNI"] = "Bank Negara Indonesia"
	IDR["BRI"] = "Bank Rakyat Indonesia"
	IDR["CIMBN"] = "CIMB Niaga"
	IDR["MDR"] = "Mandiri Bank"
	IDR["PMTB"] = "Permata Bank"
	IDR["PANIN"] = "Panin Bank"
	IDR["QRIS"] = "QRIS"
	IDR["DANAQRIS"] = "DANA QRIS"
	IDR["GOPAYQRIS"] = "GO PAY QRIS"
	IDR["LINKAJAQRIS"] = "LINK AJA QRIS"
	IDR["OVOQRIS"] = "OVO QRIS"
	IDR["SHOPEEQRIS"] = "Shopee Pay QRIS"

	mSlice["MYR"] = MYR
	mSlice["THB"] = THB
	mSlice["VND"] = VND
	mSlice["IDR"] = IDR

	return mSlice
}

//func main() {
//	status, msg, data := H2Pay("MYR", "MBB", fmt.Sprintf("%d_%d", time.Now().Unix(), 16672), time.Now().Format("2006-01-02 15:04:05"), 16672, 500)
//
//	fmt.Println("status===", status)
//	fmt.Println("msg====", msg)
//	fmt.Println("data===", data)
//}
