package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	bank2 "github.com/chenqgp/abc/third/bank"
	"github.com/gin-gonic/gin"
	"net/http"
)

func BankCardList(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}
	BankCardListReturn := struct {
		Bank   []abc.Bank   `json:"bank"`
		Wallet []abc.Wallet `json:"wallet"`
	}{
		abc.GetBankCardList(uid),
		abc.GetWalletList(uid),
	}
	r.Status, r.Data = 1, BankCardListReturn
	c.JSON(http.StatusOK, r.Response())
}

func CreateBankCare(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	bankName := c.PostForm("bank_name")
	bankNo := c.PostForm("bank_no")
	identity := c.PostForm("identity")
	name := c.PostForm("name")
	bankAddress := c.PostForm("bank_address")
	swift := c.PostForm("swift")
	iban := c.PostForm("iban")
	area := abc.ToInt(c.PostForm("area"))
	bankCardType := abc.ToInt(c.PostForm("bank_card_type"))
	bankType := abc.ToInt(c.PostForm("bank_type"))
	t := abc.ToInt(c.PostForm("type"))
	bankCode := c.PostForm("bank_code")

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if user.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if t != 5 {
		if user.AuthStatus != 1 {
			r.Status, r.Msg = 0, golbal.Wrong[language][10027]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	}

	files := HandleFilesAllFiles(c, uid, "upload", "files")
	f := string(abc.ToJSON(files))

	status := 0
	var b abc.Bank
	r.Status, r.Msg, b = abc.SaveBankInformation(bankType, user, status, name, bankName, bankNo, bankAddress,
		f, swift, iban, language, area, bankCardType, identity, bankCode)
	if r.Status == 0 {
		c.JSON(http.StatusOK, r.Response())
		return
	}
	// && (user.UserType != "sales" && user.SalesType != "admin")
	if bankType == 1 {
		ui := abc.GetUserInfoById(uid)
		if identity == "" && ui.IdentityType == "Identity card" && ui.Identity != "" {
			identity = ui.Identity
		} else if ui.IdentityType != "Identity card" && ui.ChineseIdentity != "" {
			identity = ui.ChineseIdentity
		} else if identity == "" {
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(http.StatusOK, r.Response())
			return
		}

		res := bank2.ChineseBankCard(identity, name, bankNo)
		//fmt.Println(fmt.Sprintf("%+v", res))
		status = -1
		if res.Result.RespCode == "T" {
			status = 1
			user.CoinStatus = 1
			abc.SaveUser(user)
		}
		abc.UpdateSql(b.TableName(), fmt.Sprintf("id=%d", b.Id), map[string]interface{}{
			"status": status,
		})
	}

	c.JSON(http.StatusOK, r.Response())
}

func UpdateBankCare(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	language := abc.ToString(c.MustGet("language"))
	bankName := c.PostForm("bank_name")
	bankNo := c.PostForm("bank_no")
	name := c.PostForm("name")
	bankAddress := c.PostForm("bank_address")
	identity := c.PostForm("identity")
	swift := c.PostForm("swift")
	iban := c.PostForm("iban")
	area := abc.ToInt(c.PostForm("area"))
	bankCardType := abc.ToInt(c.PostForm("bank_card_type"))
	bankType := abc.ToInt(c.PostForm("bank_type"))
	bankCode := c.PostForm("bank_code")

	old_files := c.PostForm("old_files")

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if user.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	bank := abc.GetBankCardById(id)

	if bank.Id == 0 || bank.UserId != uid {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if bank.Status == 1 {
		r.Msg = golbal.Wrong[language][10015]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	files := HandleFilesAllFiles(c, uid, "upload", "files")
	var oldfile []string
	json.Unmarshal([]byte(bank.Files), &oldfile)
	var keep []string
	json.Unmarshal([]byte(old_files), &keep)
	var del []string

	for _, s := range oldfile {
		flag := true
		for _, k := range keep {
			if s == k {
				flag = false
				files = append(files, k)
				break
			}
		}
		if flag {
			del = append(del, s)
		}
	}
	abc.DelFile(del)
	f := ""
	if len(files) != 0 {
		f = string(abc.ToJSON(files))
	}

	status := 0
	// && (user.UserType != "sales" && user.SalesType != "admin")
	if bankType == 1 {
		ui := abc.GetUserInfoById(uid)
		if identity == "" && ui.IdentityType == "Identity card" && ui.Identity != "" {
			identity = ui.Identity
		} else if ui.IdentityType != "Identity card" && ui.ChineseIdentity != "" {
			identity = ui.ChineseIdentity
		} else if identity == "" {
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(http.StatusOK, r.Response())
			return
		} else if identity != "" {
			abc.UpdateSql("user_info", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"chinese_identity": identity,
			})
		}

		res := bank2.ChineseBankCard(identity, name, bankNo)
		//fmt.Println(fmt.Sprintf("%+v", res))
		status = -1
		if res.Result.RespCode == "T" {
			status = 1
		}
	}

	bank.Status = status
	bank.TrueName = name
	bank.BankName = bankName
	bank.BankNo = bankNo
	bank.BankAddress = bankAddress
	bank.Files = f
	bank.Swift = swift
	bank.Iban = iban
	bank.Area = area
	bank.BankCardType = bankCardType
	bank.BankCode = bankCode
	if r.Status = bank.SaveBank(fmt.Sprintf("id=%d", bank.Id)); r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
	}
	c.JSON(http.StatusOK, r.Response())
}

func DeleteBankCare(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if user.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	bank := abc.GetBankCardById(id)
	if bank.UserId != user.Id {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r.Status = abc.DeleteBank(fmt.Sprintf("id=%d and user_id=%d", id, uid))
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
	}
	var remove []string
	json.Unmarshal([]byte(bank.Files), &remove)
	abc.DelFile(remove)
	c.JSON(http.StatusOK, r.Response())
}

func GetChineseIdentityStatus(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	info := abc.GetUserInfoById(uid)
	r := R{}
	r.Status = 1
	r.Data = false
	if info.IdentityType != "Identity card" && info.ChineseIdentity == "" {
		r.Data = true
	}
	c.JSON(http.StatusOK, r.Response())
}
