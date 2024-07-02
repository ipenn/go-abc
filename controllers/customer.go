package controllers

import (
	"github.com/chenqgp/abc"
	"github.com/gin-gonic/gin"
)

func RecordUsers(c *gin.Context) {
	r := &R{}
	invitationCode := c.PostForm("invitation_code")
	email := c.PostForm("email")

	i := abc.GetInviteIsExit(invitationCode)

	if i.Id == 0 {
		return
	}

	abc.CreateCustomer(i.UserId, email)

	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}
