package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc/conf"
	task_guest "github.com/chenqgp/abc/task/task-guest"
	"github.com/chenqgp/abc/third/excel"
	"github.com/chenqgp/abc/third/graphicValidation"
	nx "github.com/chenqgp/abc/third/sms/smsNx"
	ucloud "github.com/chenqgp/abc/third/sms/smsUcloud"
	uni "github.com/chenqgp/abc/third/sms/smsUni"
	"github.com/chenqgp/abc/third/sms/smsYimei"
	"github.com/go-playground/validator"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	login "github.com/chenqgp/abc/task/task-login"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/brevo"
	"github.com/chenqgp/abc/third/identity"
	"github.com/chenqgp/abc/third/ipAddress"
	"github.com/chenqgp/abc/third/pinyin"
	"github.com/chenqgp/abc/third/telegram"
	file "github.com/chenqgp/abc/third/uFile"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	r := &R{}

	username := c.PostForm("username")
	password := c.PostForm("password")
	language := abc.ToString(c.MustGet("language"))
	if username == "" || password == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	regex := regexp.MustCompile(`[\S]+@(\w+\.)+(\w+)`)

	if !regex.MatchString(username) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10511]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	u := abc.CheckUserExists(username)

	if u.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10002]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	//令牌桶
	limiter, done, consum := abc.Limiter(abc.ToString(u.Id), login.QueueLogin, abc.ForLimiter24hour)
	defer consum()

	if limiter.Burst() == 0 {
		//abc.UpdateUserStatus(username, -3)
		r.Status = 0
		r.Msg = golbal.Wrong[language][10001]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}
	//检查用户名，密码是否正确
	user := abc.CheckUsernameAndPassword(username, abc.Md5(password))

	if user.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][101]
		r.Data = limiter.Burst() - 1
		c.JSON(200, r.Response())

		return
	}

	if user.Status == -1 || user.Status == 0 {
		done()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10003]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	um := abc.GetUserMoreById(u.Id)
	ui := abc.GetUserInfoForId(u.Id)
	//if u.Status == -3 {
	//	return 0, golbal.Wrong[language][10001], nil
	//}

	abc.UpdateLoginNum(u.Id)
	token := abc.UserHandleToken(u.Id, abc.GetUserRole(u.UserType, u.SalesType))

	rebate := abc.GetRebateForId(user.RebateId)

	//统计交易账户余额
	//维护老的审核状态

	if u.AuthStatus == 1 {
		um.AccountStatus = 2
		um.FinancialStatus = 2
		um.TransactionStatus = 2
		um.DocumentsStatus = 2
	}

	if u.AuthStatus == -1 && (um.AccountStatus == 0 && um.FinancialStatus == 0 && um.TransactionStatus == 0 && um.DocumentsStatus == 0) {
		um.AccountStatus = 2
		um.FinancialStatus = 2
		um.TransactionStatus = 2
		um.DocumentsStatus = 3
	}

	vip := abc.UserVipUpgrade(u.Id)

	uv := abc.GetUserVip(u.Id)

	flag := false
	if u.Mobile != "" {
		flag = true
	}

	m := make(map[string]interface{})
	m["username"] = u.Username
	m["true_name"] = u.TrueName
	m["token"] = token
	m["role"] = abc.GetUserRole(u.UserType, u.SalesType)
	m["account_status"] = um.AccountStatus
	m["transaction_status"] = um.TransactionStatus
	m["financial_status"] = um.FinancialStatus
	m["documents_status"] = um.DocumentsStatus
	m["account_type"] = rebate.GroupType
	m["is_interest"] = um.ExtraDoc
	m["is_pamm"] = um.IsMam
	m["old_mobile"] = um.Mobile
	m["old_phonectcode"] = um.Phonectcode
	m["identity_type"] = ui.IdentityType
	m["reason"] = golbal.Wrong[language][abc.GetUserAudit(u.Id).Comment]
	m["vip"] = uv.Grade
	m["create_time"] = u.CreateTime
	m["upgrade"] = vip
	m["is_phone"] = flag
	m["ib_no"] = u.IbNo
	m["phonectcode"] = u.Phonectcode
	m["disable"] = abc.GetDisableFeature(u.Id, u.Path)

	r.Status = 1
	r.Msg = ""
	r.Data = m

	if r.Status == 1 {
		done()
	}

	abc.AddUserLog(u.Id, "Login", u.Email, time.Now().Format("2006-01-02 15:04:05"), c.ClientIP(), "")

	c.JSON(200, r.Response())
}

func Register(c *gin.Context) {
	r := &R{}
	username := c.PostForm("username")
	password := c.PostForm("password")
	mailCode := c.PostForm("mailCode")
	invitationCode := c.PostForm("invitationCode")
	transfer := abc.ToInt(c.PostForm("transfer"))
	phone := c.PostForm("phone")
	phoneCode := c.PostForm("phoneCode")
	areCode := c.PostForm("are_code")
	language := abc.ToString(c.MustGet("language"))

	if !abc.VerifyPasswordFormat(password, 0) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10501]

		c.JSON(200, r.Response())

		return
	}

	regex := regexp.MustCompile(`[\S]+@(\w+\.)+(\w+)`)

	if !regex.MatchString(username) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10511]

		c.JSON(200, r.Response())

		return
	}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, username)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if username == "" || password == "" || mailCode == "" || phone == "" || phoneCode == "" || areCode == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if invitationCode == "" {
		invitationCode = conf.InviteCode
	}

	if abc.CheckUserExists(username).Id > 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10005]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	captcha := abc.VerifyMailCode(username, mailCode)
	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10006]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	if !abc.PhoneIsExit(phone, areCode) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10072]

		c.JSON(200, r.Response())

		return
	}

	captcha1 := abc.VerifySmsCode(phone, phoneCode)

	if captcha1.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]

		c.JSON(200, r.Response())

		return
	}

	//验证邀请码是否正确
	i := abc.CheckInviteCode(invitationCode)

	if i.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10007]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	//邀请人信息
	inviter := abc.GetUserById(i.UserId)
	RebateId := 1
	UserType := i.Type
	CoinStatus := 1
	ParentId := i.UserId
	Path := inviter.Path
	RebateCate := 0
	SalesId := inviter.SalesId
	SomeId := 0
	if !strings.Contains(i.Type, "Level") {
		UserType = "user"
		RebateId = abc.GetAccountGroup(i.Name).Id
		CoinStatus = 0
	} else {
		RebateCate = abc.GetCommissionLevel(i.Type)
	}

	//如果邀请人是sales
	if inviter.UserType == "sales" {
		SalesId = inviter.Id
		if inviter.SalesType != "admin" {
			ParentId = inviter.ParentId
		}
	}

	//如果邀请人是用户，并且邀请的是用户
	if inviter.UserType == "user" {

		res, _ := abc.SqlOperator(`select GROUP_CONCAT(CONCAT(id,',',user_type,',',parent_id) order by find_in_set(id,?) desc) user_type from user where find_in_set(id,?)`, inviter.Path, inviter.Path)
		arr := strings.Split(strings.Trim(abc.PtoString(res.(map[string]interface{}), "user_type"), ","), ",")
		ParentId = inviter.ParentId
		for i := 0; i < len(arr); i++ {
			if i == 0 {
				if arr[i+1] == "sales" {
					SalesId, _ = strconv.Atoi(arr[i])
					break
				}
			} else {
				if arr[i] == "sales" {
					SalesId, _ = strconv.Atoi(arr[i-1])
					break
				}
				i += 2
			}

		}
	}

	//如果是同级推荐
	if inviter.UserType == i.Type && strings.Contains(inviter.UserType, "Level") {
		ParentId = inviter.ParentId
		SalesId = inviter.SalesId
		SomeId = inviter.Id
		Path = strings.ReplaceAll(inviter.Path, fmt.Sprintf(",%d,", inviter.Id), ",")
	}

	tx := abc.Tx()

	u := abc.User{
		RebateId:    RebateId,
		UserType:    UserType,
		CoinStatus:  CoinStatus,
		ParentId:    ParentId,
		SalesId:     SalesId,
		SomeId:      SomeId,
		Path:        Path,
		RebateCate:  RebateCate,
		Username:    username,
		Mobile:      phone,
		Phonectcode: areCode,
		Status:      1,
		Email:       username,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		RebateMulti: 1,
		Password:    abc.Md5(password),
		InviteCode:  invitationCode,
		LoginTime:   &sql.NullString{},
		Link:        fmt.Sprintf("%v%v", conf.PartnerLink, invitationCode), //"https://www.GOOLpartner.com/#/" + invitationCode,
	}

	var err error
	u, err = abc.CreateUser(tx, u)

	if err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10504]

		c.JSON(200, r.Response())

		return
	}
	if err := abc.UpdateUserPath(tx, u.Id, u.Path); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10505]

		c.JSON(200, r.Response())

		return
	}

	if err := abc.CreateUserRelatedTables(tx, u.Id); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10507]

		c.JSON(200, r.Response())

		return
	}
	if transfer == 1 {
		abc.CreateMoveReward(u.Id)
	}

	//统计专用
	userSta := abc.UserSta{}
	userSta.UserId = u.Id
	userSta.ParentId = u.ParentId
	userSta.Mobile = u.Mobile
	userSta.Email = u.Email
	userSta.TrueName = u.TrueName
	userSta.RegTime = u.CreateTime
	userSta.UserPath = u.Path
	userSta.Status = u.Status
	userSta.AuthStatus = u.AuthStatus
	if err := abc.CreateUserSta(tx, userSta); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10506]
		c.JSON(200, r.Response())
		return
	}

	if err := abc.UpdateCustomer(tx, u.Email); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10508]

		c.JSON(200, r.Response())

		return
	}

	//发送邮件
	go func() {
		m := abc.MailContent(1)
		content := fmt.Sprintf(m.Content, "user", username)

		brevo.Send(m.Title, content, u.Email)
	}()

	//发送站内信
	//messageConfig := abc.GetMessageConfig(1)
	//abc.SendMessage(u.Id, 1, messageConfig.ContentZh, messageConfig.ContentHk, messageConfig.ContentEn)

	tx.Commit()

	abc.DeleteSmsOrMail(captcha.Id)
	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}

