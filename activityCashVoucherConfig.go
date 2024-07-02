package abc

import (
	"fmt"
	golbal "github.com/chenqgp/abc/global"
	"time"
)

func GetActivityCashVoucherConfig(where string) (config []ActivityCashVoucherConfig) {
	db.Debug().Where(where).Find(&config)
	return config
}

func GetActivityCashVoucherConfigOne(where string) ActivityCashVoucherConfig {
	var config ActivityCashVoucherConfig
	db.Debug().Where(where).First(&config)
	return config
}

func GetActivityCashVoucherUsedList(page, size int, cashNo string) (int64, any) {
	var result []ActivityCashVoucherList
	db.Debug().Raw(fmt.Sprintf(`SELECT
	cv.id,
	cv.cash_no,
	cv.create_time,
	cv.end_time,
	cv.amount,
	cv.volume,
	cv.status,
	u.true_name,
	u.email
FROM
	cash_voucher cv LEFT JOIN user u on cv.user_id=u.id
WHERE
	cv.comment like '%%-%s-%%' limit %d,%d`, cashNo, (page-1)*size, size)).Scan(&result)
	var count int64
	db.Debug().Table("cash_voucher").Where(fmt.Sprintf("comment like '%%-%s-%%'", cashNo)).Count(&count)
	return count, result
}

func UseActivityCashVoucher(uid, toUid, t int, amount float64, language string) (int, string) {
	times := FormatNow()
	act := GetActivityCashVoucherConfigOne(fmt.Sprintf(`user_id=%d and is_del = 0 and type=%d
		and start_time <= '%s' and end_time > '%s'`, uid, t, times[:10], times[:10]))
	if act.Id == 0 {
		return 1, ""
	}
	toUser := GetUserById(toUid)
	if toUser.Id == 0 {
		return 0, golbal.Wrong[language][10043]
	}
	if uid == toUid {
		return 0, golbal.Wrong[language][10117]
	}
	volume := 0.00
	if act.Type == 0 {
		amount = act.InitAmount
		volume = act.Volume
	} else if act.Type == 1 {
		if amount > act.Amount {
			return 0, golbal.Wrong[language][10118]
		}

		volume = act.Volume * amount / act.InitAmount
		UpdateSql("activity_cash_voucher_config", fmt.Sprintf("id=%d", act.Id), map[string]interface{}{
			"amount": act.Amount - amount,
		})
	}
	cash := CashVoucher{
		Amount:     amount,
		Volume:     volume,
		UserId:     toUid,
		CreateTime: times,
		EndTime:    ToAddDay(act.Days),
		CashNo:     fmt.Sprintf("%d_%d", time.Now().Unix(), toUid),
		Comment:    fmt.Sprintf("现金券特权活动-%s-%d/%d/F", act.No, uid, toUid),
	}
	if act.IsDeposit != "需存款" {
		cash.Status = 1
		creditDetail := CreditDetail{
			UserId:     toUid,
			Login:      0,
			CreateTime: times,
			OverTime:   ToAddDay(act.Days),
			Balance:    cash.Amount,
			Source:     3,
			Comment:    "现金券特权活动",
			Volume:     cash.Volume,
			CouponNo:   fmt.Sprintf("%s-%d/%d/F", act.No, uid, toUid),
		}
		if status := creditDetail.CreateCreditDetail(); status == 0 {
			return status, golbal.Wrong[language][10100]
		}
	}
	if status := cash.CreateCashVoucher(); status == 0 {
		return status, golbal.Wrong[language][10100]
	}
	return 1, ""
}
