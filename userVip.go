package abc

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func GetUserVipById(uid int) UserVip {
	var uv UserVip
	db.Debug().Where("user_id = ?", uid).First(&uv)

	return uv
}

func GetVipConfig() []UserVipConfig {
	var config []UserVipConfig
	db.Debug().Find(&config)
	return config
}

func GetUserVipOrCreate(uid int) (userVip UserVip) {
	where := fmt.Sprintf("user_id=%d", uid)
	userVip.UserId = uid
	userVip.Grade = 1
	db.Debug().Where(where).FirstOrCreate(&userVip)
	return userVip
}

func (userVip UserVip) SaveUserVip(where string) int {
	if err := db.Debug().Model(userVip).Where(where).Updates(&userVip).Error; err != nil {
		log.Println(" abc SaveUserVip ", err)
		return 0
	}
	return 1
}

func GetUserVips(where string) (userVip []UserVip) {
	db.Debug().Where(where).Find(&userVip)
	return userVip
}

func GetUserGrade(t string) []UserGrade {
	var lists []UserGrade
	db.Debug().Raw(fmt.Sprintf(`SELECT
	ui.user_id,ui.birthday,uv.grade
FROM
	user_info ui
	LEFT JOIN user_vip uv on ui.user_id=uv.user_id
WHERE
	ui.birthday LIKE '%%%s%%' and uv.grade>2`, t[5:10])).Scan(&lists)
	return lists
}

func UserVipUpgrade(uid int) bool {
	config := GetVipConfig()
	total := ScoreCountDetail{}
	total.ScoreCountDetails(fmt.Sprintf("where user_id=%d", uid))
	month := ScoreCountDetail{}
	month.ScoreCountDetails(fmt.Sprintf("where user_id=%d and close_time>'%s'", uid, time.Now().Format("2006-01")))
	nowLevel, nowScore := 1, 0.00
	for _, c := range config {
		c.Flag = strings.Trim(c.Flag, "]")
		c.Flag = strings.Trim(c.Flag, "[")
		f := strings.Split(c.Flag, ",")
		t, m := f[0], f[1]
		if total.Total >= ToFloat64(t) {
			nowLevel = c.GradeId
			nowScore = total.Total

		} else if month.Total >= ToFloat64(m) {
			nowLevel = c.GradeId
			nowScore = month.Total
		}
	}
	vip := GetUserVipOrCreate(uid)
	if nowLevel > vip.Grade {
		t := FormatNow()
		flow := UserVipFlow{}
		flow.Grade = nowLevel
		flow.Score = nowScore
		flow.CreateTime = t
		flow.UserId = uid
		flow.PrevGrade = vip.Grade
		db.Debug().Create(&flow)
		var value []string
		for _, con := range config {
			if con.GradeId > vip.Grade && con.GradeId <= nowLevel {
				var key []kv
				json.Unmarshal([]byte(con.CashCoupon), &key)
				for i, k := range key {
					value = append(value, fmt.Sprintf(`(%d,'%s','%s',%d,%d,'%s')`,
						uid, fmt.Sprintf("%d%d%d", time.Now().Unix()+int64(i), uid, k.K),
						t, k.K, k.V, fmt.Sprintf("Upgrade to %d", con.GradeId)))
				}
			}
		}
		if len(value) != 0 {
			sql := fmt.Sprintf(`insert into user_vip_cash 
    ( user_id, order_no, create_time, pay_amount, deduction_amount, comment) values %s`, strings.Join(value, ","))
			if _, err := db.Debug().Raw(sql).Rows(); err != nil {
				log.Println(" abc UserVipUpgrade ", err)
				return false
			}
		}
		vip.Grade = nowLevel
		vip.Score = nowScore
		vip.UpdateTime, vip.UpTime = t, t
		vip.SaveUserVip(fmt.Sprintf("user_id=%d", vip.UserId))
		return true
	}
	return false
}