// 保存用户KYC资料
func SaveUserInformation(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	temp := c.PostForm("temp")
	oldIdFront := c.PostForm("old_id_front")
	oldIdBack := c.PostForm("old_id_back")
	oldOther := c.PostForm("old_other")
	//cType := abc.ToInt(c.PostForm("type"))
	//surname := c.PostForm("surname")
	//lastname := c.PostForm("lastname")
	//nationality := c.PostForm("nationality")
	//identityType := c.PostForm("identity_type")
	//identityNum := c.PostForm("identity")
	//birthday := c.PostForm("birthday")
	//title := c.PostForm("title")
	//currencyType := c.PostForm("currency_type")
	//accountType := c.PostForm("account_type")
	//platform := c.PostForm("platform")
	//forexp := c.PostForm("forexp")
	//investfreq := c.PostForm("investfreq")
	//incomesource := c.PostForm("incomesource")
	//employment := c.PostForm("employment")
	//business := c.PostForm("business")
	//position := c.PostForm("position")
	//company := c.PostForm("company")
	//idFront := c.PostForm("id_front")
	//idBack := c.PostForm("id_back")
	//other := c.PostForm("other")
	//idType := abc.ToInt(c.PostForm("id_type"))
	//birthcountry := c.PostForm("birthcountry")
	//country := c.PostForm("country")
	//address := c.PostForm("address")
	//address_date := c.PostForm("address_date")
	//income := c.PostForm("income")
	//netasset := c.PostForm("netasset")
	//ispolitic := c.PostForm("ispolitic")
	//istax := c.PostForm("istax")
	//isusa := c.PostForm("isusa")
	//isforusa := c.PostForm("isforusa")
	//isearnusa := c.PostForm("isearnusa")
	//otherexp := c.PostForm("otherexp")
	//investaim := c.PostForm("investaim")
	//bankType := abc.ToInt(c.PostForm("bank_type"))
	//name := c.PostForm("name")
	//bankName := c.PostForm("bank_name")
	//bankNo := c.PostForm("bank_no")
	//bankAddress := c.PostForm("bank_address")
	//bankFile := c.PostForm("bank_file")
	//swift := c.PostForm("swift")
	//iban := c.PostForm("iban")
	//area := abc.ToInt(c.PostForm("area"))
	//bankCardType := abc.ToInt(c.PostForm("bank_card_type"))
	//chineseIdentity := c.PostForm("chinese_identity")

	u := abc.GetUserById(uid)
	um := abc.GetUserMore(uid)

	validate := validator.New()
	//如果审核资料已通过  就不能在填写资料
	if u.AuthStatus == 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10015]
		c.JSON(200, r.Response())

		return
	}

	var a abc.AccountInformation
	var t abc.Transaction
	var f abc.FinancialInformation
	var i abc.IdDocument
	var b abc.BankInformation

	json.Unmarshal([]byte(temp), &a)
	json.Unmarshal([]byte(temp), &t)
	json.Unmarshal([]byte(temp), &f)
	json.Unmarshal([]byte(temp), &i)
	json.Unmarshal([]byte(temp), &b)

	switch a.Type {
	//保存账户资料
	case 1:
		flag := false
		if um.AccountStatus == 0 || um.TransactionStatus == 0 || um.FinancialStatus == 0 || um.DocumentsStatus == 0 {
			flag = true
		}

		if !flag {
			if um.AccountStatus == 1 || um.AccountStatus == 2 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10503]

				c.JSON(200, r.Response())

				return
			}
		}

		if err := validate.Struct(a); err != nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(200, r.Response())

			return
		}

		if abc.GetUserIdNumber(a.Identity) || !abc.DisableIDNumber(a.Identity) {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10510]

			c.JSON(200, r.Response())

			return
		}

		if u.Phonectcode == "+86" {
			if a.Nationality != "CN" {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10517]

				c.JSON(200, r.Response())

				return
			}

			if a.IdentityType != "Identity card" {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10516]

				c.JSON(200, r.Response())

				return
			}
		}
		//if !abc.PhoneIsExit(a.OldMobile, a.OldPhonectcode) {
		//	r.Status = 0
		//	r.Msg = golbal.Wrong[language][10072]
		//
		//	c.JSON(200, r.Response())
		//
		//	return
		//}

		res, _ := abc.SqlOperator(`SELECT COUNT(u.id) count, IFNULL(u.user_type,'') user_type FROM user u LEFT JOIN user_info ui ON u.id = ui.user_id WHERE ui.identity = ? AND u.id != ?`, a.Identity, u.Id)
		if abc.ToInt(abc.PtoString(res, "count")) > 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10510]

			c.JSON(200, r.Response())

			return
		}

		if abc.PtoString(res, "user_type") != "" {
			if abc.PtoString(res, "user_type")[0:1] == u.UserType[0:1] {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10510]

				c.JSON(200, r.Response())

				return
			}
		}

		//r.Status, r.Msg, r.Data = abc.SaveAccountInformation(uid, surname, lastname, nationality, identityType, identityNum, title, birthday, birthcountry, address, country, address_date, language)

		r.Status, r.Msg, r.Data = abc.SaveAccountInformation(uid, a, language)

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v,接口名:%v保存账户资料失败", uid, "SaveUserInformation"))
			return
		}
	//保存交易信息
	case 2:
		flag := false
		if um.AccountStatus == 0 || um.TransactionStatus == 0 || um.FinancialStatus == 0 || um.DocumentsStatus == 0 {
			flag = true
		}

		if !flag {
			if um.TransactionStatus == 1 || um.TransactionStatus == 2 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10503]

				c.JSON(200, r.Response())

				return
			}
		}

		if err := validate.Struct(t); err != nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(200, r.Response())

			return
		}

		r.Status, r.Msg, r.Data = abc.SaveTransaction(uid, language, t)

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v,接口名:%v保存交易信息失败", uid, "SaveUserInformation"))
			return
		}
		//保存财务信息
	case 3:
		flag := false
		if um.AccountStatus == 0 || um.TransactionStatus == 0 || um.FinancialStatus == 0 || um.DocumentsStatus == 0 {
			flag = true
		}

		if !flag {
			if um.FinancialStatus == 1 || um.FinancialStatus == 2 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10503]

				c.JSON(200, r.Response())

				return
			}
		}

		if err := validate.Struct(f); err != nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(200, r.Response())

			return
		}

		r.Status, r.Msg, r.Data = abc.SaveFinancialInformation(uid, language, f)

		if r.Status == 0 {
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v,接口名:%v保存财务信息失败", uid, "SaveUserInformation"))
			return
		}
	case 4:
		flag := false
		if um.AccountStatus == 0 || um.TransactionStatus == 0 || um.FinancialStatus == 0 || um.DocumentsStatus == 0 {
			flag = true
		}

		if !flag {
			if um.DocumentsStatus == 1 || um.DocumentsStatus == 2 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10503]

				c.JSON(200, r.Response())

				return
			}
		}

		if err := validate.Struct(i); err != nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10000]
			c.JSON(200, r.Response())

			return
		}
		fmt.Println("=====oldIdFront=====", oldIdFront)
		//保存图片
		if oldIdFront != "" {
			i.IdFront = oldIdFront
		} else {
			frontPath := HandleFiles(c, "upload", "id_front")
			fmt.Println("=====frontPath=====", frontPath)
			if len(frontPath) != 0 {
				i.IdFront = file.UploadFile(frontPath[0], uid)
			}
		}

		fmt.Println("=====oldIdBack=====", oldIdBack)
		if oldIdBack != "" {
			i.IdBack = oldIdBack
		} else {
			backPath := HandleFiles(c, "upload", "id_back")
			fmt.Println("=====frontPath=====", backPath)
			if len(backPath) != 0 {
				i.IdBack = file.UploadFile(backPath[0], uid)
			}
		}

		fmt.Println("=====111111=====", i.IdBack)
		fmt.Println("=========222=====",i.IdFront)
		if oldOther != "" {
			i.Other = oldOther
		} else {
			otherPath := HandleFiles(c, "upload", "other")
			if len(otherPath) != 0 {
				i.Other = file.UploadFile(otherPath[0], uid)
			}
		}

		//保存身份证文件
		r.Msg = abc.UploadIdDocument(uid, language, i, u, oldIdFront, oldIdBack, oldOther)

		if r.Msg != "" {
			r.Status = 0
			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v,接口名:%v保存身份证明文件失败", uid, "SaveUserInformation"))
			return
		}
	//case 5:
	//if err := validate.Struct(b); err != nil {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10000]
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	//保存银行卡
	//_, r.Msg, r.Data = abc.SaveBankInformation(b.BankType, uid, 0, b.Name, b.BankName, b.BankNo, b.BankAddress, b.BankFile, b.Swift, b.Iban, language, b.Area, b.BankCardType, b.ChineseIdentity)
	//
	//if r.Msg != "" {
	//	log.Println("controller SaveUserInformation ")
	//}
	//
	////如果是银联  自动审核银行卡
	//if b.BankType == 1 {
	//	ui := abc.GetUserInfoForId(uid)
	//	idNum := ui.Identity
	//
	//	if ui.IdentityType != "Identity card" {
	//		idNum = ui.ChineseIdentity
	//	}
	//	res := bank.ChineseBankCard(idNum, b.Name, b.BankNo)
	//	state := -1
	//	if res.Result.RespCode == "T" {
	//		state = 1
	//		if err := abc.UpdateSql("user_info", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
	//			"chinese_name": b.Name,
	//		}); err != nil {
	//			log.Println("controller SaveUserInformation ", err)
	//		}
	//		if err := abc.UpdateSql("bank", fmt.Sprintf("id = %v", r.Data.(abc.Bank).Id), map[string]interface{}{
	//			"status": state,
	//		}); err != nil {
	//			log.Println("controller SaveUserInformation 1", err)
	//		}
	//	}
	//}
	default:
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())

		return
	}

	r.Data = nil

	usermore := abc.GetUserMoreById(uid)
	ui := abc.GetUserInfoForId(uid)

	if (usermore.AccountStatus == 1 || usermore.AccountStatus == 2) && (usermore.TransactionStatus == 1 || usermore.TransactionStatus == 2) && (usermore.FinancialStatus == 1 || usermore.FinancialStatus == 2) && (usermore.DocumentsStatus == 1 || usermore.DocumentsStatus == 2) && ui.InfoStatus != 1 {
		abc.UpdateSql("user_info", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
			"info_status": 1,
		})
	}

	if (usermore.AccountStatus == 1 || usermore.AccountStatus == 2) && (usermore.TransactionStatus == 1 || usermore.TransactionStatus == 2) && (usermore.FinancialStatus == 1 || usermore.FinancialStatus == 2) && (usermore.DocumentsStatus == 1 || usermore.DocumentsStatus == 2) && ui.IdentityType == "Identity card" {
		uf := abc.GetIdFile(uid)
		user := abc.GetUserById(uid)

		//if um.AccountStatus == 0 || um.TransactionStatus == 0 || um.DocumentsStatus == 0 || um.FinancialStatus == 0 {
		//	r.Status = 0
		//	r.Msg = golbal.Wrong[language][10009]
		//
		//	c.JSON(200, r.Response())
		//
		//	return
		//}

		//投资交易信息和财务信息通过
		if err := abc.UpdateSql("user_more", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
			"transaction_status": 2,
			"financial_status":   2,
		}); err != nil {
			log.Println("controller SaveUserInformation 2", err)
		}

		//识别身份证明文件
		res := identity.OcrIdCard(uf.FileName)
		if res.Code != "1" || strings.ToLower(ui.Identity) != strings.ToLower(res.Result.Code) {
			if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), map[string]interface{}{
				"auth_status": -1,
			}); err != nil {
				log.Println("controller SaveUserInformation 5", err)
			}

			if err := abc.UpdateSql("user_file", fmt.Sprintf("user_id = %v and file_type = 'ID'", uid), map[string]interface{}{
				"status": -1,
			}); err != nil {
				log.Println("controller SaveUserInformation 3", err)
			}

			if err := abc.UpdateSql("user_more", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"documents_status": 3,
			}); err != nil {
				log.Println("controller SaveUserInformation 4", err)
			}

			if err := abc.UpdateSql("user_audit_log", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"old": 1,
			}); err != nil {
				log.Println("controller SaveUserInformation 5", err)
			}

			//var ual abc.UserAuditLog
			//ual.UserId = uid
			//ual.CreateTime = time.Now().Format("2006-01-02 15:04:05")
			//ual.Comment = 2
			//
			//abc.CreateUserAuditLog(ual)

			r.Status = 0
			r.Msg = golbal.Wrong[language][10010]

			c.JSON(200, r.Response())

			return
		}

		//验证姓名
		pinyin := pinyin.PinYinToArray(res.Result.Name)
		trueName := strings.ReplaceAll(user.TrueName, " ", "")

		flag := false
		for _, v := range pinyin {
			name := strings.Join(v, "")
			if trueName == name {
				flag = true
			}
		}

		if !flag {
			if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), map[string]interface{}{
				"auth_status": -1,
			}); err != nil {
				log.Println("controller SaveUserInformation 5", err)
			}
			if err := abc.UpdateSql("user_more", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"documents_status": 3,
			}); err != nil {
				log.Println("controller SaveUserInformation 6", err)
			}

			if err := abc.UpdateSql("user_audit_log", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"old": 1,
			}); err != nil {
				log.Println("controller SaveUserInformation 7", err)
			}

			//var ual abc.UserAuditLog
			//ual.UserId = uid
			//ual.CreateTime = time.Now().Format("2006-01-02 15:04:05")
			//ual.Comment = 5

			//abc.CreateUserAuditLog(ual)

			r.Status = 0
			r.Msg = golbal.Wrong[language][10011]

			c.JSON(200, r.Response())

			return
		}

		//验证身份证号码
		identityNumber := identity.IdCardAuth(res.Result.Name, res.Result.Code)
		if identityNumber.Resp.Desc != "匹配" {
			if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), map[string]interface{}{
				"auth_status": -1,
			}); err != nil {
				log.Println("controller SaveUserInformation 7", err)
			}
			if err := abc.UpdateSql("user_more", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"account_status": 3,
			}); err != nil {
				log.Println("controller SaveUserInformation 8", err)
			}

			if err := abc.UpdateSql("user_audit_log", fmt.Sprintf("user_id = %v", uid), map[string]interface{}{
				"old": 1,
			}); err != nil {
				log.Println("controller SaveUserInformation 5", err)
			}

			//var ual abc.UserAuditLog
			//ual.UserId = uid
			//ual.CreateTime = time.Now().Format("2006-01-02 15:04:05")
			//ual.Comment = 1
			//
			//abc.CreateUserAuditLog(ual)

			r.Status = 0
			r.Msg = golbal.Wrong[language][10011]

			c.JSON(200, r.Response())

			return
		}

		state, msg := abc.ModifyUserProfileStatus(uid, res.Result.Name, identityNumber.Data.Address, language)

		if state == 0 {
			r.Status = 0
			r.Msg = msg

			c.JSON(200, r.Response())

			return
		}

		r.Status, r.Msg = abc.UseActivityCashVoucher(user.ParentId, u.Id, 0, 0, language)

		if r.Status == 0 {
			c.JSON(200, r.Response())

			return

		}
		account := abc.GetUserAccounts(uid)

		//开户成功增加抽奖
		tx := abc.Tx()
		if status, msg := abc.CreateLottery(tx, u.Id, "o", language, 0.0); status == -1 {
			tx.Rollback()
			telegram.SendMsg(telegram.TEXT, 0, msg)
		}
		tx.Commit()

		if u.UserType == "user" && account == "" {
			//如果审核的是用户,没有账户的话给用户增加账户申请
			abc.AddUserAccountApplication(uid, ui, language)
		}

		teleData := fmt.Sprintf("用户审核通知 ID：%d，姓名：%s，邮箱：%s，类型：%s", user.Id, user.TrueName, user.Email, user.UserType)
		telegram.SendMsg(telegram.TEXT, 4, teleData)

		if u.CreateTime >= "2023-12-01 00:00:00" && u.InviteCode == "SIVYVT" && u.CreateTime <= "2024-04-21 23:59:59" {
			messageConfig := abc.GetMessageConfig(113)
			abc.SendMessage(u.Id, 113, messageConfig.ContentZh, messageConfig.ContentHk, messageConfig.ContentEn)
		}
	}

	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}

// 上传文件
func UploadFile(c *gin.Context) {
	r := &R{}

	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))

	form, _ := c.MultipartForm()
	files := form.File["image"]
	var pathSlice []string
	flag := false
	for _, header := range files {
		name := strings.Split(header.Filename, ".")[1]
		if name == "jpg" || name == "jpeg" || name == "png" || name == "gif" {
			flag = true
		}
	}

	if flag {
		pathSlice = HandleFiles(c, "upload", "image")
	} else {
		pathSlice = HandleFilesNoPic(c, "upload", "image")
	}

	if len(pathSlice) == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10014]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	r.Data = file.UploadFile(pathSlice[0], uid)
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

// 获取ip地址
func GetIpAddress(c *gin.Context) {
	r := &R{}

	res := ipAddress.GetIpAddress(c.ClientIP())

	if res.Data.City == "中国香港" {
		res.Data.Areacode = "HK"
	}

	if res.Data.City == "中国台湾" {
		res.Data.Areacode = "TW"
	}

	if res.Data.City == "中国澳门" {
		res.Data.Areacode = "MO"
	}

	if res.Data.Country == "" {
		res.Data.Areacode = ""
	}

	r.Data = res.Data.Areacode
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

// 忘记密码
func ForgetPassword(c *gin.Context) {
	r := &R{}

	language := abc.ToString(c.MustGet("language"))
	mail := c.PostForm("mail")
	mailCode := c.PostForm("mail_code")
	//idNum := c.PostForm("id_num")
	password := c.PostForm("password")

	regex := regexp.MustCompile(`[\S]+@(\w+\.)+(\w+)`)
	if !regex.MatchString(mail) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10511]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if !abc.VerifyPasswordFormat(password, 0) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10501]

		c.JSON(200, r.Response())

		return
	}

	if mail == "" || mailCode == "" || password == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	captcha := abc.VerifyMailCode(mail, mailCode)
	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10006]

		c.JSON(200, r.Response())

		return
	}

	res, _ := abc.SqlOperator(`SELECT IFNULL(u.id, 0) id FROM user u LEFT JOIN user_info ui ON u.Id = ui.user_id WHERE u.username = ? `, mail)

	if res == nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	uid := abc.ToInt(abc.PtoString(res, "id"))

	if uid == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	m := make(map[string]interface{})
	m["password"] = abc.Md5(password)
	if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), m); err != nil {
		log.Println("controller ForgetPassword ", err)
	}
	login.QueueLogin.RemoveUser(abc.ToString(uid))
	abc.DeleteSmsOrMail(captcha.Id)
	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}

// 个人资料
func MyProfile(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res, _ := abc.SqlOperator(`SELECT u.email, u.phonectcode, um.mobile old_mobile, um.phonectcode old_phonectcode, u.mobile, ui.title, ui.surname, ui.lastname, ui.nationality, ui.identity_type, ui.identity, ui.birthday, ui.birthcountry, ui.address, ui.address_date, ui.currency_type, ui.country,
                                        ui.account_type,ui.platform, ui.forexp, ui.investfreq, ui.otherexp, um.investaim, um.incomesource, um.employment, um.business, ui.company, um.position, um.income, um.netasset, um.ispolitic, um.istax, um.isusa, um.isforusa, um.isearnusa FROM user u
                                        LEFT JOIN user_info ui
                                        ON u.id = ui.user_id
                                                LEFT JOIN user_more um
                                                ON u.id = um.user_id
                                                WHERE u.id = ?`, uid)
	s := abc.PtoString(res, "mobile")

	if s == "" {
		s = abc.PtoString(res, "old_mobile")
		res.(map[string]interface{})["phonectcode"] = abc.PtoString(res, "old_phonectcode")
	}

	if s != "" && len(s) > 5 {
		num := len(s) - 1 - 4
		ss := ""
		for i := 0; i <= num; i++ {
			ss += "*"
		}

		res.(map[string]interface{})["mobile"] = strings.ReplaceAll(s, s[2:len(s)-2], ss)
	}

	if res != nil {
		res.(map[string]interface{})["id_front"] = ""
		res.(map[string]interface{})["id_back"] = ""
		res.(map[string]interface{})["other"] = ""
		result, _ := abc.SqlOperators(`SELECT IF(front = 1,file_name,'') id_front, IF(front = 2,file_name,'') id_back, IF(front = 3,file_name,'') other FROM user_file WHERE user_id = ?`, uid)
		for _, v := range result {
			if abc.PtoString(v, "id_front") != "" {
				res.(map[string]interface{})["id_front"] = abc.PtoString(v, "id_front")
			}
			if abc.PtoString(v, "id_back") != "" {
				res.(map[string]interface{})["id_back"] = abc.PtoString(v, "id_back")
			}
			if abc.PtoString(v, "other") != "" {
				res.(map[string]interface{})["other"] = abc.PtoString(v, "other")
			}
		}
	}

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func IncomeForMonth(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))

	res := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf(`%v-%v`, "IncomeForMonth", uid))

	if res.Val() != "" {
		m := make(map[string]interface{})

		json.Unmarshal([]byte(res.Val()), &m)
		r.Status = 1
		r.Msg = ""
		r.Data = m

		c.JSON(200, r.Response())

		return
	}

	arr1 := make([]float64, 12)
	arr2 := make([]float64, 12)
	arr3 := make([]float64, 12)
	//res1, res2, res3 := abc.ObtainAnnualIncome(uid, language)
	//
	//num := int(time.Now().Month())
	//
	//arr1 = abc.ConvertMonthlyIncome(res1, num)
	//
	//arr2 = abc.ConvertMonthlyIncome(res2, 12)
	//
	//arr3 = abc.ConvertMonthlyIncome(res3, 12)

	res1, res2, res3 := abc.ObtainAnnualIncome1(uid, language)

	num := int(time.Now().Month())

	arr1 = abc.ConvertMonthlyIncome1(res1, num)

	arr2 = abc.ConvertMonthlyIncome1(res2, 12)

	arr3 = abc.ConvertMonthlyIncome1(res3, 12)

	m := make(map[string]interface{})
	m[abc.ToString(time.Now().Year())] = arr1
	m[abc.ToString(time.Now().Year()-1)] = arr2
	m[abc.ToString(time.Now().Year()-2)] = arr3

	s, _ := json.Marshal(&m)
	abc.RDB14.Set(abc.Rctx14, fmt.Sprintf("%v-%v", "IncomeForMonth", uid), string(s), 10*time.Second)

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func SendEmailCode(c *gin.Context) {
	r := &R{}
	mail := c.PostForm("mail")
	code := abc.RandonNumber(5)
	language := abc.ToString(c.MustGet("language"))

	limiter, _, consum := abc.Limiter(mail, task_guest.MailGuest, abc.ForLimiter1Minute)
	defer consum()

	if limiter.Burst() == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	go func() {
		m := abc.MailContent(15)
		content := fmt.Sprintf(m.Content, abc.ToString(code))
		brevo.Send(m.Title, content, mail)
	}()

	abc.CreateCaptcha(mail, abc.ToString(code))
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func GetUserDeposit(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	//钱包存款，取款
	u := abc.GetUserById(uid)

	start := time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02") + " 00:00:00"
	end := time.Now().AddDate(0, 1, -time.Now().Day()).Format("2006-01-02") + " 23:59:59"
	//l := len(u.Path)
	res1, _ := abc.SqlOperator(`select sum(if(amount > 0,amount,0)) a1, ABS(sum(if(amount < 0,amount,0))) a2 from payment where type = 'transfer' and status = 1 and transfer_login != 0 and comment != 'salary' and pay_time between ? and ? and user_path like ?`, start, end, u.Path+"%")
	res, _ := abc.SqlOperator(`select SUM(IF(type = 'deposit',amount+pay_fee,0)) a1, ABS(SUM(IF(type = 'withdraw',amount+pay_fee,0))) a2 from payment where status = 1 and transfer_login = 0 and pay_time between ? and ? and user_path like ?`, start, end, u.Path+"%")

	m := make(map[string]interface{})
	m["wallet_deposit"] = 0
	m["wallet_withdrawal"] = 0
	m["account_deposit"] = 0
	m["account_withdrawal"] = 0

	if res != nil {
		m["wallet_deposit"] = abc.ToFloat64(abc.PtoString(res, "a1"))
		m["wallet_withdrawal"] = abc.ToFloat64(abc.PtoString(res, "a2"))
	}

	if res1 != nil {
		m["account_deposit"] = abc.ToFloat64(abc.PtoString(res1, "a2"))
		m["account_withdrawal"] = abc.ToFloat64(abc.PtoString(res1, "a1"))
	}

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func CheckMailCode(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	username := c.PostForm("username")
	code := c.PostForm("code")
	phone := c.PostForm("phone")
	phoneCode := c.PostForm("phone_code")

	captcha := abc.VerifyMailCode(username, code)
	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10006]
		r.Data = nil
		c.JSON(200, r.Response())

		return
	}

	captcha1 := abc.VerifySmsCode(phone, phoneCode)

	if captcha1.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]

		c.JSON(200, r.Response())

		return
	}

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func InviteUsers(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	res, _ := abc.SqlOperators(`SELECT IFNULL(COUNT(id),id) num, DATE_FORMAT(create_time,'%d') d FROM user WHERE sales_id = ? AND user_type = 'user' AND MONTH(create_time) = MONTH(NOW()) GROUP BY DAY(create_time)`, uid)

	r.Status = 1
	r.Msg = ""
	r.Data = abc.MonthExchangeDay(res)

	c.JSON(200, r.Response())
}

