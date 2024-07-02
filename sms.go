package abc

import "time"

func GetSmsChannel() Config {
	var c Config
	db.Debug().Where("name = ? and value = ?", "SMS_GLOBAL", "ok").First(&c)
	return c
}

func VerifySmsCode(phone, code string) Captcha {
	var c Captcha
	db.Debug().Where("address = ? AND `code` = ? AND type = 1 AND used = 0 and create_at > ?", phone, code, time.Now().Unix()-600).First(&c)

	return c
}

func DeleteSmsOrMail(id int) {
	db.Debug().Where("id = ?", id).Delete(Captcha{})
}

func PhoneIsDisabled(phone string) int {
	var count int64
	db.Debug().Model(&Captcha{}).Where("address = ? and used = 0 and type = 1 and create_time >= ? and create_time <= ?", phone, time.Now().Add(-30*time.Minute).Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")).Count(&count)

	var count1 int64
	db.Debug().Model(&Captcha{}).Where("address = ? and used = 0 and type = 1 and create_time >= ? and create_time <= ?", phone, time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")).Count(&count1)

	if count >= 5 && count1 < 20 {
		var ibh InfoBlackHouse
		ibh.Address = phone
		ibh.Status = 1
		ibh.CreateTime = time.Now().Format("2006-01-02 15:04:05")
		db.Debug().Create(&ibh)
		return 2
	}

	if count1 >= 20 {
		var ibh InfoBlackHouse
		ibh.Address = phone
		ibh.Status = 2
		ibh.CreateTime = time.Now().Format("2006-01-02 15:04:05")
		db.Debug().Create(&ibh)

		return 3
	}
	return 1
}

func CheckPhoneStatus(phone string) int {
	var ibh InfoBlackHouse
	db.Debug().Where("address = ?", phone).First(&ibh)

	if ibh.Id != 0 {
		if ibh.Status == 2 {
			return 3
		}

		if ibh.Status == 1 {
			t, _ := time.Parse("2006-01-02 15:04:05", ibh.CreateTime)
			if time.Now().Format("2006-01-02 15:04:05") >= t.Add(30*time.Minute).Format("2006-01-02 15:04:05") {
				db.Debug().Where("address = ?", phone).Delete(&InfoBlackHouse{})
				return 1
			} else {
				return 2
			}
		}
	}

	return 1
}
