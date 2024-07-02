package abc

import "time"

var UVC_Max_deduction_amount_sql = `SELECT MAX(deduction_amount) amount FROM user_vip_cash WHERE (pay_amount <= ? AND pay_amount > 0)`

func CheckCouponExist(uid, id int) UserVipCash {
	var uc UserVipCash
	db.Debug().Where("user_id = ? and id = ? and status = 0 and create_time > ?", uid, id, time.Now().AddDate(0, 0, -90).Format("2006-01-02 15:04:05")).First(&uc)

	return uc
}

func GetUserVipCash(where string) (userVipCash []UserVipCash) {
	db.Debug().Where(where).Order("create_time desc").Find(&userVipCash)
	return userVipCash
}

func GetUserVipCashOne(where string) (userVipCash UserVipCash) {
	db.Debug().Where(where).First(&userVipCash)
	return userVipCash
}