func InvitedPartners(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	temp := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf("%v-%v", "InvitedPartners", uid))

	if temp.Val() != "" {
		arr := make([]string, 0)
		json.Unmarshal([]byte(temp.Val()), &arr)

		r.Status = 1
		r.Msg = ""
		r.Data = arr

		c.JSON(200, r.Response())

		return
	}

	res, _ := abc.SqlOperators(`SELECT COUNT(IF(user_type = 'Level Ⅰ',1,NULL)) c1, COUNT(IF(user_type = 'Level Ⅱ',1,NULL)) c2, DATE_FORMAT(create_time,'%u') d FROM user WHERE sales_id = ? AND create_time >= DATE_FORMAT(NOW(),'%Y-%m-01 00:00:00') AND create_time <= DATE_FORMAT(LAST_DAY(NOW()),'%Y-%m-%d 23:59:59') GROUP BY DATE_FORMAT(create_time,'%u')`, uid)
	result, _ := abc.SqlOperator(`SELECT DATE_FORMAT(DATE_FORMAT(NOW(),'%Y-%m-01'),'%u') f1, DATE_FORMAT(LAST_DAY(NOW()),'%u') f2`)

	arr := make([]string, 0)
	flag := false
	start := abc.ToInt(abc.PtoString(result, "f1"))
	end := abc.ToInt(abc.PtoString(result, "f2"))
	for i := start; i <= end; i++ {
		for _, v := range res {
			if i == abc.ToInt(abc.PtoString(v, "d")) {
				arr = append(arr, fmt.Sprintf("%v-%v", abc.PtoString(v, "c1"), abc.PtoString(v, "c2")))
				flag = true
				break
			}
		}

		if !flag {
			arr = append(arr, "0-0")
		}
		flag = false
	}

	s, _ := json.Marshal(&arr)
	abc.RDB14.Set(abc.Rctx14, fmt.Sprintf("%v-%v", "InvitedPartners", uid), string(s), 10*time.Second)
	r.Status = 1
	r.Msg = ""
	r.Data = arr

	c.JSON(200, r.Response())
}

