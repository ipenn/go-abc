package abc

//
//import (
//	"encoding/json"
//	"net/url"
//	"strings"
//)
//
//type IpRegionRes struct {
//	Code      int    `json:"code"`
//	Ip2region string `json:"ip2region"`
//}
//
//type BackCardRes struct {
//	ErrorCode int    `json:"error_code"`
//	Reason    string `json:"reason"`
//	Result    struct {
//		RespCode     string `json:"respCode"`
//		RespMsg      string `json:"respMsg"`
//		DetailCode   string `json:"detailCode"`
//		BancardInfor struct {
//			BankName string `json:"bankName"`
//			BankId   string `json:"BankId"`
//			Type     string `json:"type"`
//			Cardname string `json:"cardname"`
//			Tel      string `json:"tel"`
//			Icon     string `json:"Icon"`
//		} `json:"bancardInfor"`
//	} `json:"result"`
//}
//
//type IdCardAuthRes struct {
//	Resp struct {
//		Code int    `json:"code"`
//		Desc string `json:"desc"`
//	} `json:"resp"`
//	Data struct {
//		Sex      string `json:"sex"`
//		Address  string `json:"address"`
//		Birthday string `json:"birthday"`
//	} `json:"data"`
//}
//
//type OcrIdCardRes struct {
//	Code   string `json:"code"`
//	Msg    string `json:"msg"`
//	Result struct {
//		Address  string `json:"address"`
//		Birthday string `json:"birthday"`
//		Name     string `json:"name"`
//		Code     string `json:"code"`
//		Sex      string `json:"sex"`
//		Nation   string `json:"nation"`
//	} `json:"result"`
//}
//
////识别身份证图片
//func OcrIdCard(path string) *OcrIdCardRes {
//
//	params := url.Values{}
//	params.Set("image", path)
//	m := make(map[string][]string, 0)
//	m["Authorization"] = []string{"APPCODE " + "11f390a61f0845bd9dd656673b49036c"}
//	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
//	r := DoRequest("POST", "https://ocridcard.market.alicloudapi.com/idCardAuto", strings.NewReader(params.Encode()), m)
//
//	res := &OcrIdCardRes{}
//
//	json.Unmarshal(r, &res)
//
//	return res
//}
//
////识别身份证号码
//func IdCardAuth(name, code string) *IdCardAuthRes {
//	params := url.Values{}
//	params.Set("cardno", code)
//	params.Set("name", name)
//
//	m := make(map[string][]string, 0)
//	m["Authorization"] = []string{"APPCODE " + "11f390a61f0845bd9dd656673b49036c"}
//	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
//	r := DoRequest("GET", "http://idcard.market.alicloudapi.com/lianzhuo/idcard", strings.NewReader(params.Encode()), m)
//	res := &IdCardAuthRes{}
//
//	json.Unmarshal(r, &res)
//	return res
//}
//
////识别银行卡
//func ChineseBankCard(code, name, cardNo string) *BackCardRes {
//	params := url.Values{}
//	params.Set("accountNo", cardNo)
//	params.Set("idCardCode", code)
//	params.Set("name", name)
//
//	m := make(map[string][]string, 0)
//	m["Authorization"] = []string{"APPCODE " + "11f390a61f0845bd9dd656673b49036c"}
//	m["Content-Type"] = []string{"application/x-www-form-urlencoded"}
//
//	r := DoRequest("POST", "https://zball.market.alicloudapi.com/v2/bcheck", strings.NewReader(params.Encode()), m)
//
//	res := &BackCardRes{}
//
//	json.Unmarshal(r, &res)
//	return res
//}
//
////获取ip地址
//func GetIpAddress(ip string) string {
//	paramsMap := make(map[string]string, 0)
//	paramsMap["ip"] = ip
//	s, _ := json.Marshal(&paramsMap)
//
//	m := make(map[string][]string, 0)
//	m["Content-Type"] = []string{"application/json"}
//
//	r := DoRequest("POST", "https://tools.fun/api/v1/tools/ipinfo", strings.NewReader(string(s)), m)
//
//	res := &IpRegionRes{}
//	json.Unmarshal(r, &res)
//
//	return res.Ip2region
//}
