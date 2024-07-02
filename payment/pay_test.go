package payment

import (
	"fmt"
	"github.com/chenqgp/abc/payment/tron"
	alpapay "github.com/chenqgp/abc/payment/yinlian/yinlian-alpapay"
	"testing"
	"time"
)

func TestAlpapay(T *testing.T) {
	status, msg, data := alpapay.AlpapayPay(fmt.Sprintf("%d_%d", time.Now().UnixMicro(), 16672), "CN", 200, 7.65)

	fmt.Println("==status====", status)
	fmt.Println("====msg=====", msg)
	fmt.Println("====data===", data)
}

func TestGetTokenInfo(T *testing.T) {
	t1, err := tron.GetTokenInfo("f2cb0566cbb271a32086e9e3d39c8f56cd98ebd297afc8a8149dcc226c98ed8c")
	if t1.Receipt.Result == "SUCCESS" {
		fmt.Println(t1)
	}
	fmt.Println(err)
}
