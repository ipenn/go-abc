package m

import (
	"net/http"
	"strings"
	"time"
	golbal "github.com/chenqgp/abc/global"

	"github.com/chenqgp/abc"
	"github.com/chenqgp/abc/controllers"
	"github.com/gin-gonic/gin"
)

// 权限中间件
// -------------------------------------管理员-------------------------------------
func UA() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Admin], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------二级业务员+三级业务员-----------------------------------------------------
func SalesTwoOrThree() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Sale2Sale3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------销售+业务员-----------------------------------------------------
func SalesOrAdminOrIb() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[AdminSale1Sale2Sale3Ib1Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------所有业务员-----------------------------------------------------
func SalesAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Sale1Sale2Sale3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------所有伙伴-----------------------------------------------------
func IbAll() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Ib1Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------用户+所有伙伴-----------------------------------------------------
func IbAllOrUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[UserIb1Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------销售+所有伙伴-----------------------------------------------------
func AdminOrIb() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))

		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[AdminIb1Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------伙伴二三级-----------------------------------------------------
func IbTwo() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------所有伙伴+所有业务员-----------------------------------------------------
func IBAndSales() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Sale1Sale2Sale3Ib1Ib2Ib3], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------伙伴+admin+user-----------------------------------------------------
func AdminOrUserOrIb() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))

		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Ib1Ib2Ib3AdminUser], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// -----------------------------------2,3伙伴+2,3级业务员-----------------------------------------------------
func AdminIbAndSalesTwoOrThree() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))

		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		if !filter.verify(Keys[Ib2Ib3Sale2Sale3Admin], token.Role) {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 402
			r.Msg = golbal.Wrong[language][402]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// 公共中间件
func PublicRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		r := &controllers.R{}
		preHeaderHandle(c)
		token := abc.FetchFromToken(req(c, "token"))
		if token.Expire < time.Now().Unix() {
			language := abc.ToString(c.MustGet("language"))
			r.Status = 401
			r.Msg = golbal.Wrong[language][401]
			c.Abort()
			c.JSON(http.StatusNonAuthoritativeInfo, r.Response())
			return
		}

		c.Set("uid", token.Uid)
	}
}

// 访客中间件
func Guest() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()
		preHeaderHandle(c)
		c.Set("uid", 0)
	}
}

// -----------------------------请求头预处理-----------------------------------
func preRequest(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Abort()
		c.JSON(http.StatusNonAuthoritativeInfo, map[string]interface{}{"status": "ok"})
		return
	}
}

func nonSimpleRequest(c *gin.Context) {
	xxx := c.Request.Header.Get("token")
	c.Set("token", xxx)
	yyy := c.Request.Header.Get("language")

	if yyy == "" {
		yyy = "EN"
	} else {
		if !strings.Contains("EN,CN,TC,VIE,THA,IDN,NYS", yyy) {
			yyy = "EN"
		}
	}

	c.Set("language", yyy)
}

func header(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "token,language")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
}

func json(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
}

func req(c *gin.Context, key string) string {
	t := c.PostForm(key)
	if c.Request.Method == "GET" {
		t = c.Param(key)
	}
	if t == "" {
		t = c.MustGet(key).(string)
	}
	return t
}

//-----------------------------请求头预处理结束--------------------------------

func preHeaderHandle(c *gin.Context) {
	header(c)
	preRequest(c)
	nonSimpleRequest(c)
}
