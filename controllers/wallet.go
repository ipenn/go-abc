package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	address2 "github.com/chenqgp/abc/third/address"
	validator "github.com/chenqgp/abc/third/google"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CreateWallet(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	address := c.PostForm("address")
	addressType := c.PostForm("address_type")
	name := c.PostForm("name")
	tag := c.PostForm("tag")
	smsCode := c.PostForm("sms_code")
	gCode := c.PostForm("google_code")
	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	if strings.HasPrefix(address, "T") {
		if len(address) != 34 {
			r.Msg = golbal.Wrong[language][10123]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else if strings.HasPrefix(address, "41") {
		if len(address) != 42 {
			r.Msg = golbal.Wrong[language][10123]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else {
		r.Msg = golbal.Wrong[language][10123]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if !address2.VerifyAddress(address) {
		r.Msg = golbal.Wrong[language][10124]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	g := abc.GetUserG2(uid)

	if g.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10097]

		c.JSON(200, r.Response())

		return
	}

	myCode, _ := validator.GetCode(g.Secret)

	if myCode != gCode {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10098]

		c.JSON(200, r.Response())

		return
	}

	user := abc.GetUserById(uid)
	sms := abc.VerifySmsCode(user.Mobile, smsCode)
	if sms.Id == 0 {
		r.Msg = golbal.Wrong[language][10033]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if address == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

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

	res,_:=abc.SqlOperator(`select id from wallet where address=?`,address)
	if res!=nil{
		r.Msg=golbal.Wrong[language][10521]
		c.JSON(http.StatusOK,r.Response())
		return
	}
	abc.DeleteSmsOrMail(sms.Id)
	wallet := abc.Wallet{
		Address:     address,
		AddressType: addressType,
		Name:        name,
		Tag:         tag,
		CreateTime:  abc.FormatNow(),
		UserId:      uid,
		IsDel:       0,
		Status:      1,
	}

	if r.Status = wallet.CreateWallet(); r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
	}

	abc.AddUserLog(uid, "CreateUsdtWallet", user.Email, abc.FormatNow(), c.ClientIP(), tag)
	c.JSON(http.StatusOK, r.Response())
}

func UpdateWallet(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))
	address := c.PostForm("address")
	addressType := c.PostForm("address_type")
	name := c.PostForm("name")
	tag := c.PostForm("tag")
	smsCode := c.PostForm("sms_code")
	gCode := c.PostForm("google_code")
	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()
	
	if strings.HasPrefix(address, "T") {
		if len(address) != 34 {
			r.Msg = golbal.Wrong[language][10123]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else if strings.HasPrefix(address, "41") {
		if len(address) != 42 {
			r.Msg = golbal.Wrong[language][10123]
			c.JSON(http.StatusOK, r.Response())
			return
		}
	} else {
		r.Msg = golbal.Wrong[language][10123]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	if !address2.VerifyAddress(address) {
		r.Msg = golbal.Wrong[language][10124]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	g := abc.GetUserG2(uid)

	if g.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10097]

		c.JSON(200, r.Response())

		return
	}

	myCode, _ := validator.GetCode(g.Secret)

	if myCode != gCode {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10098]

		c.JSON(200, r.Response())

		return
	}
	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	sms := abc.VerifySmsCode(user.Mobile, smsCode)
	if sms.Id == 0 {
		r.Msg = golbal.Wrong[language][10033]
		c.JSON(http.StatusOK, r.Response())
		return
	}

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
	if address == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	wallet := abc.GetWalletById(id)
	if wallet.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	abc.DeleteSmsOrMail(sms.Id)
	wallet.Address = address
	wallet.AddressType = addressType
	wallet.Name = name
	wallet.Tag = tag
	wallet.CreateTime = abc.FormatNow()

	if r.Status = wallet.SaveWallet(fmt.Sprintf("id=%d", wallet.Id)); r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
	}
	c.JSON(http.StatusOK, r.Response())
}

func DeleteWallet(c *gin.Context) {
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
	r.Status = abc.DeleteWallet(fmt.Sprintf("id=%d and user_id=%d", id, uid))
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
	}

	wallet := abc.GetWalletById(id)
	abc.AddUserLog(uid, "DeleteUsdtWallet", user.Email, abc.FormatNow(), c.ClientIP(), wallet.Tag)
	c.JSON(http.StatusOK, r.Response())
}
