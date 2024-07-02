package controllers

import (
	"fmt"
	"github.com/chenqgp/abc/third/mt4"
	"testing"
)

func TestCrmPostPosition(t *testing.T) {
	//5010297
	m := map[string]interface{}{
		"login": "5010293",
	}
	res := mt4.CrmPost("api/position", m)
	t.Log(fmt.Sprintf("%+v", res))
	//data := make([]map[string]any, 0)
	for _, a := range res["list"].([]interface{}) {
		t.Log(fmt.Sprintf("%+v", a))
	}
}

func TestCrmPostAccountInfo(t *testing.T) {
	//5010297
	m := map[string]interface{}{
		"login": "8",
	}
	res := mt4.CrmPost("api/account_info", m)
	t.Log(fmt.Sprintf("%+v", res))
}

func TestCrmPostOpen(t *testing.T) {
	m := map[string]interface{}{
		"name":            "SHE GUORUI",
		"email":           "",
		"group":           "CFH-STD-0",
		"country":         "",
		"city":            "",
		"address":         "",
		"phone":           "",
		"zipcode":         "99",
		"read_only":       "0",
		"id":              "16659",
		"noswap":          "0",
		"experience":      "0",
		"send_reports":    "1",
		"leverage":        "200",
		"leverage_status": "1",
	}
	res := mt4.CrmPost("api/openaccount", m)
	t.Log(fmt.Sprintf("%+v", res))
}

// 自动任务
//
//	func TestVipBirth(t *testing.T) {
//		VipBirth()
//	}
//func TestInterests(t *testing.T) {
//	Interest()
//}

//func TestPaymentExpired(t *testing.T) {
//	PaymentExpired()
//}
//func TestUserDisable(t *testing.T) {
//	UserDisable()
//}
//func TestSalesDisable(t *testing.T) {
//	SalesDisable()
//}
//
//	func TestScoreExpired(t *testing.T) {
//		ScoreExpired()
//	}
//func TestAccountLeverage(t *testing.T) {
//	AccountLeverage()
//}
//
//func TestMargin(t *testing.T) {
//	Margin()