// 销售邀请伙伴数量
func AdminInvitedPartners(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	u := abc.GetUserById(uid)
	res, _ := abc.SqlOperators(`select count(id) as count, user_type
from user
where left(path, ?) = ?
  and user_type != 'user'
  and user_type != 'sales' and auth_status=1
  AND create_time >= DATE_FORMAT(NOW(),'%Y-%m-01 00:00:00') AND create_time <= DATE_FORMAT(LAST_DAY(NOW()),'%Y-%m-%d 23:59:59')
group by user_type`, len(u.Path), u.Path)
	m := make(map[string]int)
	m["one_level"] = 0
	m["two_level"] = 0
	m["three_level"] = 0
	for _, i := range res {
		if abc.PtoString(i, "user_type") == "Level Ⅰ" {
			m["one_level"] = abc.ToInt(abc.PtoString(i, "count"))
			continue
		}
		if abc.PtoString(i, "user_type") == "Level Ⅱ" {
			m["two_level"] = abc.ToInt(abc.PtoString(i, "count"))
			continue
		}
		if abc.PtoString(i, "user_type") == "Level Ⅲ" {
			m["three_level"] = abc.ToInt(abc.PtoString(i, "count"))
			continue
		}
	}
	m["total"] = m["two_level"] + m["three_level"] + m["one_level"]
	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func StatisticalTransactionData(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	u := abc.GetUserById(uid)

	result := abc.RDB.Get(abc.Rctx14, fmt.Sprintf("StatisticalTransactionData-%v", uid)).Val()
	if result != "" {
		r.Status = 1
		r.Msg = ""
		r.Data = result

		c.JSON(200, r.Response())

		return
	}
	res, _ := abc.SqlOperators(`SELECT SUM(a.volume) volume, SUM(a.fee) fee, a.symbol_type symbol_type FROM (select sum(volume) volume, sum(fee) fee,symbol_type from commission  where commission_type = 0 and close_time >= DATE_FORMAT(NOW(),'%Y-%m-01 00:00:00') AND close_time <= DATE_FORMAT(LAST_DAY(NOW()),'%Y-%m-%d 23:59:59') and left(user_path,?)= ? group by SUBSTRING_INDEX(user_path ,',', FIND_IN_SET(?,user_path) + 1),symbol_type) a
						 		   GROUP BY a.symbol_type`, len(u.Path), u.Path, uid)

	m := make(map[string]float64)
	m["forex"] = 0.0
	m["metal"] = 0.0
	m["silver"] = 0.0
	m["stock_commission"] = 0.0
	m["dma"] = 0.0

	if len(res) != 0 {
		for _, v := range res {
			switch {
			case abc.ToInt(abc.PtoString(v, "symbol_type")) == 0:
				m["forex"] = abc.ToFloat64(abc.PtoString(v, "volume"))
			case abc.ToInt(abc.PtoString(v, "symbol_type")) == 1:
				m["metal"] = abc.ToFloat64(abc.PtoString(v, "volume"))
			case abc.ToInt(abc.PtoString(v, "symbol_type")) == 2:
				m["stock_commission"] = abc.ToFloat64(abc.PtoString(v, "fee"))
			case abc.ToInt(abc.PtoString(v, "symbol_type")) == 3:
				m["silver"] = abc.ToFloat64(abc.PtoString(v, "volume"))
			case abc.ToInt(abc.PtoString(v, "symbol_type")) == 4:
				m["dma"] = abc.ToFloat64(abc.PtoString(v, "volume"))
			}
		}
	}

	s, _ := json.Marshal(&m)
	abc.RDB.Set(abc.Rctx14, fmt.Sprintf("StatisticalTransactionData-%v", uid), string(s), 10*time.Second)
	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func RefreshKycStatus(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	um := abc.GetUserMore(uid)
	ua := abc.GetUserAudit(uid)

	m := make(map[string]interface{})
	m["account_status"] = um.AccountStatus
	m["transaction_status"] = um.TransactionStatus
	m["financial_status"] = um.FinancialStatus
	m["documents_status"] = um.DocumentsStatus
	m["reason"] = golbal.Wrong[language][ua.Comment]

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

func GetInvitationLink(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res, _ := abc.SqlOperators(`SELECT code, name, rights, type FROM invite_code WHERE user_id = ?`, uid)

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func UserIsBindPhone(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	u := abc.GetUserById(uid)

	flag := true
	if u.Mobile == "" {
		flag = false
	}

	r.Status = 1
	r.Msg = ""
	r.Data = flag

	c.JSON(200, r.Response())
}

func SaveBindPhone(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	areaCode := c.PostForm("area_code")
	phone := c.PostForm("phone")
	code := c.PostForm("code")

	u := abc.GetUserById(uid)
	ui := abc.GetUserInfoById(uid)

	if u.Mobile != "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10034]

		c.JSON(200, r.Response())

		return
	}

	if areaCode == "" || phone == "" || code == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if !abc.PhoneIsExit(phone, areaCode) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10072]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if ui.IdentityType == "Identity card" && areaCode != "+86" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10514]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	captcha := abc.VerifySmsCode(phone, code)
	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10033]

		c.JSON(200, r.Response())

		return
	}

	m := make(map[string]interface{})
	m["phonectcode"] = areaCode
	m["mobile"] = phone
	abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), m)

	abc.DeleteSmsOrMail(captcha.Id)

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func GetPartnerList(c *gin.Context) {
	r := &R{}
	r1 := &ResponseLimit{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	deposit := abc.ToInt(c.PostForm("deposit"))
	withdraw := abc.ToInt(c.PostForm("withdraw"))
	walletIn := abc.ToInt(c.PostForm("walletIn"))
	walletOut := abc.ToInt(c.PostForm("walletOut"))
	equity := abc.ToInt(c.PostForm("equity"))
	balance := abc.ToInt(c.PostForm("balance"))
	forex := abc.ToInt(c.PostForm("forex"))
	metal := abc.ToInt(c.PostForm("metal"))
	stockCommission := abc.ToInt(c.PostForm("stockCommission"))
	silver := abc.ToInt(c.PostForm("silver"))
	dma := abc.ToInt(c.PostForm("dma"))
	walletBalance := abc.ToInt(c.PostForm("wallet_balance"))
	userType := abc.ToInt(c.PostForm("user_type"))
	userId := abc.ToInt(c.PostForm("user_id"))
	inviter := c.PostForm("inviter")
	s := "AND left(u.user_type,1)= 'L'"
	if userId != 0 {
		uid = userId
	}

	u := abc.GetUserById(uid)

	if userType == 2 {
		s = "AND left(u.user_type,1)= 'u'"
	}

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	where := ""

	if strings.Contains(u.UserType, "Level") {
		if userType == 2 {
			where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %v %v`, u.SalesId, u.Id, s)
		} else {
			where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.SalesId, uid, s)
		}
	}

	if strings.Contains(u.UserType, "sales") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.ParentId, s)
	}

	if strings.Contains(u.UserType, "sales") && u.SalesType == "admin" {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.Id, s)
	}

	otherWhere := where
	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || inviter != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, 0, inviter, "")
	}

	if where == otherWhere && page == 1 && userType == 0 {
		temp := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "GetPartnerList", userId, size))

		if temp.Val() != "" {
			m := make(map[string]interface{})
			json.Unmarshal([]byte(temp.Val()), &m)

			c.JSON(200, m)

			return
		}
	}
	var res []interface{}
	var res1 interface{}
	var countNum int64
	order, flag := abc.SortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance)

	if where == otherWhere && userId == 0 && page == 1 && userType == 0 && !flag {
		temp := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "GetPartnerList", userId, size))

		if temp.Val() != "" {
			m := make(map[string]interface{})
			json.Unmarshal([]byte(temp.Val()), &m)

			c.JSON(200, m)

			return
		}
	}

	if userType == 2 {
		res, res1, countNum = abc.CustomList(uid, page, size, where, startTime, endTime, order, 2, otherWhere)
	} else {
		res, res1, countNum = abc.PartnerList(uid, page, size, where, startTime, endTime, order, 0)
	}

	response := r1.Response(page, size, countNum)

	m := make(map[string]interface{})
	m["list"] = res
	m["total"] = res1

	response.Status = 1
	response.Msg = ""
	response.Data = m

	if where == otherWhere && userId == 0 && page == 1 && userType == 0 && !flag {
		s, _ := json.Marshal(&response)
		abc.RDB14.Set(abc.Rctx14, fmt.Sprintf("%v-%v", "GetPartnerList", userId), string(s), 10*time.Second)
	}

	c.JSON(200, response)
}

func UserDetails(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	userId := abc.ToInt(c.PostForm("userId"))
	language := abc.ToString(c.MustGet("language"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	if userId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	user := abc.GetUserById(userId)

	if user.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	flag := true
	if user.ParentId != uid {
		flag = false
	}

	arr := strings.Split(strings.Trim(user.Path, ","), ",")

	var newArr []string
	num := 0
	for k, v := range arr {
		if abc.ToInt(v) == userId {
			num = k
			break
		}
	}

	if num == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	for i := 1; i <= 3; i++ {
		if num-i >= 0 {
			if abc.ToInt(arr[num-i]) == uid {
				break
			}
			newArr = append(newArr, arr[num-i])
		}
	}

	result, _ := abc.SqlOperators(`SELECT id, true_name FROM user WHERE FIND_IN_SET(id,?)`, strings.Join(newArr, ","))

	res, _ := abc.SqlOperator(`SELECT u.id, u.email, u.mobile, u.phonectcode, um.mobile old_mobile, um.phonectcode old_phonectcode, u.create_time, u.user_type, u.sales_type, u.wallet_balance, u.true_name, uv.grade, IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, IFNULL(COUNT(a.login),0) account_count, IFNULL(GROUP_CONCAT(a.login SEPARATOR ','),'') accounts, IFNULL(c1.forex,0.00) forex, IFNULL(c1.metal,0.00) metal, IFNULL(c1.stockCommission,0.00) stockCommission, IFNULL(c1.silver,0.00) silver, IFNULL(c1.dma,0.00) dma,
						 IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(pa.walletIn,0) walletIn, IFNULL(pa.walletOut,0) walletOut  FROM user u
						 LEFT JOIN account a
						 ON u.id = a.user_id
						 LEFT JOIN user_more um
						 ON u.id = um.user_id
						 LEFT JOIN user_vip uv
					 	 ON u.id = uv.user_id
						 LEFT JOIN 
						 (select com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume,sum(profit+storage) profit,sum(fee) fee,symbol_type FROM commission WHERE commission_type = 0 and close_time between ? and ? AND uid = ? group by symbol_type) com) c1
						 ON u.Id = c1.uid
						 LEFT JOIN
						 (select user_id, sum(profit > 0) deposit, SUM(profit < 0) withdraw from orders where cmd = 6 and close_time between ? and ? and user_id = ?) ord
						 ON u.Id = ord.user_id
						 LEFT JOIN                                                                                                    
						 (select user_id, SUM(IF(amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(amount < 0,pay_fee+amount,0)) walletOut from payment where status = 1 and transfer_login = 0 and pay_time between ? and ? and user_id = ? group by type) pa
						 ON u.Id = pa.user_id
						 WHERE u.id = ?`, startTime, endTime, userId, startTime, endTime, userId, startTime, endTime, userId, userId)

	res.(map[string]interface{})["route"] = result

	if res != nil {
		if abc.PtoString(res, "mobile") == "" {
			res.(map[string]interface{})["mobile"] = abc.PtoString(res, "old_mobile")
		}

		if abc.PtoString(res, "phonectcode") == "" {
			res.(map[string]interface{})["phonectcode"] = abc.PtoString(res, "old_phonectcode")
		}
	}

	if !flag {
		m := res.(map[string]interface{})
		delete(m, "email")
		delete(m, "mobile")
		delete(m, "phonectcode")
		delete(m, "create_time")
		delete(m, "grade")
		delete(m, "account_count")
		delete(m, "accounts")
		delete(m, "old_mobile")
		delete(m, "old_phonectcode")
	}

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func AdjustUserLevel(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	grade := abc.ToInt(c.PostForm("grade"))
	userId := abc.ToInt(c.PostForm("user_id"))

	path := abc.GetPathIb(uid)

	ok, done := abc.LimiterWait(nonConcurrent.Queue, strings.Split(path, ",")...)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if grade <= 0 || grade > 3 || userId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	userType := ""
	cate := 0
	switch grade {
	case 1:
		cate = 1
		userType = "Level Ⅰ"
	case 2:
		cate = 4
		userType = "Level Ⅱ"
	case 3:
		cate = 5
		userType = "Level Ⅲ"
	}

	if cate == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	user := abc.GetUserById(userId)
	ui := abc.GetUserInfoForId(userId)
	u := abc.GetUserById(uid)

	flag := false

	if user.UserType == "user" || strings.Contains(user.UserType, "Level") {
		flag = true
	}

	if !flag {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10509]

		c.JSON(200, r.Response())

		return
	}

	if user.UserType == "user" && abc.GetUserCountByIdentity(ui.Identity) >= 2 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10509]

		c.JSON(200, r.Response())

		return
	}

	if user.AuthStatus != 1 || u.AuthStatus != 1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10054]

		c.JSON(200, r.Response())

		return
	}

	//if user.UserType != "user" && user.IbNo == "" {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10074]
	//
	//	c.JSON(200, r.Response())
	//
	//	return
	//}

	if u.RebateCate < 4 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10077]

		c.JSON(200, r.Response())

		return
	}

	if user.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	if user.ParentId != uid {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10045]

		c.JSON(200, r.Response())

		return
	}

	arr := strings.Split(strings.Trim(user.Path, ","), ",")

	//如果不是我直邀的或者不是我创建的业务员邀请的
	if len(arr) >= 2 {
		previousUser := abc.GetUserById(abc.ToInt(arr[len(arr)-2]))

		if previousUser.UserType == "user" {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10045]

			c.JSON(200, r.Response())

			return
		}

		if previousUser.Id != uid && previousUser.ParentId != uid {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10045]

			c.JSON(200, r.Response())

			return
		}

	}

	someId := user.SomeId
	ParentId := user.ParentId
	SalesId := user.SalesId

	if user.RebateCate == cate {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10044]

		c.JSON(200, r.Response())

		return
	}

	if user.RebateCate > cate {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10046]

		c.JSON(200, r.Response())

		return
	}

	//如果调整的是用户
	tx := abc.Tx()
	if user.UserType == "user" {
		if cate != 1 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10047]

			c.JSON(200, r.Response())

			return
		}

		if u.RebateCate < 4 {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10076]
		}

		//修改我邀请人的parent_id,sales_id

		if err := abc.UpdateUserParentId(tx, userId, user.Path); err != nil {
			tx.Rollback()
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，修改user表失败", userId))
			c.JSON(200, r.Response())

			return
		}

		//删除邀请码
		if err := abc.DeleteInvite(tx, userId); err != nil {
			tx.Rollback()
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]

			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，删除邀请码失败", userId))
			return
		}

		//生成佣金模板
		if err := abc.CreateCommissionRecord(tx, user.Id, user.Path); err != nil {
			tx.Rollback()
			r.Status = 0
			r.Msg = golbal.Wrong[language][10094]

			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，生成代理佣金模板失败", userId))
			return
		}

		go func() {
			mail := abc.MailContent(77)
			conetnt := fmt.Sprintf(mail.Content, user.TrueName)

			brevo.Send(mail.Title, conetnt, user.Email)
		}()
	}

	if u.RebateCate == 2 || u.RebateCate == 3 {
		u.RebateCate = 1
	}

	//如果是同级提升
	if u.RebateCate == cate {
		someId = uid
		ParentId = u.ParentId
		SalesId = u.SalesId
		//生成新的path
		newPath := strings.ReplaceAll(u.Path, abc.ToString(u.Id), abc.ToString(user.Id))
		state, msg := abc.ReplaceUserPath(tx, newPath, user.Path, language)
		if state == 0 {
			tx.Rollback()
			r.Status = state
			r.Msg = msg

			c.JSON(200, r.Response())
			telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，同级提升，替换userpath时失败", userId))
			return
		}

		//如果调成最高等级，删除自定义佣金
		superiorUser := abc.GetUserById(ParentId)

		if !strings.Contains(superiorUser.UserType, "Level") {
			if err := abc.DeleteCommissionById(tx, userId); err != nil {
				tx.Rollback()
				telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提理等级时，调成最高等级，删除自定义佣金", userId))
				r.Status = 0
				r.Msg = golbal.Wrong[language][10048]

				c.JSON(200, r.Response())

				return
			}
		}
	}

	//提升等级后要删除邀请码，重新签订协议
	if err := abc.DeleteInvite(tx, userId); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10048]

		c.JSON(200, r.Response())
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，删除邀请码失败", userId))
		return
	}

	if err := abc.ClearUserProtocol(tx, userId); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，删除协议失败", userId))
		c.JSON(200, r.Response())

		return
	}
	if err := abc.ModifyUserParentID(tx, userId, someId, ParentId, SalesId, cate, userType); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10094]

		c.JSON(200, r.Response())

		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("uid:%v:提升等级时，更新user表中的数据失败", userId))
		return
	}
	//if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", userId), map[string]interface{}{
	//	"some_id":      someId,
	//	"parent_id":    ParentId,
	//	"sales_id":     SalesId,
	//	"user_type":    userType,
	//	"rebate_cate":  cate,
	//	"rebate_multi": 1,
	//}); err != nil {
	//	log.Println("abc AdjustUserLevel 6", err)
	//}

	tx.Commit()
	//abc.UpdateSql("token", fmt.Sprintf(`uid = %v`, userId), map[string]interface{}{
	//	"role": cate,
	//})
	abc.DeleteToken(userId)

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func EditUserPwd(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	code := c.PostForm("code")
	identity := c.PostForm("identity")
	password := c.PostForm("password")

	if !abc.VerifyPasswordFormat(password, 0) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10501]

		c.JSON(200, r.Response())

		return
	}

	if code == "" || identity == "" || password == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	captcha := abc.VerifyMailCode(u.Email, code)
	if captcha.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10006]

		c.JSON(200, r.Response())

		return
	}

	if u.Password == abc.Md5(password) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10050]

		c.JSON(200, r.Response())

		return
	}

	res, _ := abc.SqlOperator(`SELECT IFNULL(u.id, 0) id FROM user u LEFT JOIN user_info ui ON u.Id = ui.user_id WHERE u.id = ? AND RIGHT(ui.identity,4) = ?`, uid, identity)

	if res == nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	userId := abc.ToInt(abc.PtoString(res, "id"))

	if userId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	m := make(map[string]interface{})
	m["password"] = abc.Md5(password)
	if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", uid), m); err != nil {
		log.Println("controller ForgetPassword ", err)
	}

	abc.DeleteSmsOrMail(captcha.Id)
	abc.AddUserLog(uid, "Password Modify", u.Email, abc.FormatNow(), c.ClientIP(), "")
	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}

// 获取我的佣金比例
func GetUserRebateRatio(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	userId := abc.ToInt(c.PostForm("user_id"))

	if userId == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	user := abc.GetUserById(userId)

	if user.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10043]

		c.JSON(200, r.Response())

		return
	}

	if user.ParentId != uid && user.SalesId != uid {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10064]

		c.JSON(200, r.Response())

		return
	}

	//获取我的佣金权限
	my := abc.GetMyCommissionAuthority(uid, u.RebateCate, abc.GetCommissionType(uid))

	//我的直属
	other := abc.GetMyCommissionAuthority(userId, user.RebateCate, abc.GetCommissionType(uid))

	//是否是第一次调佣
	flag := false
	commission := abc.CheckCommissionReview(userId)

	if commission.Id == 0 {
		flag = true
	}

	m := make(map[string]interface{})
	m["my"] = my
	m["other"] = other
	m["first"] = flag
	m["state"] = abc.GetCommissionState(userId)
	m["times"] = abc.GetCommissionNum(userId)
	m["time"] = abc.GetCommissionPassed(userId).CreateTime

	r.Status = 1
	r.Msg = ""
	r.Data = m

	c.JSON(200, r.Response())
}

// 调整佣金
func AdjustCommission(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	userId := abc.ToInt(c.PostForm("user_id"))
	data := c.PostForm("data")

	user := abc.GetUserById(userId)
	u := abc.GetUserById(uid)

	ok, done := abc.LimiterWait(nonConcurrent.Queue, userId)
	if !ok {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]

		c.JSON(200, r.Response())

		return
	}
	defer done()

	if userId == 0 && data == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	adjustCommission := struct {
		Num0  float64 `json:"0" validate:"gte=0"`
		Num1  float64 `json:"1" validate:"gte=0"`
		Num2  float64 `json:"2" validate:"gte=0"`
		Num3  float64 `json:"3" validate:"gte=0"`
		Num4  float64 `json:"4" validate:"gte=0"`
		Num5  float64 `json:"5" validate:"gte=0"`
		Num6  float64 `json:"6" validate:"gte=0"`
		Num7  float64 `json:"7" validate:"gte=0"`
		Num8  float64 `json:"8" validate:"gte=0"`
		Num9  float64 `json:"9" validate:"gte=0"`
		Num10 float64 `json:"10" validate:"gte=0"`
		Num11 float64 `json:"11" validate:"gte=0"`
		Num12 float64 `json:"12" validate:"gte=0"`
		Num13 float64 `json:"13" validate:"gte=0"`
		Num14 float64 `json:"14" validate:"gte=0"`
	}{}

	if err := json.Unmarshal([]byte(data), &adjustCommission); err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	validate := validator.New()

	if err := validate.Struct(adjustCommission); err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	if !strings.Contains(user.UserType, "Level") {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10068]

		c.JSON(200, r.Response())

		return
	}

	if user.ParentId != uid {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10068]

		c.JSON(200, r.Response())

		return
	}

	arr := strings.Split(strings.Trim(user.Path, ","), ",")

	//如果不是我直邀的或者不是我创建的业务员邀请的
	if len(arr) >= 2 {
		previousUser := abc.GetUserById(abc.ToInt(arr[len(arr)-2]))

		if previousUser.UserType == "user" {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10045]

			c.JSON(200, r.Response())

			return
		}

		if previousUser.Id != uid && previousUser.ParentId != uid {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10045]

			c.JSON(200, r.Response())

			return
		}

	}

	//检查自己是否在审核中
	myCus := abc.CheckCommissionReview(uid)
	//检查他人是否在审核中
	otherCus := abc.CheckCommissionReview(userId)
	if myCus.Status == -1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10065]

		c.JSON(200, r.Response())

		return
	}

	if otherCus.Status == -1 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10066]

		c.JSON(200, r.Response())

		return
	}

	//被调整人的佣金比例
	other := abc.GetMyCommissionAuthority(userId, user.RebateCate, abc.GetCommissionType(uid))

	m := make(map[string]interface{})
	json.Unmarshal([]byte(data), &m)

	if len(other) != len(m) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	num := 0
	for k, v := range other {
		for kk, vv := range m {
			if k == kk && v == abc.ToFloat64(vv) {
				num++
			}
		}
	}

	if num == abc.GetCommissionType(uid)+1 {
		//r.Status = 0
		//r.Msg = golbal.Wrong[language][10067]
		//
		//c.JSON(200, r.Response())
		//
		//return
		r.Status = 1
		r.Msg = ""
		c.JSON(200, r.Response())

		return
	}

	if otherCus.CreateTime != "" {
		if otherCus.CreateTime[0:10] == time.Now().Format("2006-01-02") {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10069]

			c.JSON(200, r.Response())

			return
		}
	}

	if user.RebateCate == 0 || user.RebateCate == 5 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10070]

		c.JSON(200, r.Response())

		return
	}

	my := abc.GetMyCommissionTemplate(uid, u.RebateCate, abc.GetCommissionType(uid))

	sql := ""
	state := -1
	if otherCus.Id == 0 {
		state = 1
	}

	for _, v := range my {
		for kk, vv := range m {
			if abc.ToInt(abc.PtoString(v, "type")) == abc.ToInt(kk) {
				if abc.ToFloat64(vv) > abc.ToFloat64(abc.PtoString(v, "amount")) {
					r.Status = 0
					r.Msg = golbal.Wrong[language][10071]

					c.JSON(200, r.Response())

					return
				}
				sql += fmt.Sprintf(",(%d,'%s',%.2f,%d,'%s',%d,'%s')", userId, abc.PtoString(v, "symbol"), abc.ToFloat64(vv), abc.ToInt(kk), abc.FormatNow(), state, user.Path)
			}
		}
	}

	if len(sql) > 0 {
		err := abc.CreateCommission(sql)
		if err != nil {
			r.Status = 0
			r.Msg = golbal.Wrong[language][10100]

			c.JSON(200, r.Response())

			return
		}
	} else {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10100]

		c.JSON(200, r.Response())

		return
	}

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func CreateUserAddress(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	r := R{}
	trueName := c.PostForm("true_name")
	address := c.PostForm("address")
	phone := c.PostForm("phone")
	if trueName == "" || address == "" || phone == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	r.Data = abc.UserAddress{
		UserId:     abc.ToInt(c.MustGet("uid")),
		CreateTime: abc.FormatNow(),
		TrueName:   trueName,
		Area:       c.PostForm("area"),
		Phone:      phone,
		Zip:        c.PostForm("zip"),
		Address:    address,
	}.CreateUserAddress()
	if r.Data == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(200, r.Response())
		return
	}
	r.Status = 1
	c.JSON(200, r.Response())
}

func GetUserAddress(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	r := R{}
	r.Status, r.Data = 1, abc.GetUserAddress(fmt.Sprintf("user_id=%d", uid))
	c.JSON(200, r.Response())
}

func CreateSales(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	email := c.PostForm("email")
	phone := c.PostForm("phone")
	areaCode := c.PostForm("area_code")
	name := c.PostForm("name")
	password := c.PostForm("password")

	if email == "" || phone == "" || name == "" || password == "" || areaCode == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	regex := regexp.MustCompile(`[\S]+@(\w+\.)+(\w+)`)

	if !regex.MatchString(email) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10511]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if abc.CheckUserExists(email).Id > 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10005]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if abc.CheckUserPhone(phone, areaCode).Id > 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10072]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	if !strings.Contains(u.UserType, "Level") {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10073]

		c.JSON(200, r.Response())

		return
	}

	if u.IbNo == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10074]

		c.JSON(200, r.Response())

		return
	}

	user := abc.User{
		Username:   email,
		Password:   abc.Md5(password),
		CreateTime: abc.FormatNow(),
		Status:     1,
		AuthStatus: 1,
		LoginTime: &sql.NullString{
			String: abc.FormatNow(),
			Valid:  true,
		},
		LoginTimes:  1,
		UserType:    "sales",
		ParentId:    uid,
		SalesId:     u.SalesId,
		Email:       email,
		TrueName:    name,
		Mobile:      phone,
		Phonectcode: areaCode,
		CustomerId:  0,
		SalesType:   u.UserType,
	}

	tx := abc.Tx()
	newUser, err := abc.CreateUser(tx, user)

	if err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10504]

		c.JSON(200, r.Response())

		return
	}
	if err := abc.EditUserPth(tx, newUser.Id, u.Path); err != nil {
		tx.Rollback()
		r.Status = 0
		r.Msg = golbal.Wrong[language][10505]

		c.JSON(200, r.Response())

		return
	}

	iSlice := abc.GetMyInvite(uid)

	for _, v := range iSlice {
		if v.Type == u.UserType {
			continue
		}

		//处理邀请码
		code := ""
		for true {
			time.Sleep(100 * time.Microsecond)
			code = abc.RandStr(6)

			if abc.GetInviteIsExit(code).Id == 0 {
				break
			}
		}

		i := abc.InviteCode{
			UserId:  newUser.Id,
			Code:    code,
			Name:    v.Name,
			Rights:  u.UserType,
			Type:    v.Type,
			Comment: v.Comment,
		}

		if err := abc.CreateInvite(tx, i); err != nil {
			tx.Rollback()
			r.Status = 0
			r.Msg = golbal.Wrong[language][10506]

			c.JSON(200, r.Response())

			return
		}
	}

	abc.AddUserLog(uid, "Create Sales", u.Email, abc.FormatNow(), c.ClientIP(), email)
	r.Status = 1
	r.Msg = ""

	tx.Commit()
	c.JSON(200, r.Response())
}

func EditSalesPwd(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	Userid := abc.ToInt(c.PostForm("user_id"))
	password := c.PostForm("password")

	if Userid == 0 || password == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	user := abc.GetUserById(Userid)

	if !strings.Contains(u.UserType, "Level") {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10073]

		c.JSON(200, r.Response())

		return
	}

	if u.IbNo == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10074]

		c.JSON(200, r.Response())

		return
	}

	if user.ParentId != uid {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10075]

		c.JSON(200, r.Response())

		return
	}
	if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", Userid), map[string]interface{}{
		"password": abc.Md5(password),
	}); err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10100]

		c.JSON(200, r.Response())

		return
	}

	abc.AddUserLog(uid, "Modify Sales", u.Email, abc.FormatNow(), c.ClientIP(), user.Email)
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func EditSalesState(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	Userid := abc.ToInt(c.PostForm("user_id"))
	state := abc.ToInt(c.PostForm("state"))

	if Userid == 0 || !(state == -1 || state == 1) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)
	user := abc.GetUserById(Userid)

	if !strings.Contains(u.UserType, "Level") {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10073]

		c.JSON(200, r.Response())

		return
	}

	if u.IbNo == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10074]

		c.JSON(200, r.Response())

		return
	}

	if user.ParentId != uid {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10075]

		c.JSON(200, r.Response())

		return
	}

	if err := abc.UpdateSql("user", fmt.Sprintf("id = %v", Userid), map[string]interface{}{
		"status": state,
	}); err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10100]

		c.JSON(200, r.Response())

		return
	}

	abc.AddUserLog(uid, "Modify Sales", u.Email, abc.FormatNow(), c.ClientIP(), user.Email)
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func CustomerManagementList(c *gin.Context) {
	r := &R{}
	r1 := &ResponseLimit{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	deposit := abc.ToInt(c.PostForm("deposit"))
	withdraw := abc.ToInt(c.PostForm("withdraw"))
	walletIn := abc.ToInt(c.PostForm("walletIn"))
	walletOut := abc.ToInt(c.PostForm("walletOut"))
	equity := abc.ToInt(c.PostForm("equity"))
	balance := abc.ToInt(c.PostForm("balance"))
	forex := abc.ToInt(c.PostForm("forex"))
	metal := abc.ToInt(c.PostForm("metal"))
	stockCommission := abc.ToInt(c.PostForm("stockCommission"))
	silver := abc.ToInt(c.PostForm("silver"))
	dma := abc.ToInt(c.PostForm("dma"))
	walletBalance := abc.ToInt(c.PostForm("wallet_balance"))
	userId := abc.ToInt(c.PostForm("user_id"))
	vipGrade := abc.ToInt(c.PostForm("vip_grade"))
	inviter := c.PostForm("inviter")
	groupName := c.PostForm("group_name")
	s := "AND u.user_type= 'user'"
	//if userId != 0 {
	//	uid = userId
	//}

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	u := abc.GetUserById(uid)

	where := ""
	if strings.Contains(u.UserType, "Level") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.SalesId, u.Id, s)
	}

	if strings.Contains(u.UserType, "sales") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.ParentId, s)
	}

	otherWhere := where

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || vipGrade != 0 || inviter != "" || groupName != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, vipGrade, inviter, groupName)
	}

	order, flag := abc.SortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance)

	if where == otherWhere && page == 1 && !flag {
		temp := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "CustomerManagementList", userId, size))

		if temp.Val() != "" {
			m := make(map[string]interface{})
			json.Unmarshal([]byte(temp.Val()), &m)

			c.JSON(200, m)

			return
		}
	}

	var res []interface{}
	var res1 interface{}
	var countNum int64
	res, res1, countNum = abc.CustomList(uid, page, size, where, startTime, endTime, order, 2, otherWhere)

	response := r1.Response(page, size, countNum)

	m := make(map[string]interface{})
	m["list"] = res
	m["total"] = res1

	response.Status = 1
	response.Msg = ""
	response.Data = m

	if where == otherWhere && page == 1 && !flag {
		s, _ := json.Marshal(&response)
		abc.RDB14.Set(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "CustomerManagementList", userId, size), string(s), 10*time.Second)
	}

	c.JSON(200, response)
}

func SalesList(c *gin.Context) {
	r := &R{}
	r1 := &ResponseLimit{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	deposit := abc.ToInt(c.PostForm("deposit"))
	withdraw := abc.ToInt(c.PostForm("withdraw"))
	walletIn := abc.ToInt(c.PostForm("walletIn"))
	walletOut := abc.ToInt(c.PostForm("walletOut"))
	equity := abc.ToInt(c.PostForm("equity"))
	balance := abc.ToInt(c.PostForm("balance"))
	forex := abc.ToInt(c.PostForm("forex"))
	metal := abc.ToInt(c.PostForm("metal"))
	stockCommission := abc.ToInt(c.PostForm("stockCommission"))
	silver := abc.ToInt(c.PostForm("silver"))
	dma := abc.ToInt(c.PostForm("dma"))
	walletBalance := abc.ToInt(c.PostForm("wallet_balance"))
	userId := abc.ToInt(c.PostForm("user_id"))
	vipGrade := abc.ToInt(c.PostForm("vip_grade"))
	inviter := c.PostForm("inviter")
	groupName := c.PostForm("group_name")
	userType := abc.ToInt(c.PostForm("user_type"))

	if userId != 0 {
		uid = userId
	}
	s := ""
	switch userType {
	case 1:
		s = "AND left(u.user_type,1)= 'L'"
		break
	case 2:
		s = "AND left(u.user_type,1)= 'u'"
		break
	default:
		s = "AND u.user_type = 'sales'"
	}

	u := abc.GetUserById(uid)
	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	where := fmt.Sprintf(`u.parent_id = %d %v`, uid, s)

	//if userType == 2 && u.UserType == "user" {
	//	where = fmt.Sprintf(`left(u.path,%d)='%s' AND u.id != %v %v`, len(u.Path), u.Path, uid, s)
	//} else {
	//	where = fmt.Sprintf(`(u.sales_id = %d OR u.parent_id = %d) %v`, uid, uid, s)
	//}

	if strings.Contains(u.UserType, "sales") && strings.Contains(u.SalesType, "admin") && (userType == 1 || userType == 2) {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %v %v`, uid, uid, s)
	} else {
		if (userType == 1 || userType == 2) && strings.Contains(u.UserType, "sales") {
			where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %v %v`, uid, u.ParentId, s)
		}
	}

	if userType == 1 && strings.Contains(u.UserType, "Level") {
		where = fmt.Sprintf(`(u.sales_id = %d AND u.parent_id = %d) %v`, u.SalesId, uid, s)
	}

	if userType == 2 && strings.Contains(u.UserType, "Level") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %v %v`, u.SalesId, u.Id, s)
	}

	otherWhere := where
	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || vipGrade != 0 || inviter != "" || groupName != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, vipGrade, inviter, groupName)
	}

	order, _ := abc.SortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance)

	//if where == otherWhere && userType == 0 && page == 1 && userId == 0 && !flag {
	//	temp := abc.RDB14.Get(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "SalesList", userId, size))
	//
	//	if temp.Val() != "" {
	//		m := make(map[string]interface{})
	//		json.Unmarshal([]byte(temp.Val()), &m)
	//
	//		c.JSON(200, m)
	//
	//		return
	//	}
	//}

	var res []interface{}
	var res1 interface{}
	var countNum int64
	if userId == 0 {
		res, res1, countNum = abc.SalesmanList(uid, page, size, where, startTime, endTime, order, 1)
	} else if userId != 0 && userType == 1 {
		res, res1, countNum = abc.PartnerList(uid, page, size, where, startTime, endTime, order, 0)
	} else {
		res, res1, countNum = abc.CustomList(uid, page, size, where, startTime, endTime, order, 2, otherWhere)
	}

	response := r1.Response(page, size, countNum)

	m := make(map[string]interface{})
	m["list"] = res
	m["total"] = res1

	response.Status = 1
	response.Msg = ""
	response.Data = m

	//if where == otherWhere && userType == 0 && page == 1 && userId == 0 && !flag {
	//	s, _ := json.Marshal(&response)
	//	abc.RDB14.Set(abc.Rctx14, fmt.Sprintf("%v-%v-%v", "SalesList", userId, size), string(s), 10*time.Second)
	//}

	c.JSON(200, response)
}

