package controllers

import (
	"github.com/chenqgp/abc"
	"github.com/gin-gonic/gin"
)

func GetRebateAll(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	res, _ := abc.SqlOperators(`SELECT rc.group_name, rc.name FROM invite_code ic
									LEFT JOIN rebate_config rc
									ON ic.name = rc.group_name
									WHERE ic.user_id = ? AND rc.is_invite = 1`, uid)

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}
