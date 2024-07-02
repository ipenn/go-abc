package controllers

import (
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	validator "github.com/chenqgp/abc/third/google"
	"github.com/gin-gonic/gin"
)

func GetSecret(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))

	g := abc.BindOrNot(uid)
	if g.Id > 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10096]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	//删除之前的
	abc.DelGoogleSecret(uid)

	secret := validator.GetSecret()

	abc.CreateSecret(secret, uid)

	m := make(map[string]interface{}, 0)
	m["secret"] = secret
	m["otp"] = fmt.Sprintf("otpauth://totp/ACCOUNT?secret=%s&issuer=IEXS(%s)", secret, u.Username)

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func UserIsBound(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	g := abc.BindOrNot(uid)

	flag := true

	if g.Id == 0 {
		flag = false
	}

	r.Status = 1
	r.Msg = ""
	r.Data = flag

	c.JSON(200, r.Response())
}

func VerifyG2Code(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	code := c.PostForm("code")

	g := abc.GetUserG2(uid)

	if g.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10097]

		c.JSON(200, r.Response())

		return
	}

	myCode, _ := validator.GetCode(g.Secret)

	if myCode != code {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10098]

		c.JSON(200, r.Response())

		return
	}

	if g.Id > 0 && g.Status == 0 {
		abc.UpdateSql("user_g2code", fmt.Sprintf("id = %v", g.Id), map[string]interface{}{
			"status": 1,
		})
	}

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}