func UnregisteredList(c *gin.Context) {
	r := &R{}
	r1 := &ResponseLimit{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	email := c.PostForm("email")
	start_time := c.PostForm("start_time")
	end_time := c.PostForm("end_time")

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}
	where := ""

	if email != "" {
		where += fmt.Sprintf(" AND email = '%v'", email)
	}

	if start_time != "" {
		where += fmt.Sprintf(" AND create_time >= '%v'", start_time)
	}

	if end_time != "" {
		where += fmt.Sprintf(" AND create_time <= '%v'", end_time)
	}
	r1.Data, r1.Count = abc.UnregisteredList(uid, page, size, where)
	r1.Status = 1

	c.JSON(200, r1.Response(page, size, r1.Count))
}

func MyHomePage(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res := abc.GetMyData(uid)

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func SendUserEmailCode(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))

	u := abc.GetUserById(uid)

	if u.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	code := abc.RandonNumber(5)
	m := abc.MailContent(15)
	content := fmt.Sprintf(m.Content, abc.ToString(code))

	abc.CreateCaptcha(u.Email, abc.ToString(code))
	brevo.Send(m.Title, content, u.Email)

	r.Status = 1
	r.Msg = ""
	c.JSON(200, r.Response())
}

func GetUserWalletBalance(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	user := abc.GetUser(fmt.Sprintf("id=%d", uid))

	r := &R{}
	if user.Id == 0 {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(200, r.Response())
		return
	}
	r.Data = struct {
		WalletBalance float64 `json:"wallet_balance"`
	}{user.WalletBalance}
	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func UnregisterInformation(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	code := c.PostForm("code")
	email := c.PostForm("email")

	i := abc.GetInviteCodeOne(fmt.Sprintf("code = %v", code))

	if i.Id == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10500]

		c.JSON(200, r.Response())

		return
	}

	abc.SaveUnRegister(email, i.UserId)

	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

//func GetCommissionState(c *gin.Context) {
//	r := &R{}
//	uid := abc.ToInt(c.MustGet("uid"))
//	language := abc.ToString(c.MustGet("language"))
//	userId := abc.ToInt(c.PostForm("user_id"))
//}

func CheckUserExists(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	email := c.PostForm("email")

	if email == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	regex := regexp.MustCompile(`[\S]+@(\w+\.)+(\w+)`)

	if !regex.MatchString(email) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10511]

		c.JSON(200, r.Response())

		return
	}

	flag := false
	if u := abc.CheckUserExists(email); u.Id > 0 {
		flag = true
	}

	r.Status = 1
	r.Msg = ""
	r.Data = flag

	c.JSON(200, r.Response())
}

func GetMyBalance(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))

	res := abc.GetMyAccount(uid)
	money := 0.0
	if len(res) == 0 {
		r.Status = 1
		r.Msg = ""
		r.Data = money

		c.JSON(200, r.Response())

		return
	}

	for _, v := range res {
		a := GetMt4AccountInfo(v.Login)
		if a.Code != 0 {
			money += abc.ToFloat64(a.Balance)
		}
	}

	r.Status = 1
	r.Msg = ""
	r.Data = money

	c.JSON(200, r.Response())
}

func GetSystemConfiguration(c *gin.Context) {
	r := &R{}

	r.Status = 1
	r.Msg = ""
	r.Data = abc.GetSystemConfiguration()

	c.JSON(200, r.Response())
}

func GetLiveChat(c *gin.Context) {
	r := &R{}

	r.Status = 1
	r.Msg = ""
	r.Data = abc.GetLiveChat()

	c.JSON(200, r.Response())
}

func ExportPartner(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	userId := abc.ToInt(c.PostForm("user_id"))
	inviter := c.PostForm("inviter")
	s := "AND left(u.user_type,1)= 'L'"
	if userId != 0 {
		uid = userId
	}

	u := abc.GetUserById(uid)

	where := ""

	if strings.Contains(u.UserType, "Level") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.SalesId, uid, s)
	}

	if strings.Contains(u.UserType, "sales") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.ParentId, s)
	}

	if strings.Contains(u.UserType, "sales") && u.SalesType == "admin" {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.Id, s)
	}

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || inviter != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, 0, inviter, "")
	}

	res := abc.ExportPartner(uid, where, startTime, endTime)

	str, _ := json.Marshal(&res)

	var ud []abc.UserListData
	json.Unmarshal(str, &ud)

	path := excel.PartnersExcel(ud, startTime, endTime)

	r.Status = 1
	r.Msg = ""
	r.Data = file.UploadFile(path, uid)

	c.JSON(200, r.Response())
}

func ExportSales(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	userId := abc.ToInt(c.PostForm("user_id"))
	vipGrade := abc.ToInt(c.PostForm("vip_grade"))
	inviter := c.PostForm("inviter")
	groupName := c.PostForm("group_name")

	if userId != 0 {
		uid = userId
	}

	s := "AND u.user_type = 'sales'"

	where := fmt.Sprintf(`u.parent_id = %d %v`, uid, s)
	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || vipGrade != 0 || inviter != "" || groupName != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, vipGrade, inviter, groupName)
	}

	res := abc.ExportSalesman(uid, where, startTime, endTime)

	str, _ := json.Marshal(&res)

	var ud []abc.UserListData
	json.Unmarshal(str, &ud)

	path := excel.SalespersonsExcel(ud, startTime, endTime)

	r.Status = 1
	r.Msg = ""
	r.Data = file.UploadFile(path, uid)

	c.JSON(200, r.Response())
}

func ExportCustomer(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	name := c.PostForm("name")
	grade := abc.ToInt(c.PostForm("grade"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	email := c.PostForm("email")
	vipGrade := abc.ToInt(c.PostForm("vip_grade"))
	inviter := c.PostForm("inviter")
	groupName := c.PostForm("group_name")
	s := "AND u.user_type= 'user'"

	u := abc.GetUserById(uid)
	where := ""

	if strings.Contains(u.UserType, "Level") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.SalesId, u.Id, s)
	}

	if strings.Contains(u.UserType, "sales") {
		where = fmt.Sprintf(`u.sales_id = %d AND u.parent_id = %d %v`, u.Id, u.ParentId, s)
	}
	otherWhere := where

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if name != "" || grade != 0 || startTime != "" || endTime != "" || state != 0 || email != "" || vipGrade != 0 || inviter != "" || groupName != "" {
		where = abc.GetFilterCriteria(where, name, grade, state, email, vipGrade, inviter, groupName)
	}

	res := abc.ExportCustom(where, startTime, endTime, otherWhere)

	str, _ := json.Marshal(&res)

	var ud []abc.UserListData
	json.Unmarshal(str, &ud)

	path := excel.ClientsExcel(ud, startTime, endTime)

	r.Status = 1
	r.Msg = ""
	r.Data = file.UploadFile(path, uid)

	c.JSON(200, r.Response())
}

func ExportReport(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	language := abc.ToString(c.MustGet("language"))
	userId := abc.ToInt(c.PostForm("user_id"))

	if userId != 0 {
		uid = userId
	}

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}
	res, res1 := abc.ExportReport(uid, startTime, endTime)

	r.Data = map[string]interface{}{
		"exportList": res,
		"customList": res1,
	}
	r.Status = 1
	r.Msg = ""

	c.JSON(200, r.Response())
}

func NewExportReport(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	language := abc.ToString(c.MustGet("language"))
	userId := abc.ToInt(c.PostForm("user_id"))
	deposit := abc.ToInt(c.PostForm("deposit"))
	withdraw := abc.ToInt(c.PostForm("withdraw"))
	walletIn := abc.ToInt(c.PostForm("walletIn"))
	walletOut := abc.ToInt(c.PostForm("walletOut"))
	equity := abc.ToInt(c.PostForm("equity"))
	balance := abc.ToInt(c.PostForm("balance"))
	forex := abc.ToInt(c.PostForm("forex"))
	metal := abc.ToInt(c.PostForm("metal"))
	stockCommission := abc.ToInt(c.PostForm("stockCommission"))
	silver := abc.ToInt(c.PostForm("silver"))
	dma := abc.ToInt(c.PostForm("dma"))
	walletBalance := abc.ToInt(c.PostForm("wallet_balance"))
	name := c.PostForm("name")
	r := &R{}

	if userId != 0 {
		uid = userId
	}

	u := abc.GetUserById(uid)

	where := ""

	if u.UserType == "sales" {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.ParentId, u.Id)
	} else {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.Id, u.SalesId)
	}

	if u.UserType == "sales" && u.SalesType == "admin" {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.Id, u.Id)
	}

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	where1 := where

	if name != "" {
		if name != "" {
			if !strings.Contains(name, "'") {
				where += fmt.Sprintf(" AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v' OR FIND_IN_SET('%v',account.login))", strings.ReplaceAll(name, " ", ""), strings.ReplaceAll(name, " ", ""), strings.ReplaceAll(name, " ", ""))
			}
		}
	}

	order, flag := abc.NewSortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance)

	r.Status = 1
	r.Msg = ""
	r.Data = map[string]interface{}{
		"list": abc.NewExportReport(uid, where, startTime, endTime, order, flag, name, where1),
		"id":   uid,
	}

	c.JSON(200, r.Response())
}

