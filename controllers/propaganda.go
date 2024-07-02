package controllers

import (
	"github.com/chenqgp/abc"
	"github.com/gin-gonic/gin"
	"net/http"
)

// PropagandaList 市场资料列表
func PropagandaList(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	t := abc.ToInt(c.PostForm("type"))
	language := abc.ToString(c.MustGet("language"))
	r := R{}
	if !(language == "CN" || language == "TC") {
		language = "EN"
	}
	r.Status, r.Msg, r.Data = abc.PropagandaList(uid, t, language)
	c.JSON(http.StatusOK, r.Response())
}
