package abc

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

var (
	InterestVip = map[int]float64{
		3: 0.1,
		4: 0.2,
		5: 0.3,
	}
)

func GetInterest(where string) (interest []Interest) {
	db.Debug().Where(where).Find(&interest)
	return interest
}

func (interest Interest) CreateInterestAndAddBalance() int {
	//todo 错误提示
	tx := db.Begin()
	if err := tx.Debug().Create(&interest).Error; err != nil {
		log.Println(" abc CreateInterest1 ", err)
		//telegram.SendMsg(telegram.TEXT, telegram.TEST,
		//	fmt.Sprintf("活动明细interest插入失败,UID:%d,Amount:%2f\n", interest.UserId, interest.Fee))
		tx.Rollback()
		return 0
	}
	if err := tx.Debug().Model(User{}).Where("id=?", interest.UserId).Updates(map[string]any{
		"wallet_balance": gorm.Expr(fmt.Sprintf("wallet_balance + (%.2f)", interest.Fee)),
	}).Error; err != nil {
		log.Println(" abc CreateInterest2 ", err)
		//telegram.SendMsg(telegram.TEXT, telegram.TEST,
		//	fmt.Sprintf("用户钱包余额更新失败,UID:%d,Amount:%2f\n", interest.UserId, interest.Fee))
		tx.Rollback()
		return 0
	}
	tx.Commit()
	return 1
}

func GetInterestData(uid int) (data InterestData) {
	//total
	db.Debug().Select("sum(fee) as total").Table("interest").
		Where("user_id=? and type=0", uid).Scan(&data)
	//yesterday
	yTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	db.Debug().Select("sum(fee) as yesterday").Table("interest").
		Where("user_id=? and type=0 and date_format(create_time,'%Y-%m-%d')=?", uid, yTime).Scan(&data)
	//yesterday extra rate
	db.Debug().Select("value as extra_rate").Table("extra_rate").
		Where("date_format(create_time,'%Y-%m-%d')=?", yTime).Scan(&data)
	return data
}

func GetInterestDateData(uid int, dateType string, dateNum int) (int, []DateData) {
	status, data := GetDateData(dateType, dateNum, "sum(fee) as value", "create_time",
		"interest", fmt.Sprintf("and user_id=%d and type=0", uid), true)
	if status == 0 {
		return status, nil
	}
	return status, data
}

func GetInterestList(page, size int, where string) (int, int64, []Interest) {

	var interest []Interest
	db.Debug().Where(where).Limit(size).Offset((page - 1) * size).Order("create_time desc").Find(&interest)
	var count int64
	db.Debug().Model(Interest{}).Where(where).Count(&count)
	return 1, count, interest
}

func InterestSum(where string) (data AmountInfo) {
	db.Debug().Model(Interest{}).Select("sum(fee) as amount").Where(where).Scan(&data)
	return data
}