func SendPhoneCode(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	phone := c.PostForm("phone")
	areaCode := c.PostForm("area_code")
	code := abc.ToString(abc.RandonNumber(5))

	limiter, done, consum := abc.Limiter(phone, task_guest.PhoneGuest, abc.ForLimiter1Minute)
	defer consum()

	if limiter.Burst() == 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10119]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	state := abc.CheckPhoneStatus(phone)

	if state == 2 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10522]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if state == 3 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10523]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	state1 := abc.PhoneIsDisabled(phone)

	if state1 == 2 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10522]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	if state1 == 3 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10523]
		r.Data = nil

		c.JSON(200, r.Response())

		return
	}

	prompt := ""

	if abc.InvalidVerificationCode(phone) {
		prompt = golbal.Wrong[language][10527]
	}

	newPhone := phone
	if areaCode == "+86" || areaCode == "86" {
		res, _ := abc.SqlOperator("SELECT GROUP_CONCAT(`key`) sms_name FROM config WHERE `name` = 'SMS_CHINA'")
		if res != nil {
			arr := strings.Split(abc.PtoString(res, "sms_name"), ",")
			name := arr[rand.Intn(len(arr))]
			if name == "uni" {
				uni := uni.SendSmsUni(phone, code)
				if uni.Code != "0" {
					r.Status = 0
					r.Msg = golbal.Wrong[language][10026]

					c.JSON(200, r.Response())
					done()
					return
				}
			} else if name == "yimei" {
				yimei := smsYimei.SendSmsYimei(phone, code)

				if yimei.Code != "SUCCESS" {
					r.Status = 0
					r.Msg = golbal.Wrong[language][10026]

					c.JSON(200, r.Response())
					done()
					return
				}
			}
		}
		//phone = strings.ReplaceAll(areaCode, "+", "") + phone
		//m := ucloud.SendSmsUcloud(phone, code)
		//
		//if m.RetCode != 0 {
		//	r.Status = 0
		//	r.Msg = golbal.Wrong[language][10026]
		//
		//	c.JSON(200, r.Response())
		//	done()
		//	return
		//}
	} else {
		//获取短信通道
		config := abc.GetSmsChannel()
		if config.Key == "nx" {
			phone = strings.ReplaceAll(areaCode, "+", "") + phone
			nx := nx.SendSmsNx(phone, code)
			if nx.Code != "0" {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10026]

				c.JSON(200, r.Response())
				done()
				return
			}
		} else if config.Key == "uni" {
			phone = strings.ReplaceAll(areaCode, "+", "") + phone
			uni := uni.SendSmsUni(phone, code)
			if uni.Code != "0" {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10026]

				c.JSON(200, r.Response())
				done()
				return
			}
		} else if config.Key == "ucloud" {
			phone = "(" + strings.ReplaceAll(areaCode, "+", "") + ")" + phone
			log.Println(phone)
			m := ucloud.SendSmsUcloudGlobal(phone, code)
			if m.RetCode != 0 {
				r.Status = 0
				r.Msg = golbal.Wrong[language][10026]

				c.JSON(200, r.Response())
				done()
				return
			}
		}
	}

	abc.CreateCapture(newPhone, code, abc.FormatNow(), time.Now().Unix(), 1)

	m := make(map[string]interface{})
	m["prompt"] = prompt

	r.Status = 1
	r.Msg = ""
	r.Data = m
	c.JSON(200, r.Response())

}

func CheckPhoneExists(c *gin.Context) {
	r := &R{}
	phone := c.PostForm("phone")
	areCode := c.PostForm("are_code")

	flag := false
	if !abc.PhoneIsExit(phone, areCode) {
		flag = true
	}

	r.Status = 1
	r.Msg = ""
	r.Data = flag

	c.JSON(200, r.Response())
}

func ProxyRelationshipList(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	userId := abc.ToInt(c.PostForm("user_id"))

	if userId != 0 {
		uid = userId
	}

	u := abc.GetUserById(uid)
	where := ""

	if u.UserType == "sales" {
		where = fmt.Sprintf("(u.parent_id = %v and u.sales_id = %v)", u.ParentId, u.Id)
	} else {
		where = fmt.Sprintf("(u.parent_id = %v)", u.Id)
	}

	if u.UserType == "sales" && u.SalesType == "admin" {
		where = fmt.Sprintf("(u.parent_id = %v and u.sales_id = %v)", u.Id, u.Id)
	}

	r.Status = 1
	r.Msg = ""
	r.Data = abc.ProxyRelationshipList(where, uid)

	c.JSON(200, r.Response())
}


func SearchForName(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	name := c.PostForm("name")

	u := abc.GetUserById(uid)

	path := u.Path + "%"
	res, _ := abc.SqlOperators(`SELECT id, true_name, user_type FROM user WHERE REPLACE(true_name,' ','') = ? and path like ?`, strings.ReplaceAll(name, " ", ""), path)

	r.Status = 1
	r.Msg = ""
	r.Data = res

	c.JSON(200, r.Response())
}

func SearchReport(c *gin.Context) {
	uid := abc.ToInt(c.MustGet("uid"))
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	language := abc.ToString(c.MustGet("language"))
	deposit := abc.ToInt(c.PostForm("deposit"))
	withdraw := abc.ToInt(c.PostForm("withdraw"))
	walletIn := abc.ToInt(c.PostForm("walletIn"))
	walletOut := abc.ToInt(c.PostForm("walletOut"))
	equity := abc.ToInt(c.PostForm("equity"))
	balance := abc.ToInt(c.PostForm("balance"))
	forex := abc.ToInt(c.PostForm("forex"))
	metal := abc.ToInt(c.PostForm("metal"))
	stockCommission := abc.ToInt(c.PostForm("stockCommission"))
	silver := abc.ToInt(c.PostForm("silver"))
	dma := abc.ToInt(c.PostForm("dma"))
	walletBalance := abc.ToInt(c.PostForm("wallet_balance"))
	name := c.PostForm("name")
	r := &R{}

	if strings.Contains(name, "'") {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())
	}

	user := abc.GetUserById(uid)

	account := abc.GetAccountOne(fmt.Sprintf("login = '%v'", name))
	u1 := abc.GetUserById(account.UserId)

	u2 := abc.GetUser(fmt.Sprintf("REPLACE(true_name,' ','') = '%v' OR REPLACE(email,' ','') = '%v'", strings.ReplaceAll(name, " ", ""), strings.ReplaceAll(name, " ", "")))

	if (u2.Id == 0 && u1.Id == 0) || (!strings.Contains(user.Path, abc.ToString(u2.Id)) && !strings.Contains(user.Path, abc.ToString(u1.Id))) {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u3 := u1.Id

	if u2.Id != 0 {
		u3 = u2.Id
	}

	u := abc.GetUserById(u3)
	where := ""

	if u.UserType == "sales" {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.ParentId, u.Id)
	} else {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.Id, u.SalesId)
	}

	if u.UserType == "sales" && u.SalesType == "admin" {
		where = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.Id, u.Id)
	}

	if startTime == "" && endTime == "" {
		startTime = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endTime = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	count, err := abc.GetDaysBetweenDate("2006-01-02", startTime, endTime)

	if err != nil {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	if count > 31 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10091]
		r.Data = nil

		c.JSON(200, r.Response())
		return
	}

	order, flag := abc.NewSortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance)

	r.Status = 1
	r.Msg = ""
	r.Data = map[string]interface{}{
		"list": abc.NewExportReport(u3, where, startTime, endTime, order, flag, name, where),
		"id":   u3,
	}

	c.JSON(200, r.Response())
}

func NewGetPartnerList(c *gin.Context) {
	r1 := &ResponseLimit{}
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	content := c.PostForm("content")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	userType := abc.ToInt(c.PostForm("user_type"))
	userState := abc.ToInt(c.PostForm("user_state"))

	u := abc.GetUserById(uid)
	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
	}

	s := fmt.Sprintf("u.parent_id = %v", uid)

	if strings.Contains(u.UserType, "sales") {
		s = fmt.Sprintf("u.parent_id = %v and u.sales_id = %v", u.ParentId, u.Id)
	}

	if content != "" {
		content = fmt.Sprintf(` AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v' OR u.mobile = '%v' OR FIND_IN_SET('%v',a.login))`, strings.ReplaceAll(content, " ", ""), content, content, content)
	}

	where := ""
	if startTime != "" {
		where += fmt.Sprintf(" AND u.create_time >= '%v'", startTime+" 00:00:00")
	}
	if endTime != "" {
		where += fmt.Sprintf(" AND u.create_time <= '%v'", endTime+" 23:59:59")
	}

	if state != 0 {
		switch state {
		case 1:
			where += fmt.Sprintf(` AND u.auth_status = 0`)
		case 2:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn != 0`)
		case 3:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn IS NULL`)
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		case 6:
			where += fmt.Sprintf(` AND u.auth_status = -1`)
		}
	}

	if userState != 0 {
		switch userState {
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		}
	}

	if userType != 0 {
		switch userType {
		case 1:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅰ")
		case 2:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅱ")
		case 3:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅲ")
		default:

		}
	}

	res, _ := abc.SqlOperators(fmt.Sprintf(`SELECT u.id, u.true_name, u.email, u.mobile, u.user_type, ui.grade, u.status, u.auth_status, u.create_time, u.path, u1.true_name inviter_name, a.login_count, a.login, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                                   LEFT JOIN user_vip ui
                                                   ON u.id = ui.user_id
                                                   LEFT JOIN user u1
                                                   ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',',-3),',',1) = u1.id
                                                   LEFT JOIN (SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
												   LEFT JOIN account a
  												   ON u.id = a.user_id
												   WHERE %v AND LEFT(u.user_type,1) = 'L'
											       GROUP BY user_id) a
                                                   ON u.id = a.id
												   LEFT JOIN
                                                   (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE %v AND LEFT(u.user_type,1) = 'L')) GROUP BY p.user_id) p1
                                                   ON u.id = p1.user_id
                                                   WHERE %v AND LEFT(u.user_type,1) = 'L' %v %v
                                                   ORDER BY u.create_time DESC LIMIT %v,%v`, s, s, s, content, where, (page-1)*size, size))

	count, _ := abc.SqlOperator(fmt.Sprintf(`SELECT count(u.id) count FROM user u
                                                   LEFT JOIN user_vip ui
                                                   ON u.id = ui.user_id
                                                   LEFT JOIN user u1
                                                   ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',',-3),',',1) = u1.id
                                                   LEFT JOIN
                                                   (SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
												   LEFT JOIN account a
  												   ON u.id = a.user_id
												   WHERE %v AND LEFT(u.user_type,1) = 'L'
											       GROUP BY user_id) a
                                                   ON u.id = a.id
												   LEFT JOIN
                                                   (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE %v AND LEFT(u.user_type,1) = 'L')) GROUP BY p.user_id) p1
                                                   ON u.id = p1.user_id
                                                   WHERE %v AND LEFT(u.user_type,1) = 'L' %v %v`, s, s, s, content, where))
	var total int64

	if count != nil {
		total = abc.ToInt64(abc.PtoString(count, "count"))
	}

	var arr []string
	for _, v := range res {

		arr = append(arr, abc.PtoString(v, "id"))

		if abc.ToInt(abc.PtoString(v, "status")) == -1 || abc.ToInt(abc.PtoString(v, "status")) == 0 {
			v.(map[string]interface{})["user_status"] = 4
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 && abc.ToFloat64(abc.PtoString(v, "walletIn")) != 0 {
			v.(map[string]interface{})["user_status"] = 2
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 {
			v.(map[string]interface{})["user_status"] = 3
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 0 || abc.ToInt(abc.PtoString(v, "auth_status")) == -1 {
			v.(map[string]interface{})["user_status"] = 1
			continue
		}
	}

	result, _ := abc.SqlOperators(`SELECT * FROM (SELECT comment,user_id FROM user_audit_log WHERE FIND_IN_SET(user_id,?) and old = 0 ORDER BY create_time DESC) t GROUP BY t.user_id`, strings.Join(arr, ","))

	for _, v := range res {
		v.(map[string]interface{})["reason"] = ""
		for _, vv := range result {
			if abc.PtoString(v, "id") == abc.PtoString(vv, "user_id") {
				v.(map[string]interface{})["reason"] = golbal.Wrong[language][abc.ToInt(abc.PtoString(vv, "comment"))]
			}
		}
	}

	r1.Data = res
	r1.Status = 1

	c.JSON(200, r1.Response(page, size, total))
}

func NewUserManagementList(c *gin.Context) {
	r1 := &ResponseLimit{}
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	content := c.PostForm("content")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	vipGrade := abc.ToInt(c.PostForm("vip_grade"))
	groupName := c.PostForm("group_name")
	userState := abc.ToInt(c.PostForm("user_state"))

	u := abc.GetUser(fmt.Sprintf("id = %v", uid))

	parentId := u.Id
	salesId := u.SalesId

	if strings.Contains(u.UserType, "sales") {
		parentId = u.ParentId
		salesId = u.Id
	}

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
	}

	if content != "" {
		content = fmt.Sprintf(` AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v' OR u.mobile = '%v' OR FIND_IN_SET('%v',a.login))`, strings.ReplaceAll(content, " ", ""), content, content, content)
	}

	where := ""
	if startTime != "" {
		where += fmt.Sprintf(" AND u.create_time >= '%v'", startTime+" 00:00:00")
	}
	if endTime != "" {
		where += fmt.Sprintf(" AND u.create_time <= '%v'", endTime+" 23:59:59")
	}

	if vipGrade != 0 {
		where += fmt.Sprintf(` AND uv.grade = %v`, vipGrade)
	}

	if state != 0 {
		switch state {
		case 1:
			where += fmt.Sprintf(` AND u.auth_status = 0`)
		case 2:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn != 0`)
		case 3:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn IS NULL`)
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		case 6:
			where += fmt.Sprintf(` AND u.auth_status = -1`)
		}
	}

	if userState != 0 {
		switch userState {
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		}
	}

	if groupName != "" {
		if !strings.Contains(groupName, "'") {
			where += fmt.Sprintf(` AND rc.group_name = '%v'`, groupName)
		}
	}

	res, _ := abc.SqlOperators(fmt.Sprintf(`SELECT u.user_type, u.id, u.true_name, u.email, u.mobile, uv.grade, u.status, u.auth_status, u.create_time, u.path, u1.true_name inviter_name, a.login_count, a.login, rc.group_name, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut  FROM user u
								    LEFT JOIN user_vip uv
									ON u.id = uv.user_id
									LEFT JOIN rebate_config rc
								    ON u.rebate_id = rc.id
									LEFT JOIN user u1
									ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',',-3),',',1) = u1.id
									LEFT JOIN 
									(SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
									LEFT JOIN account a
  									ON u.id = a.user_id
									WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u'
									GROUP BY user_id) a
                                    ON u.id = a.id
									LEFT JOIN 
                                    (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u')) GROUP BY p.user_id) p1
                                    ON u.id = p1.user_id                                                                                                                                     		
									WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u' %v %v
									ORDER BY u.create_time DESC LIMIT %v,%v`, parentId, salesId, parentId, salesId, parentId, salesId, content, where, (page-1)*size, size))

	count, _ := abc.SqlOperator(fmt.Sprintf(`SELECT count(u.id) count FROM user u
								    LEFT JOIN user_vip uv
									ON u.id = uv.user_id
									LEFT JOIN rebate_config rc
								    ON u.rebate_id = rc.id
									LEFT JOIN user u1
									ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',',-3),',',1) = u1.id
									LEFT JOIN 
									(SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
									LEFT JOIN account a
  									ON u.id = a.user_id
									WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u'
									GROUP BY user_id) a
                                    ON u.id = a.id
									LEFT JOIN 
                                    (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u')) GROUP BY p.user_id) p1
                                    ON u.id = p1.user_id                                                                                                                                     		
									WHERE u.parent_id = %v AND u.sales_id = %v AND LEFT(u.user_type,1) = 'u' %v %v`, parentId, salesId, parentId, salesId, parentId, salesId, content, where))

	var total int64

	if count != nil {
		total = abc.ToInt64(abc.PtoString(count, "count"))
	}

	var arr []string
	for _, v := range res {

		arr = append(arr, abc.PtoString(v, "id"))

		if abc.ToInt(abc.PtoString(v, "status")) == -1 || abc.ToInt(abc.PtoString(v, "status")) == 0 {
			v.(map[string]interface{})["user_status"] = 4
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 && abc.ToFloat64(abc.PtoString(v, "walletIn")) != 0 {
			v.(map[string]interface{})["user_status"] = 2
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 {
			v.(map[string]interface{})["user_status"] = 3
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 0 || abc.ToInt(abc.PtoString(v, "auth_status")) == -1 {
			v.(map[string]interface{})["user_status"] = 1
			continue
		}
	}

	result, _ := abc.SqlOperators(`SELECT * FROM (SELECT comment,user_id FROM user_audit_log WHERE FIND_IN_SET(user_id,?) and old = 0 ORDER BY create_time DESC) t GROUP BY t.user_id`, strings.Join(arr, ","))

	for _, v := range res {
		v.(map[string]interface{})["reason"] = ""
		for _, vv := range result {
			if abc.PtoString(v, "id") == abc.PtoString(vv, "user_id") {
				v.(map[string]interface{})["reason"] = golbal.Wrong[language][abc.ToInt(abc.PtoString(vv, "comment"))]
			}
		}
	}

	r1.Data = res
	r1.Status = 1

	c.JSON(200, r1.Response(page, size, total))
}

func NewSalesList(c *gin.Context) {
	r1 := &ResponseLimit{}
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	content := c.PostForm("content")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	start := c.PostForm("start")
	end := c.PostForm("end")
	id := abc.ToInt(c.PostForm("id"))

	u := abc.GetUser(fmt.Sprintf("id = %v", uid))

	newParentId := 0

	parentId := 0
	salesId := 0

	if strings.Contains(u.UserType, "sales") && u.SalesType == "admin" {
		parentId = u.Id
		salesId = u.Id
	} else {
		parentId = u.Id
		salesId = u.SalesId
	}

	if strings.Contains(u.UserType, "L") || (strings.Contains(u.UserType, "sales") && u.SalesType == "admin") {
		newParentId = u.Id
	} else if strings.Contains(u.UserType, "sales") && u.SalesType != "admin" {
		newParentId = u.ParentId
	}

	if page <= 0 || size <= 0 {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]
	}

	if content != "" {
		content = fmt.Sprintf(` AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v')`, strings.ReplaceAll(content, " ", ""), content)
	}

	where := ""
	if startTime != "" {
		where += fmt.Sprintf(" AND u.create_time >= '%v'", startTime+" 00:00:00")
	}
	if endTime != "" {
		where += fmt.Sprintf(" AND u.create_time <= '%v'", endTime+" 23:59:59")
	}

	if state != 0 {
		switch state {
		case 1:
			where += fmt.Sprintf(` AND u.auth_status = 0`)
		case 2:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn != 0`)
		case 3:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn IS NULL`)
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		case 6:
			where += fmt.Sprintf(` AND u.auth_status = -1`)
		}
	}

	if id != 0 {
		parentId = id
		salesId = id
		newParentId = id
	}
	res, _ := abc.SqlOperators(fmt.Sprintf(`SELECT u.id, u.true_name, u.email, u.status, u.mobile, u.create_time, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                                   LEFT JOIN 
                                    			   (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales')) GROUP BY p.user_id) p1
                                    				ON u.id = p1.user_id  
												   WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales' %v %v ORDER BY u.create_time DESC LIMIT %v,%v`, parentId, salesId, parentId, salesId, content, where, (page-1)*size, size))

	count, _ := abc.SqlOperator(fmt.Sprintf(`SELECT count(u.id) count FROM user u
                                                   LEFT JOIN 
                                    			   (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales')) GROUP BY p.user_id) p1
                                    				ON u.id = p1.user_id  
												   WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales' %v %v`, parentId, salesId, parentId, salesId, content, where))
	var total int64

	if count != nil {
		total = abc.ToInt64(abc.PtoString(count, "count"))
	}

	var arr []string
	for _, v := range res {
		arr = append(arr, abc.PtoString(v, "id"))
		if abc.ToInt(abc.PtoString(v, "status")) == -1 || abc.ToInt(abc.PtoString(v, "status")) == 0 {
			v.(map[string]interface{})["user_status"] = 4
			continue
		}

		if abc.ToInt(abc.PtoString(v, "status")) == 1 {
			v.(map[string]interface{})["user_status"] = 5
			continue
		}
	}

	where1 := ""

	if start != "" || end != "" {
		where1 += fmt.Sprintf(" AND auth_status = 1")
	}

	if start != "" {
		where1 += fmt.Sprintf(" AND create_time >= '%v'", start+" 00:00:00")
	}

	if end != "" {
		where1 += fmt.Sprintf(" AND create_time <= '%v'", end+" 23:59:59")
	}

	var res1, res2 []interface{}

	if strings.Contains(u.UserType, "sales") && u.SalesType == "admin" {
		res1, _ = abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id, sales_id FROM user WHERE FIND_IN_SET(sales_id,'%v') AND LEFT(user_type,1) = 'L' AND parent_id = sales_id %v GROUP BY sales_id`, strings.Join(arr, ","), where1))
		res2, _ = abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id, sales_id FROM user WHERE FIND_IN_SET(sales_id,'%v') AND user_type = 'user' AND parent_id = sales_id %v GROUP BY sales_id`, strings.Join(arr, ","), where1))
	} else {
		res1, _ = abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id, sales_id FROM user WHERE parent_id = %v AND FIND_IN_SET(sales_id,'%v') AND LEFT(user_type,1) = 'L' %v GROUP BY sales_id`, newParentId, strings.Join(arr, ","), where1))
		res2, _ = abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id, sales_id FROM user WHERE parent_id = %v AND FIND_IN_SET(sales_id,'%v') AND user_type = 'user' %v GROUP BY sales_id`, newParentId, strings.Join(arr, ","), where1))
	}

	for _, v := range res {
		v.(map[string]interface{})["level_count"] = 0
		if res1 != nil {
			for _, vv := range res1 {
				if abc.PtoString(v, "id") == abc.PtoString(vv, "sales_id") {
					v.(map[string]interface{})["level_count"] = abc.ToInt(abc.PtoString(vv, "count"))
					break
				}
			}
		}
	}

	for _, v := range res {
		v.(map[string]interface{})["user_count"] = 0
		for _, vv := range res2 {
			if res2 != nil {
				if abc.PtoString(v, "id") == abc.PtoString(vv, "sales_id") {
					v.(map[string]interface{})["user_count"] = abc.ToInt(abc.PtoString(vv, "count"))
					break
				}
			}
		}
	}

	r1.Data = res
	r1.Status = 1

	c.JSON(200, r1.Response(page, size, total))
}

