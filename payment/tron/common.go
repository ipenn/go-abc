package tron

import (
	"bytes"
	"encoding/hex"
	"github.com/chenqgp/abc/conf"
	"strings"

	"github.com/chenqgp/abc"
)

//const (
//	HTTPN = conf.Scheme + "18.139.248.26:8090"
//)

var Header = map[string][]string{
	"accept":       {"application/json"},
	"content-type": {"application/json"},
}

func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func aToHex(address string) string {
	haddress := hex.EncodeToString(Base58Decode([]byte(address)))
	haddress = haddress[2 : len(haddress)-8]
	return haddress
}

type rtnGetContractInfo struct {
	ContractState struct {
		UpdateCycle  int   `json:"update_cycle"`
		EnergyUsage  int64 `json:"energy_usage"`
		EnergyFactor int   `json:"energy_factor"`
	} `json:"contract_state"`
}

func GetContractInfo() rtnGetContractInfo {
	url := conf.HTTPN + `/wallet/getcontractinfo`
	payload := struct {
		Value   string `json:"value"`
		Visible bool   `json:"visible"`
	}{Value: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", Visible: true}
	body := abc.DoRequest("POST", url, bytes.NewReader(abc.ToJSON(payload)), Header)
	var rtn rtnGetContractInfo
	abc.ParseJSON(body, &rtn)
	return rtn
}

type rtnTriggerConstantContract struct {
	Result struct {
		Result bool `json:"result"`
	} `json:"result"`
	EnergyUsed int `json:"energy_used"`
}

func TriggerConstantContract(address string) rtnTriggerConstantContract {
	var rtn rtnTriggerConstantContract
	if isHex(address) || address[0] != 'T' {
		return rtn
	}
	var param string
	prefix := strings.Repeat("0", 64-40)
	param = prefix + conf.USDTContractHex[2:]

	url := conf.HTTPN + `/wallet/triggerconstantcontract`
	payload := struct {
		OwnerAddress     string `json:"owner_address"`
		ContractAddress  string `json:"contract_address"`
		FunctionSelector string `json:"function_selector"`
		Parameter        string `json:"parameter"`
		Visible          bool   `json:"visible"`
	}{
		OwnerAddress:     address,
		ContractAddress:  conf.USDTContract,
		FunctionSelector: "transfer(address,uint256)",
		Parameter:        param,
		Visible:          true,
	}
	body := abc.DoRequest("POST", url, bytes.NewReader(abc.ToJSON(payload)), Header)
	abc.ParseJSON(body, &rtn)
	return rtn
}
