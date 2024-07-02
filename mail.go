package abc

import (
	"time"
)

func MailContent(id int) MailTpl {
	var m MailTpl
	db.Debug().Where("id = ?", id).First(&m)

	return m
}

func VerifyMailCode(username, mailCode string) Captcha {
	var c Captcha
	db.Debug().Where("type = 0 and address = ? and code = ? and create_at > ? and used = 0 and type = 0", username, mailCode, time.Now().Unix()-600).First(&c)

	return c
}