func SalesRelationshipDiagram(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))

	if id != 0 {
		uid = id
	}

	res, _ := abc.SqlOperators(`SELECT id, true_name, user_type FROM user WHERE sales_type = 'admin' AND parent_id = ? AND sales_id = ?`, uid, uid)

	r.Status = 1
	r.Data = res

	c.JSON(200, r.Response())
}

func SubordinateInformation(c *gin.Context) {
	//r1 := &ResponseLimit{}
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	//page := abc.ToInt(c.PostForm("page"))
	//size := abc.ToInt(c.PostForm("size"))
	content := c.PostForm("content")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	state := abc.ToInt(c.PostForm("state"))
	start := c.PostForm("start")
	end := c.PostForm("end")
	userType := abc.ToInt(c.PostForm("user_type"))
	userState := abc.ToInt(c.PostForm("user_state"))
	userId := abc.ToInt(c.PostForm("user_id"))

	if content != "" {
		user := abc.GetUserById(uid)
		res, _ := abc.SqlOperator(fmt.Sprintf(`SELECT u.id, u.parent_id FROM user u
							 LEFT JOIN account a
							 ON u.id = a.user_id
							 WHERE u.path LIKE '%v' AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v' OR u.mobile = '%v' OR FIND_IN_SET('%v',a.login))`, user.Path+"%", strings.ReplaceAll(content, " ", ""), strings.ReplaceAll(content, " ", ""), strings.ReplaceAll(content, " ", ""), strings.ReplaceAll(content, " ", "")))

		if res != nil {
			userId = abc.ToInt(abc.PtoString(res, "parent_id"))
		}
	}

	if userId != 0 {
		uid = userId
	}

	u := abc.GetUserById(uid)

	arrLength := strings.Split(u.Path, ",")

	parentId := 0
	salesId := 0

	if u.UserType == "sales" {
		parentId = u.ParentId
		salesId = u.Id
	} else {
		parentId = u.Id
		salesId = u.SalesId
	}

	if u.UserType == "sales" && u.SalesType == "admin" {
		parentId = u.Id
		salesId = u.Id
	}

	//newParentId := 0
	//if strings.Contains(u.UserType, "L") || (strings.Contains(u.UserType, "sales") && u.SalesType == "admin") {
	//	newParentId = u.Id
	//} else if strings.Contains(u.UserType, "sales") && u.SalesType != "admin" {
	//	newParentId = u.ParentId
	//}

	//if page <= 0 || size <= 0 {
	//	r.Status = 0
	//	r.Msg = golbal.Wrong[language][10000]
	//}

	if content != "" {
		content = fmt.Sprintf(` AND (REPLACE(u.true_name,' ','') = '%v' OR u.email = '%v' OR u.mobile = '%v' OR FIND_IN_SET('%v',a.login))`, strings.ReplaceAll(content, " ", ""), content, content, content)
	}

	where := ""
	if startTime != "" {
		where += fmt.Sprintf(" AND u.create_time >= '%v'", startTime+" 00:00:00")
	}
	if endTime != "" {
		where += fmt.Sprintf(" AND u.create_time <= '%v'", endTime+" 23:59:59")
	}

	if state != 0 {
		switch state {
		case 1:
			where += fmt.Sprintf(` AND u.auth_status = 0`)
		case 2:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn != 0`)
		case 3:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn IS NULL`)
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		case 6:
			where += fmt.Sprintf(` AND u.auth_status = -1`)
		}
	}

	if userState != 0 {
		switch userState {
		case 4:
			where += fmt.Sprintf(` AND (u.status = -1 OR u.status = 0)`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		}
	}

	if userType != 0 {
		switch userType {
		case 1:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅰ")
		case 2:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅱ")
		case 3:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅲ")
		case 4:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "user")
		case 5:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "sales")
		default:

		}
	}

	var res []interface{}

	if content != "" || where != "" {
		res, _ = abc.SqlOperators(fmt.Sprintf(`SELECT u.user_type, u.sales_type, u.id, u.true_name, u.email, u.mobile, u.phonectcode, u.login_time, uv.grade, u.status, u.auth_status, u.create_time, u.path, a.login_count, a.login, rc.group_name, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
								    LEFT JOIN user_vip uv
									ON u.id = uv.user_id
									LEFT JOIN rebate_config rc
								    ON u.rebate_id = rc.id
									LEFT JOIN 
									(SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
									LEFT JOIN account a
  									ON u.id = a.user_id
									GROUP BY user_id) a
                                    ON u.id = a.id
                                    LEFT JOIN
                                    (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 GROUP BY p.user_id) p1
                                    ON u.id = p1.user_id                                                                                                                                                   
									WHERE u.path like '%v' %v %v
									ORDER BY u.create_time DESC`, u.Path+"%", content, where))
	} else {
		res, _ = abc.SqlOperators(fmt.Sprintf(`SELECT u.user_type, u.sales_type, u.id, u.true_name, u.email, u.mobile, u.phonectcode, u.login_time, uv.grade, u.status, u.auth_status, u.create_time, u.path, a.login_count, a.login, rc.group_name, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
								    LEFT JOIN user_vip uv
									ON u.id = uv.user_id
									LEFT JOIN rebate_config rc
								    ON u.rebate_id = rc.id
									LEFT JOIN 
									(SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
									LEFT JOIN account a
  									ON u.id = a.user_id
									WHERE u.parent_id = %v AND u.sales_id = %v
									GROUP BY user_id) a
                                    ON u.id = a.id
                                    LEFT JOIN
                                    (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v)) GROUP BY p.user_id) p1
                                    ON u.id = p1.user_id                                                                                                                                                   
									WHERE u.parent_id = %v AND u.sales_id = %v %v %v
									ORDER BY u.create_time DESC`, parentId, salesId, parentId, parentId, salesId, content, where))
	}

	//count, _ := abc.SqlOperator(fmt.Sprintf(`SELECT COUNT(u.id) count FROM user u
	//							    LEFT JOIN user_vip uv
	//								ON u.id = uv.user_id
	//								LEFT JOIN rebate_config rc
	//							    ON u.rebate_id = rc.id
	//								LEFT JOIN
	//								(SELECT COUNT(a.login) login_count, IFNULL(GROUP_CONCAT(a.login),'') login, u.id FROM user u
	//								LEFT JOIN account a
	//								ON u.id = a.user_id
	//								WHERE u.parent_id = %v AND u.sales_id = %v
	//								GROUP BY user_id) a
	//                                ON u.id = a.id
	//                                LEFT JOIN
	//                                (SELECT p.user_id, SUM(IF(p.amount > 0,p.amount+p.pay_fee,0)) walletIn, SUM(IF(p.amount < 0 ,p.amount+p.pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 AND FIND_IN_SET(p.user_id,(SELECT GROUP_CONCAT(u.id) FROM user u WHERE u.parent_id = %v)) GROUP BY p.user_id) p1
	//                                ON u.id = p1.user_id
	//								WHERE u.parent_id = %v AND u.sales_id = %v %v %v`, parentId, salesId, parentId, parentId, salesId, content, where))

	//total := len(res)

	//if count != nil {
	//	total = abc.ToInt(abc.PtoString(count, "count"))
	//}

	for _, v := range res {
		if abc.ToInt(abc.PtoString(v, "status")) == -1 || abc.ToInt(abc.PtoString(v, "status")) == 0 {
			v.(map[string]interface{})["user_status"] = 4
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 && abc.ToFloat64(abc.PtoString(v, "walletIn")) != 0 {
			v.(map[string]interface{})["user_status"] = 2
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 1 {
			v.(map[string]interface{})["user_status"] = 3
			continue
		}

		if abc.ToInt(abc.PtoString(v, "auth_status")) == 0 || abc.ToInt(abc.PtoString(v, "auth_status")) == -1 {
			v.(map[string]interface{})["user_status"] = 1
			continue
		}
	}

	result, _ := abc.SqlOperators(fmt.Sprintf(`SELECT * FROM (SELECT comment,user_id FROM user_audit_log WHERE FIND_IN_SET(user_id,(SELECT GROUP_CONCAT(u.id) ids FROM user u WHERE u.parent_id = %v AND u.sales_id = %v)) and old = 0 ORDER BY create_time DESC) t GROUP BY t.user_id`), parentId, salesId)

	for _, v := range res {
		v.(map[string]interface{})["reason"] = ""
		for _, vv := range result {
			if abc.PtoString(v, "id") == abc.PtoString(vv, "user_id") {
				v.(map[string]interface{})["reason"] = golbal.Wrong[language][abc.ToInt(abc.PtoString(vv, "comment"))]
			}
		}
	}

	where1 := ""

	if start != "" || end != "" {
		where1 += fmt.Sprintf(" AND auth_status = 1")
	}

	if start != "" {
		where1 += fmt.Sprintf(" AND create_time >= '%v'", start+" 00:00:00")
	}

	if end != "" {
		where1 += fmt.Sprintf(" AND create_time <= '%v'", end+" 23:59:59")
	}

	re1, _ := abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, sales_id id FROM user WHERE LEFT(user_type,1) = 'L' AND FIND_IN_SET(sales_id,(SELECT GROUP_CONCAT(u.id) ids FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales')) AND LENGTH(path) -LENGTH(REPLACE(path,',','')) - 1 = %v AND auth_status = 1 %v GROUP BY sales_id`, parentId, salesId, len(arrLength), where1))
	re2, _ := abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, sales_id id FROM user WHERE LEFT(user_type,1) = 'u' AND FIND_IN_SET(sales_id,(SELECT GROUP_CONCAT(u.id) ids FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND u.user_type = 'sales')) AND LENGTH(path) -LENGTH(REPLACE(path,',','')) - 1 = %v AND auth_status = 1 %v GROUP BY sales_id`, parentId, salesId, len(arrLength), where1))

	re3, _ := abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id id FROM user WHERE LEFT(user_type,1) = 'L' AND FIND_IN_SET(parent_id,(SELECT GROUP_CONCAT(u.id) ids FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND left(u.user_type,1)= 'L')) AND auth_status = 1 %v GROUP BY parent_id`, parentId, salesId, where1))
	re4, _ := abc.SqlOperators(fmt.Sprintf(`SELECT COUNT(id) count, parent_id id FROM user WHERE LEFT(user_type,1) = 'u' AND FIND_IN_SET(parent_id,(SELECT GROUP_CONCAT(u.id) ids FROM user u WHERE u.parent_id = %v AND u.sales_id = %v AND left(u.user_type,1)= 'L')) AND auth_status = 1 %v GROUP BY parent_id`, parentId, salesId, where1))

	type SalesCount struct {
		LevelCount string `json:"level_count"`
		UserCount  string `json:"user_count"`
	}

	m := make(map[string]SalesCount)

	for _, v := range re1 {
		var s SalesCount
		s.LevelCount = abc.PtoString(v, "count")
		m[abc.PtoString(v, "id")] = s
	}

	for _, v := range re2 {
		a := m[abc.PtoString(v, "id")]
		a.UserCount = abc.PtoString(v, "count")

		m[abc.PtoString(v, "id")] = a
	}

	for _, v := range re3 {
		a := m[abc.PtoString(v, "id")]
		a.LevelCount = abc.PtoString(v, "count")
		m[abc.PtoString(v, "id")] = a
	}

	for _, v := range re4 {
		a := m[abc.PtoString(v, "id")]
		a.UserCount = abc.PtoString(v, "count")
		m[abc.PtoString(v, "id")] = a
	}

	for _, v := range res {
		v.(map[string]interface{})["level_count"] = 0
		v.(map[string]interface{})["user_count"] = 0
		if m != nil {
			for kk, vv := range m {
				if abc.PtoString(v, "id") == kk {
					v.(map[string]interface{})["level_count"] = abc.ToInt(vv.LevelCount)
					v.(map[string]interface{})["user_count"] = abc.ToInt(vv.UserCount)
					break
				}
			}
		}
	}

	if userId != 0 {
		for _, v := range res {
			arr := strings.Split(abc.PtoString(v, "email"), "@")
			if len(arr) == 2 {
				if len(arr[0]) <= 4 {
					v.(map[string]interface{})["email"] = "****" + "@" + arr[1]
				} else {
					v.(map[string]interface{})["email"] = "****" + arr[0][4:len(arr[0])] + "@" + arr[1]
				}
			}

			arr1 := abc.PtoString(v, "mobile")

			if len(arr1) <= 4 {
				v.(map[string]interface{})["mobile"] = "****"
			} else {
				v.(map[string]interface{})["mobile"] = "****" + arr1[4:len(arr1)]
			}
		}
	}

	r.Data = res
	r.Status = 1

	c.JSON(200, r.Response())
}

func GraphicValidationForGuest(c *gin.Context) {
	r := &R{}
	language := abc.ToString(c.MustGet("language"))
	phone := c.PostForm("phone")
	areaCode := c.PostForm("area_code")
	code := abc.ToString(abc.RandonNumber(5))
	sceneId := c.PostForm("sceneId")
	captchaVerifyParam := c.PostForm("captchaVerifyParam")

	r.Status, r.Msg, r.Data = graphicValidation.GraphicValidation(sceneId, captchaVerifyParam, language, phone, areaCode, code)

	c.JSON(200, r.Response())
}

func GraphicValidation(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	code := abc.ToString(abc.RandonNumber(5))
	sceneId := c.PostForm("sceneId")
	captchaVerifyParam := c.PostForm("captchaVerifyParam")

	if sceneId == "" || captchaVerifyParam == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	u := abc.GetUserById(uid)

	r.Status, r.Msg, r.Data = graphicValidation.GraphicValidation(sceneId, captchaVerifyParam, language, u.Mobile, u.Phonectcode, code)

	c.JSON(200, r.Response())
}

//执行一次
func CommissionCache(c *gin.Context) {
	//t2 := time.Date(time.Now().Year()-2, time.Now().Month()-1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02") + " 00:00:00"
	//t1 := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1).Format("2006-01-02") + " 23:59:59"
	t2 := "2023-01-01 00:00:00"
	t1 := "2024-05-31 23:59:59"

	month_res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, DATE_FORMAT(close_time,'%Y-%m') t FROM commission WHERE close_time > ? AND close_time < ? GROUP BY ib_id, DATE_FORMAT(close_time,'%Y-%m')`, t2, t1)

	month_res1, _ := abc.SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m') t, com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
                                           WHERE com1.commission_type = 0 and com1.close_time between ? and ? GROUP BY com1.ib_id, com1.symbol_type, DATE_FORMAT(com1.close_time,'%Y-%m')) com GROUP BY com.ib_id, DATE_FORMAT(com.close_time,'%Y-%m')`, t2, t1)

	ibIdRes, _ := abc.SqlOperators(`SELECT ib_id FROM commission GROUP BY ib_id`)

	var result []interface{}
	for _, v := range ibIdRes {
		res2, _ := abc.SqlOperators(`SELECT temp.ib_id, count(temp.id) quantity, SUM(temp.volume) volume, DATE_FORMAT(temp.close_time,'%Y-%m') t FROM (SELECT * FROM commission WHERE close_time > ? AND close_time < ? AND ib_id = ? GROUP BY ticket) temp GROUP BY DATE_FORMAT(temp.close_time,'%Y-%m')`, t2, t1, abc.ToInt(abc.PtoString(v, "ib_id")))

		result = append(result, res2...)
	}

	var arr []int
	for _, v := range month_res {
		flag := true
		var cache abc.Cache
		cache.Uid = abc.ToInt(abc.PtoString(v, "ib_id"))
		cache.Commission = abc.ToFloat64(abc.PtoString(v, "c1"))
		cache.CommissionDifference = abc.ToFloat64(abc.PtoString(v, "c2"))
		cache.Fee = abc.ToFloat64(abc.PtoString(v, "c3"))
		cache.Time = abc.PtoString(v, "t")
		cache.Type = 2
		for _, vv := range month_res1 {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vv, "t") {
				cache.Forex = abc.ToFloat64(abc.PtoString(vv, "forex"))
				cache.Dma = abc.ToFloat64(abc.PtoString(vv, "dma"))
				cache.Silver = abc.ToFloat64(abc.PtoString(vv, "silver"))
				cache.Metal = abc.ToFloat64(abc.PtoString(vv, "metal"))
				cache.StockCommission = abc.ToFloat64(abc.PtoString(vv, "stockCommission"))
				break
			}
		}

		for _, vvv := range result {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vvv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vvv, "t") {
				cache.Volume = abc.ToFloat64(abc.PtoString(vvv, "volume"))
				cache.Quantity = abc.ToInt(abc.PtoString(vvv, "quantity"))
				break
			}
		}

		abc.CreateCache(cache)

		if !abc.IsValueExist(arr, cache.Uid) {
			arr = append(arr, cache.Uid)
			flag = false
		}

		if !flag {
			abc.CreateSalesMonthCache(cache, t1, t2)
		}
	}

	c.JSON(200, gin.H{
		"status": 1,
		"msg":    "",
	})
}

