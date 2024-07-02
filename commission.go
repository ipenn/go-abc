package abc

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

func GetCommissionState(uid int) int {
	var com CommissionSetCustom
	db.Debug().Where("user_id = ?", uid).Order("create_time DESC").First(&com)

	return com.Status
}

func TxGetCommissionSetCustom(tx *gorm.DB, where any) (csc []CommissionSetCustom) {
	tx.Debug().Where(where).Find(&csc)
	return
}

func TxUpdateUserCommission(tx *gorm.DB, commission Commission) error {
	if err := tx.Debug().Create(&commission).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func CreateCommission(sql string) error {
	sql = "insert into commission_set_custom (user_id,symbol,amount,type,create_time,status,user_path) values " + sql[1:]
	_, err := db.Debug().Raw(sql).Rows()

	return err
}

func GetMyCommissionAuthority(uid, rebateCate, cType int) map[string]float64 {
	var temp []interface{}
	m := make(map[string]float64)
	res, _ := SqlOperators(`SELECT type, MAX(am ount) amount FROM commission_set_custom WHERE user_id = ? AND status = -1 AND type >= 0 GROUP BY type`, uid)
	temp = res

	if len(res) == 0 {
		res1, _ := SqlOperators(`SELECT type, MAX(amount) amount FROM commission_set_custom WHERE user_id = ? AND status = 1 AND type >= 0 GROUP BY type`, uid)
		temp = res1

		if len(res1) == 0 {
			res2, _ := SqlOperators(`SELECT type, MAX(amount) amount FROM commission_set WHERE cate = ? AND type >= 0 AND type <= ? GROUP BY type`, rebateCate, cType)
			temp = res2
		}
	}

	for i := 0; i <= cType; i++ {
		flag := true
		for _, v := range temp {
			if i == ToInt(PtoString(v, "type")) {
				m[PtoString(v, "type")] = ToFloat64(PtoString(v, "amount"))
				flag = false
				break
			}
		}

		if flag {
			m[ToString(i)] = 0
		}
	}

	return m
}

func GetCommissionType(uid int) int {
	var i []InviteCode
	db.Debug().Where("user_id = ? AND FIND_IN_SET(name,'CFH-STD-02P,CFH-STD-102P,CFH-STD-03P')", uid).Find(&i)

	if len(i) != 0 {
		return 14
	}

	return 8
}

func GetMyCommissionTemplate(uid, rebateCate, cType int) []interface{} {
	var temp []interface{}

	res, _ := SqlOperators(`SELECT type,  amount, symbol FROM commission_set_custom WHERE user_id = ? AND status = 1 AND type >= 0 `, uid)
	temp = res

	if len(res) == 0 {
		res1, _ := SqlOperators(`SELECT type,  amount, symbol FROM commission_set WHERE cate = ? AND type >= 0 AND type <= ?`, rebateCate, cType)
		temp = res1
	}

	return temp
}

// 获取已通过审核的佣金
func GetMyCommission(uid, rebateCate int) map[string]interface{} {
	var temp []interface{}
	m := make(map[string]interface{})

	res, _ := SqlOperators(`SELECT type, MAX(amount) amount, symbol FROM commission_set_custom WHERE user_id = ? AND status = 1 AND type >= 0 GROUP BY type`, uid)
	temp = res

	if len(res) == 0 {
		res1, _ := SqlOperators(`SELECT type, MAX(amount) amount, symbol FROM commission_set WHERE cate = ? AND type >= 0 GROUP BY type`, rebateCate)
		temp = res1
	}

	for _, v := range temp {
		m[PtoString(v, "type")] = v
	}

	return m
}

func CheckCommissionReview(uid int) CommissionSetCustom {
	var c CommissionSetCustom
	db.Debug().Where("user_id = ?", uid).First(&c)

	return c
}

func GetCommissionSetAll() (commissionSet []CommissionSet) {
	db.Debug().Find(&commissionSet)
	return
}

func CreateCommissionRecord(tx *gorm.DB, uid int, path string) error {
	if err := db.Debug().Raw(`insert into commission_set_custom (user_id,symbol,amount,type,create_time,status,user_path) select ?,symbol,0,type,NOW(),1, ? from commission_set where cate = 1 and type > -1`, uid, path).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteCommissionById(tx *gorm.DB, uid int) error {
	if err := tx.Debug().Where("user_id = ?", uid).Delete(&CommissionSetCustom{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetCommissionSetCustomOne(uid int) CommissionSetCustom {
	var c CommissionSetCustom
	db.Debug().Where("user_id = ?", uid).First(&c)

	return c
}

func GetCommissionSetCustomByUid(uid int, where string) (csc []CommissionSetCustom) {
	db.Debug().Where("user_id=?", uid).Where(where).Group("type").Find(&csc)
	return csc
}

func GetCommissionSet(cate int) (cs []CommissionSet) {
	db.Debug().Where("cate=?", cate).Group("type").Find(&cs)
	return cs
}

func GetCommissionList(page, size int, where string) (int, int64, []Commission) {
	var com []Commission
	db.Debug().Where(where).Limit(size).Order("close_time desc").Offset((page - 1) * size).Find(&com)
	var count int64
	db.Debug().Model(Commission{}).Where(where).Count(&count)
	return 1, count, com
}

func GetWageData(where string) []Wage {
	var wage []Wage
	db.Debug().Where(where).Order("create_time desc").Find(&wage)
	return wage
}

func GetCommissionSum(where string) AmountInfo {
	var data AmountInfo
	db.Debug().Model(Commission{}).Select("sum(volume) as volume,sum(fee) as fee").Where(where).Scan(&data)
	return data
}

func GetCommissionNum(uid int) bool {
	var c CommissionSetCustom
	db.Debug().Where("user_id = ?", uid).Order("create_time DESC").First(&c)

	if c.Id != 0 {
		if ToInt(c.CreateTime[5:7]) == ToInt(time.Now().Format("01")) {
			return false
		}
	}

	return true
}

func GetCommissionPassed(uid int) CommissionSetCustom {
	var c CommissionSetCustom
	db.Debug().Where("user_id = ? and status = 1", uid).First(&c)

	return c
}

func StatisticalIncome(uid int, t string) (float64, float64) {
	res, _ := SqlOperator(`SELECT SUM(fee) fee, SUM(volume) volume FROM commission WHERE ib_id = ? AND close_time >= ? `, uid, t)

	res1, _ := SqlOperator(`SELECT SUM(temp.volume) volume FROM (SELECT volume FROM commission WHERE ib_id = ? AND close_time >= ? GROUP BY ticket) temp`, uid, t)

	fee := 0.0
	volume := 0.0

	if res != nil {
		fee = ToFloat64(PtoString(res, "fee"))
		volume = ToFloat64(PtoString(res1, "volume"))
	}

	return fee, volume
}

func RevenueList(where string, page, size int) ([]interface{}, int64, interface{}, interface{}) {
	res, _ := SqlOperators(fmt.Sprintf(`SELECT c.id, c.ticket, c.login, c.symbol, c.volume, c.fee, c.close_time, c.commission_type, u1.true_name trade_name, u2.true_name ib_name FROM commission c
					  LEFT JOIN user u1
					  ON c.uid = u1.id
					  LEFT JOIN user u2
					  ON c.ib_id = u2.id
					  WHERE %v ORDER BY close_time DESC LIMIT %v,%v`, where, (page-1)*size, size))
	res1, _ := SqlOperator(fmt.Sprintf(`SELECT count(c.id) count FROM commission c
					  LEFT JOIN user u1
					  ON c.uid = u1.id
					  LEFT JOIN user u2
					  ON c.ib_id = u2.id
					  WHERE %v `, where))
	res2, _ := SqlOperator(fmt.Sprintf(`SELECT sum(c.fee) fee FROM commission c
					  LEFT JOIN user u1
					  ON c.uid = u1.id
					  LEFT JOIN user u2
					  ON c.ib_id = u2.id
					  WHERE %v`, where))
	res3, _ := SqlOperator(fmt.Sprintf(`SELECT SUM(temp.volume) volume FROM (SELECT c.volume FROM commission c
					  LEFT JOIN user u1
					  ON c.uid = u1.id
					  LEFT JOIN user u2
					  ON c.ib_id = u2.id
					  WHERE %v GROUP BY c.ticket) temp`, where))

	return res, ToInt64(PtoString(res1, "count")), res2, res3
}

func GetSalesCommission(where string) (data []SalesCommission) {
	db.Debug().Where(where).Order("create_time desc").Find(&data)
	return
}
