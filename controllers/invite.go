package controllers

import (
	"database/sql"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/pdf"
	file "github.com/chenqgp/abc/third/uFile"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetInviteInfo(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}

	inviteInfoReturn := struct {
		InviteCode  []abc.InviteCode `json:"invite_code"`
		Info        abc.InviteInfo   `json:"info"`
		InviteCount abc.InviteCount  `json:"invite_count"`
	}{
		InviteCode:  abc.GetInviteCode(uid),
		Info:        abc.GetInviteInfo(uid),
		InviteCount: abc.GetInviteCount(uid),
	}
	r.Status, r.Msg, r.Data = 1, "", inviteInfoReturn
	c.JSON(http.StatusOK, r.Response())
}

func AgreeAgreement2(c *gin.Context) {
	uid := abc.ToInt(c.Query("uid"))
	for i := 0; i < 300; i++ {
		sql := fmt.Sprintf("select id, ib_no, username,true_name,mobile,rebate_cate,path from user left join user_info on user.id=user_info.user_id where left(user_type,1) = 'L' and ib_no != '' and agreement = '' and agreement_fee = ''  and id > %d limit 1", uid)
		res, err := abc.SqlOperator(sql)
		if err != nil {
			c.String(200, fmt.Sprintf("no data:%d", uid))
			return
		}
		if res != nil {
			uid = abc.ToInt(abc.PtoString(res, "id"))
			if uid == 0{
				c.String(200, "id = 0")
				return
			}
			code := abc.PtoString(res, "ib_no")
			trueName := abc.PtoString(res, "true_name")
			mobile := abc.PtoString(res, "mobile")
			username := abc.PtoString(res, "username")
			path := abc.PtoString(res, "path")
			rebateCate := abc.ToInt(abc.PtoString(res, "rebate_cate"))

			userinfo := abc.GetUserInfoById(uid)
			superior := abc.GetUserIdAndUserTypeByPath(path)
			var csc []abc.CommissionSetCustom
			for _, i := range superior {
				if i.UserType[:1] == "L" && uid != i.Id {
					//Whether a commission is set
					csc = abc.GetCommissionSetCustomByUid(uid, "status=1")
					if len(csc) == 0 {
						c.String(200, fmt.Sprintf("csc:%d", uid))
						return
					}
				}
			}

			path1, path2 := pdf.RpPDF(code, trueName, userinfo.Identity, mobile, username, rebateCate, csc)
			p1 := file.UploadFile(path1, uid)
			p2 := file.UploadFile(path2, uid)
			userinfo.Agreement = p1
			userinfo.AgreementFee = p2
			abc.SaveUserInfo(userinfo)
		}
	}
	c.String(200,  fmt.Sprintf("SUCCESS:%d", uid))
	return
}

func AgreeAgreement(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUserById(uid)
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	if user.Id == 0 || user.UserType[0:1] != "L" {
		r.Status, r.Msg = 0, golbal.Wrong[language][10102]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	//Have you signed the contract
	userinfo := abc.GetUserInfoById(uid)
	if user.IbNo != "" && userinfo.Agreement != "" && userinfo.AgreementFee != "" {
		r.Status, r.Msg = 0, golbal.Wrong[language][10107]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	//If an upper-level agent exists,
	//the contract is signed after the agent is designated
	superior := abc.GetUserIdAndUserTypeByPath(user.Path)
	var csc []abc.CommissionSetCustom
	for _, i := range superior {
		if i.UserType[:1] == "L" && uid != i.Id {
			//Whether a commission is set
			csc = abc.GetCommissionSetCustomByUid(uid, "status=1")
			if len(csc) == 0 {
				r.Status, r.Msg = 0, golbal.Wrong[language][10108]
				c.JSON(http.StatusOK, r.Response())
				return
			}
		}
	}
	code := abc.ToString(abc.RandonNumber(8))
	//Generate pdf file
	path1, path2 := pdf.RpPDF(code, user.TrueName, userinfo.Identity, user.Mobile, user.Username, user.RebateCate, csc)
	//todo 上传文件
	p1 := file.UploadFile(path1, uid)
	p2 := file.UploadFile(path2, uid)
	userinfo.Agreement = p1
	userinfo.AgreementFee = p2
	userinfo.AgreementTime = &sql.NullString{
		String: abc.FormatNow(),
		Valid:  true,
	}
	user.IbNo = code
	abc.SaveUser(user)
	abc.SaveUserInfo(userinfo)

	//Generate invitation code
	abc.CreateInviteCode(user)
	abc.AddUserLog(uid, "IB GenerationProtocol", "", abc.FormatNow(), c.ClientIP(), fmt.Sprintf("ib_no=%v path1=%v path2=%v", code, path1, path2))
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func GetMyInviteUserList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	start := c.PostForm("start")
	end := c.PostForm("end")
	UserStatus := c.PostForm("user_status")
	user := abc.GetUserById(uid)
	if user.Id == 0 {
		r := R{}
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	where := ""
	if UserStatus != "" {
		switch UserStatus {
		case "1":
			where += fmt.Sprintf(" and u.auth_status=0 and status!=-1")
		case "2":
			where += fmt.Sprintf(" and u.auth_status=1 and status!=-1 and p.count>0")
		case "3":
			where += fmt.Sprintf(" and u.auth_status=1 and status!=-1 and p.count is null")
		case "4":
			where += fmt.Sprintf(" and u.status=-1")
		case "5":
			where += fmt.Sprintf(" and u.status=1")
		}
	}
	if start != "" && end != "" {
		if len(start) <= 10 {
			start += " 00:00:00"
		}
		if len(end) <= 10 {
			end += " 23:59:59"
		}
		if abc.StringToUnix(start) > 0 && abc.StringToUnix(end) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", start, end)
		}
	}

	re := false
	if page == 1 && UserStatus == "" && start == "" && end == "" {
		re = true
	}

	r := ResponseLimit{}
	r.Status = 1

	r.Count, r.Data = abc.GetDirectInvitationList(re, page, size, user.Id, user.Path, where)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}