//执行一次
func CommissionDayCache(c *gin.Context) {
	//t := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	t := "2024-06-01"
	start := t + " 00:00:00"
	end := "2024-06-30" + " 23:59:59"
	//res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, SUM(volume) volume, DATE_FORMAT(close_time,'%Y-%m-%d') t FROM commission WHERE close_time > ? AND close_time < ? GROUP BY ib_id`, start, end)
	//res1, _ := abc.SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m-%d') t, com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
	//                                      WHERE com1.commission_type = 0 and com1.close_time between ? and ? GROUP BY com1.ib_id, com1.symbol_type) com GROUP BY com.ib_id`, start, end)

	res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, DATE_FORMAT(close_time,'%Y-%m-%d') t FROM commission WHERE close_time >= ? AND close_time <= ? GROUP BY ib_id, DATE_FORMAT(close_time,'%Y-%m-%d')`, start, end)
	res1, _ := abc.SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m-%d') t, com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
	                                      WHERE com1.commission_type = 0 and com1.close_time between ? and ? GROUP BY com1.ib_id, DATE_FORMAT(com1.close_time,'%Y-%m-%d'), com1.symbol_type) com GROUP BY com.ib_id, DATE_FORMAT(com.close_time,'%Y-%m-%d')`, start, end)
	ibIdRes, _ := abc.SqlOperators(`SELECT ib_id FROM commission GROUP BY ib_id`)

	var result []interface{}
	for _, v := range ibIdRes {
		res2, _ := abc.SqlOperators(`SELECT temp.ib_id, COUNT(temp.id) quantity, SUM(temp.volume) volume, DATE_FORMAT(temp.close_time,'%Y-%m-%d') t FROM (SELECT * FROM commission WHERE close_time >= ? AND close_time <= ? AND ib_id = ? GROUP BY ticket) temp GROUP BY DATE_FORMAT(temp.close_time,'%Y-%m-%d')`, start, end, abc.ToInt(abc.PtoString(v, "ib_id")))

		result = append(result, res2...)
	}

	var arr []int
	for _, v := range res {
		flag := true
		var cache abc.Cache
		cache.Uid = abc.ToInt(abc.PtoString(v, "ib_id"))
		cache.Commission = abc.ToFloat64(abc.PtoString(v, "c1"))
		cache.CommissionDifference = abc.ToFloat64(abc.PtoString(v, "c2"))
		cache.Fee = abc.ToFloat64(abc.PtoString(v, "c3"))
		cache.Time = abc.PtoString(v, "t")
		cache.Type = 3
		for _, vv := range res1 {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vv, "t") {
				cache.Forex = abc.ToFloat64(abc.PtoString(vv, "forex"))
				cache.Dma = abc.ToFloat64(abc.PtoString(vv, "dma"))
				cache.Silver = abc.ToFloat64(abc.PtoString(vv, "silver"))
				cache.Metal = abc.ToFloat64(abc.PtoString(vv, "metal"))
				cache.StockCommission = abc.ToFloat64(abc.PtoString(vv, "stockCommission"))
				break
			}
		}
		for _, vvv := range result {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vvv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vvv, "t") {
				cache.Quantity = abc.ToInt(abc.PtoString(vvv, "quantity"))
				cache.Volume = abc.ToFloat64(abc.PtoString(vvv, "volume"))
				break
			}
		}

		abc.CreateCache(cache)

		if !abc.IsValueExist(arr, cache.Uid) {
			arr = append(arr, cache.Uid)
			flag = false
		}

		if !flag {
			abc.CreateSalesDayCache(cache, start, end)
		}
	}

	c.JSON(200, gin.H{
		"status": 1,
		"msg":    "",
	})
}

//每天执行
func CommissionEveryDayCache(c *gin.Context) {
	t := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	//t := "2023-01-31"
	start := t + " 00:00:00"
	end := t + " 23:59:59"
	res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, DATE_FORMAT(close_time,'%Y-%m-%d') t FROM commission WHERE close_time >= ? AND close_time <= ? GROUP BY ib_id`, start, end)
	res1, _ := abc.SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m-%d') t, com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
	                                     WHERE com1.commission_type = 0 and com1.close_time between ? and ? GROUP BY com1.ib_id, com1.symbol_type) com GROUP BY com.ib_id`, start, end)
	res2, _ := abc.SqlOperators(`SELECT temp.ib_id, SUM(temp.volume) volume, count(temp.id) quantity, DATE_FORMAT(temp.close_time,'%Y-%m-%d') t FROM (SELECT * FROM commission WHERE close_time >= ? AND close_time <= ? GROUP BY ticket, ib_id) temp GROUP BY temp.ib_id`, start, end)
	//res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, SUM(volume) volume, DATE_FORMAT(close_time,'%Y-%m-%d') t FROM commission WHERE close_time > ? AND close_time < ? GROUP BY ib_id, DATE_FORMAT(close_time,'%Y-%m-%d')`, start, end)
	//res1, _ := abc.SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m-%d') t, com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
	//                                      WHERE com1.commission_type = 0 and com1.close_time between ? and ? GROUP BY com1.ib_id, DATE_FORMAT(com1.close_time,'%Y-%m-%d'), com1.symbol_type) com GROUP BY com.ib_id, DATE_FORMAT(com.close_time,'%Y-%m-%d')`, start, end)
	for _, v := range res {
		var cache abc.Cache
		cache.Uid = abc.ToInt(abc.PtoString(v, "ib_id"))
		cache.Commission = abc.ToFloat64(abc.PtoString(v, "c1"))
		cache.CommissionDifference = abc.ToFloat64(abc.PtoString(v, "c2"))
		cache.Fee = abc.ToFloat64(abc.PtoString(v, "c3"))
		cache.Time = t
		cache.Type = 3
		for _, vv := range res1 {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vv, "t") {
				cache.Forex = abc.ToFloat64(abc.PtoString(vv, "forex"))
				cache.Dma = abc.ToFloat64(abc.PtoString(vv, "dma"))
				cache.Silver = abc.ToFloat64(abc.PtoString(vv, "silver"))
				cache.Metal = abc.ToFloat64(abc.PtoString(vv, "metal"))
				cache.StockCommission = abc.ToFloat64(abc.PtoString(vv, "stockCommission"))
				break
			}
		}
		for _, vvv := range res2 {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vvv, "ib_id")) && abc.PtoString(v, "t") == abc.PtoString(vvv, "t") {
				cache.Quantity = abc.ToInt(abc.PtoString(vvv, "quantity"))
				cache.Volume = abc.ToFloat64(abc.PtoString(vvv, "volume"))
				break
			}
		}
		abc.CreateCache(cache)
		abc.CreateSalesDayCache(cache, start, end)
	}

	ibIdRes, _ := abc.SqlOperators(`SELECT ib_id FROM commission GROUP BY ib_id`)

	for _, v := range ibIdRes {
		if v != nil {
			if time.Now().Day() == 1 || time.Now().Day() == 2 || time.Now().Day() == 3 {
				abc.GenerateLastMonthData(abc.ToInt(abc.PtoString(v, "ib_id")))
			}
		}
	}
	c.JSON(200, gin.H{
		"status": 1,
		"msg":    "",
	})
}

//执行一次
func CommissionTotalCache(c *gin.Context) {
	total_res, _ := abc.SqlOperators(`SELECT ib_id, SUM(IF(commission_type = 0,fee, 0)) c1, SUM(IF(commission_type = 1,fee,0)) c2, SUM(IF(commission_type = 2,fee,0)) c3, SUM(volume) volume FROM commission WHERE close_time <= ? GROUP BY ib_id`, "2022-12-31 23:59:59")

	res1, _ := abc.SqlOperators(`SELECT com.ib_id, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.ib_id,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type, com1.close_time FROM commission com1
                                           WHERE com1.commission_type = 0 and com1.close_time <= ? GROUP BY com1.ib_id, com1.symbol_type) com GROUP BY com.ib_id`, "2022-12-31 23:59:59")
	ibIdRes, _ := abc.SqlOperators(`SELECT ib_id FROM commission GROUP BY ib_id`)

	var result []interface{}
	for _, v := range ibIdRes {
		res2, _ := abc.SqlOperators(`SELECT temp.ib_id, COUNT(temp.id) quantity, SUM(temp.volume) volume FROM (SELECT * FROM commission WHERE close_time <= ? AND ib_id = ? GROUP BY ticket) temp`, "2022-12-31 23:59:59", abc.ToInt(abc.PtoString(v, "ib_id")))

		result = append(result, res2...)
	}

	var arr []int
	for _, v := range total_res {
		flag := true
		var cache abc.Cache
		cache.Uid = abc.ToInt(abc.PtoString(v, "ib_id"))
		cache.Commission = abc.ToFloat64(abc.PtoString(v, "c1"))
		cache.CommissionDifference = abc.ToFloat64(abc.PtoString(v, "c2"))
		cache.Fee = abc.ToFloat64(abc.PtoString(v, "c3"))
		cache.Type = 1
		for _, vv := range res1 {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vv, "ib_id")) {
				cache.Forex = abc.ToFloat64(abc.PtoString(vv, "forex"))
				cache.Dma = abc.ToFloat64(abc.PtoString(vv, "dma"))
				cache.Silver = abc.ToFloat64(abc.PtoString(vv, "silver"))
				cache.Metal = abc.ToFloat64(abc.PtoString(vv, "metal"))
				cache.StockCommission = abc.ToFloat64(abc.PtoString(vv, "stockCommission"))
				break
			}
		}
		for _, vvv := range result {
			if abc.ToInt(abc.PtoString(v, "ib_id")) == abc.ToInt(abc.PtoString(vvv, "ib_id")) {
				cache.Quantity = abc.ToInt(abc.PtoString(vvv, "quantity"))
				cache.Volume = abc.ToFloat64(abc.PtoString(vvv, "volume"))
				break
			}
		}

		abc.CreateCache(cache)

		if !abc.IsValueExist(arr, cache.Uid) {
			arr = append(arr, cache.Uid)
			flag = false
		}

		if !flag {
			abc.CreateSalesTotalCache(cache, "2022-12-31 23:59:59")
		}
	}

	c.JSON(200, gin.H{
		"status": 1,
		"msg":    "",
	})
}

func ExportCommission(c *gin.Context) {
	r := &R{}
	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	t := c.PostForm("t")

	if t == "" {
		r.Status = 0
		r.Msg = golbal.Wrong[language][10000]

		c.JSON(200, r.Response())

		return
	}

	startTime := t + " 00:00:00"
	endTime := t + " 23:59:59"

	r.Status = 1
	r.Data = abc.CreateCommissionCache(uid, startTime, endTime, language)

	c.JSON(200, r.Response())
}
