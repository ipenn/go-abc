package controllers

import (
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	task_guest "github.com/chenqgp/abc/task/task-guest"
	nx "github.com/chenqgp/abc/third/sms/smsNx"
	ucloud "github.com/chenqgp/abc/third/sms/smsUcloud"
	uni "github.com/chenqgp/abc/third/sms/smsUni"
	"github.com/chenqgp/abc/third/sms/smsYimei"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"strings"
	"time"
)

func SmsCode(c *gin.Context) {
	r := &R{}

	uid := abc.ToInt(c.MustGet("uid"))
	language := abc.ToString(c.MustGet("language"))
	phone := c.PostForm("phone")
	areaCode := c.PostForm("area_code")
	code := abc.ToString(abc.RandonNumber(5))

	u := abc.GetUserById(uid)

	if u.Mobile != "" {
		phone = u.Mobile
	}

	if u.Phonectcode != "" {
		areaCode = u.Phonectcode
	}

	newPhone := phone
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
			phone = "(" + strings.ReplaceAll(areaCode, "+", "") +")" + phone
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
