package abc

import (
	"fmt"
	"log"
	"strings"
	"time"

	golbal "github.com/chenqgp/abc/global"
	"gorm.io/gorm"
)

func TxUpdateScore(tx *gorm.DB, score ScoreLog) error {
	if err := tx.Debug().Create(&score).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetScoreLog(where string) (scoreLog []ScoreLog) {
	db.Debug().Where(where).Order("create_time desc").Find(&scoreLog)
	return scoreLog
}

func (scoreLog ScoreLog) CreateScoreLog() int {
	if err := db.Debug().Create(&scoreLog).Error; err != nil {
		log.Println(" abc CreateScoreLog ", err)
		return 0
	}
	return 1
}

func (detail *ScoreCountDetail) ScoreCountDetails(where string) {
	db.Debug().Raw(fmt.Sprintf(`select sum(amount) as total,sum(use_amount) as used from score_log %s`, where)).
		Scan(&detail)
	detail.Using = detail.Total - detail.Used
}

func ExpireScore(uid int) float64 {
	detail := struct {
		Expire float64 `json:"expire"`
	}{}
	t := fmt.Sprintf("%s%s", time.Now().AddDate(0, -11, 0).Format("2006-01"), "-01")
	db.Debug().Raw(fmt.Sprintf(`select user_id, sum(amount - use_amount) expire from score_log 
		where user_id=%d and create_time < '%s' and use_amount < amount and amount > 0`, uid, t)).Scan(&detail)
	return detail.Expire

}

func ScoreDetails(uid int, t, startTime, endTime string, page, size int) (int, string, int64, any) {

	var scoreLogReturn []struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Amount     float64 `json:"amount"`
		Type       int     `json:"type"`
		UseAmount  float64 `json:"use_amount"`
	}
	where := fmt.Sprintf("user_id=%v", uid)
	if startTime != "" && endTime != "" {
		if len(startTime) <= 10 {
			startTime += " 00:00:00"
		}
		if len(endTime) <= 10 {
			endTime += " 23:59:59"
		}
		if StringToUnix(startTime) > 0 && StringToUnix(endTime) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", startTime, endTime)
		}
	}
	if t != "" {
		where += fmt.Sprintf(" and type=%d", ToInt(t))
	}
	db.Debug().Model(ScoreLog{}).Where(where).
		Limit(size).Offset((page - 1) * size).Order("create_time desc").Scan(&scoreLogReturn)
	var count int64
	db.Debug().Model(ScoreLog{}).Where(where).Count(&count)

	return 1, "", count, scoreLogReturn
}

func ScoreUsed(uid, page, size int, startTime, endTime string) (int, string, int64, any) {
	var scoreUsed []struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		PayAmount  float64 `json:"pay_amount"`
		Amount     float64 `json:"amount"`
		OrderNo    string  `json:"order_no"`
		IsGoods    int     `json:"is_goods"`
	}
	where := fmt.Sprintf("user_id=%v", uid)
	if startTime != "" && endTime != "" {
		if len(startTime) <= 10 {
			startTime += " 00:00:00"
		}
		if len(endTime) <= 10 {
			endTime += " 23:59:59"
		}
		if StringToUnix(startTime) > 0 && StringToUnix(endTime) > 0 {
			where += fmt.Sprintf(" and create_time >= '%s' and create_time<='%s'", startTime, endTime)
		}
	}
	db.Debug().Select(`sd.id,sd.user_id,sd.create_time,sd.pay_amount,sd.amount,sd.order_no,sc.is_goods`).
		Table("score_detail sd").Joins(`left join score_config sc on sd.goods_id=sc.id`).Where(where).
		Limit(size).Offset((page - 1) * size).Order("create_time desc").Scan(&scoreUsed)
	for i, s := range scoreUsed {
		if s.IsGoods == 0 {
			scoreUsed[i].OrderNo = fmt.Sprintf("$%.2f", s.Amount)
		}
	}

	var count int64
	db.Debug().Model(ScoreDetail{}).Where(where).Count(&count)

	return 1, "", count, scoreUsed
}

func GetScoreConfig(uid int) (int, string, any) {
	var scoreConfig []ScoreConfigReturn
	db.Debug().Model(ScoreConfig{}).Where("is_del=0").Order("weight").Scan(&scoreConfig)

	var useCount []struct {
		GoodsId int `json:"goods_id"`
		Count   int `json:"count"`
	}
	db.Debug().Raw(`SELECT goods_id,COUNT(goods_id) as count FROM score_detail
		WHERE user_id=? and create_time>? GROUP BY goods_id`, uid, FormatNow()[:7]).Scan(&useCount)

	for i, s := range scoreConfig {
		scoreConfig[i].Surplus = scoreConfig[i].BuyNum
		if len(useCount) > 0 {
			for ii, use := range useCount {
				if s.Id == use.GoodsId {
					scoreConfig[i].Surplus = scoreConfig[i].BuyNum - use.Count
					useCount = append(useCount[:ii], useCount[ii+1:]...)
					continue
				}
			}
		}
	}

	res := make(map[string][]ScoreConfigReturn, 0)
	var key []string
	for _, s := range scoreConfig {
		res[fmt.Sprintf("%d,%s", s.Weight, s.Name)] = append(res[fmt.Sprintf("%d,%s", s.Weight, s.Name)], s)
	}
	for k, _ := range res {
		key = append(key, k)
	}
	key = SortCouponList(key)
	var data [][]ScoreConfigReturn
	for _, s := range key {
		data = append(data, res[s])
	}

	return 1, "", data
}

// 排序
func SortCouponList(arr []string) []string {
	if len(arr) <= 1 {
		return arr
	}
	for i := 1; i < len(arr); i++ {
		for j := i - 1; j >= 0; j-- {
			if ToInt(strings.Split(arr[j], ",")[0]) > ToInt(strings.Split(arr[j+1], ",")[0]) {
				swaps(arr, j, j+1)
			}

		}
	}
	return arr
}

func swaps(arr []string, i, j int) []string {
	temp := arr[j]
	arr[j] = arr[i]
	arr[i] = temp
	return arr
}

func ExchangeGoods(id, uid, addressId int, language string) (int, string, string, float64, string, float64,int) {
	status, messages := 1, ""
	var scoreConfig ScoreConfig
	db.Debug().Where("id=? and is_del=0", id).First(&scoreConfig)
	if scoreConfig.Id == 0 {
		status, messages = 0, golbal.Wrong[language][10000]
		return status, messages, "", 0, "", 0,0
	}

	now := FormatNow()
	if (scoreConfig.StartTime != "" && scoreConfig.StartTime > now) || (scoreConfig.EndTime != "" && scoreConfig.EndTime < now) {
		status, messages = 0, golbal.Wrong[language][10103]
		return status, messages, "", 0, "", 0,0
	}

	var count int64
	month := time.Now().Format("2006-01")
	db.Debug().Model(ScoreDetail{}).Where("user_id = ? and goods_id=? and create_time>?", uid, scoreConfig.Id, month).Count(&count)
	if int(count) >= scoreConfig.BuyNum {
		status, messages = 0, golbal.Wrong[language][10104]
		return status, messages, "", 0, "", 0,0
	}

	var score ScoreCountDetail
	score.ScoreCountDetails(fmt.Sprintf("where user_id=%d and amount>0", uid))
	if scoreConfig.PayAmount > score.Using {
		status, messages = 0, golbal.Wrong[language][10105]
		return status, messages, "", 0, "", 0,0
	}
	address := ""
	if scoreConfig.IsGoods == 1 {
		userAddress := GetUserAddressOne(fmt.Sprintf("id=%d", addressId))
		if userAddress.Id == 0 {
			messages = golbal.Wrong[language][10116]
			return status, messages, "", 0, "", 0,0
		}
		add := Address{
			TrueName: userAddress.TrueName,
			Area:     userAddress.Area,
			Phone:    userAddress.Phone,
			Zip:      userAddress.Zip,
			Address:  userAddress.Address,
		}
		address = string(JSONMarshal(&add))
	}
	var users User
	db.Debug().Where("id=?", uid).First(&users)

	//if scoreConfig.Id == 76 {
	//	if users.UserType != "" && users.UserType[:1] == "L" {
	//		c := 0
	//		db.Debug().Where("user_id=? and goods_id=76", uid).Count(&c)
	//		if count > 0 {
	//			status, messages = 0, "该商品已达兑换上限"
	//			return status, messages
	//		}
	//	} else {
	//		status, messages = 0, "没有权限兑换该商品"
	//		return status, messages
	//	}
	//}
	scoreDetail := ScoreDetail{
		UserId:     uid,
		Title:      "",
		CreateTime: now,
		PayAmount:  scoreConfig.PayAmount,
		Amount:     scoreConfig.Amount * scoreConfig.UnitValue,
		Balance:    score.Using - scoreConfig.PayAmount,
		OrderNo:    "",
		GoodsId:    scoreConfig.Id,
		Address:    address,
	}
	goods := ""
	if scoreConfig.IsGoods == 1 {
		scoreDetail.OrderNo = scoreConfig.Unit + " " + scoreConfig.Name
		goods = scoreDetail.OrderNo
	} else {
		goods = fmt.Sprintf("%.2f %s", scoreConfig.Amount, scoreConfig.Unit)
	}
	//创建积分兑换详情
	tx := db.Begin()
	if err := tx.Debug().Create(&scoreDetail).Error; err != nil {
		log.Println("abc ExchangeGoods1 ", err)
		tx.Rollback()
		status, messages = 0, golbal.Wrong[language][10100]
		return status, messages, "", 0, "", 0,0
	}
	//创建金额详情
	if scoreConfig.IsGoods == 0 {
		interest := Interest{}
		interest.Login = 0
		interest.UserId = uid
		interest.Fee = scoreConfig.Amount * scoreConfig.UnitValue
		interest.CreateTime = now
		interest.Type = 3
		if err := tx.Debug().Create(&interest).Error; err != nil {
			log.Println("abc ExchangeGoods2 ", err)
			tx.Rollback()
			status, messages = 0, golbal.Wrong[language][10100]
			return status, messages, "", 0, "", 0,0
		}
		users.WalletBalance += interest.Fee
		if err := tx.Debug().Where("id=?", users.Id).Save(&users).Error; err != nil {
			log.Println("abc ExchangeGoods3 ", err)
			tx.Rollback()
			status, messages = 0, golbal.Wrong[language][10100]
			return status, messages, "", 0, "", 0,0
		}
	}
	//更新积分
	var scoreLog []ScoreLog
	pay := scoreConfig.PayAmount
	if err := tx.Debug().Where("user_id=? and amount>use_amount", uid).Order("id").Find(&scoreLog).Error; err != nil {
		log.Println("abc ExchangeGoods4 ", err)
		tx.Rollback()
		status, messages = 0, golbal.Wrong[language][10100]
		return status, messages, "", 0, "", 0,0
	}
	var useAll []int
	var lastId int
	for _, logs := range scoreLog {
		using := logs.Amount - logs.UseAmount
		if pay >= using {
			pay -= using
			useAll = append(useAll, logs.Id)
		} else {
			pay += logs.UseAmount
			lastId = logs.Id
			break
		}
	}

	if len(useAll) == 0 && lastId == 0 {
		tx.Rollback()
		status, messages = 0, golbal.Wrong[language][10100]
		return status, messages, "", 0, "", 0,0
	}
	if len(useAll) != 0 {
		if _, err := tx.Debug().Raw(`update score_log set use_amount = amount where user_id = ? and amount > 0 and id < ?`, uid, lastId).Rows(); err != nil {
			log.Println("abc ExchangeGoods5 ", err)
			tx.Rollback()
			status, messages = 0, golbal.Wrong[language][10100]
			return status, messages, "", 0, "", 0,0
		}
	}
	if lastId > 0 {
		if _, err := tx.Debug().Raw("update score_log set use_amount=? where user_id=? and id=?", pay, uid, lastId).Rows(); err != nil {
			log.Println("abc ExchangeGoods6 ", err)
			tx.Rollback()
			status, messages = 0, golbal.Wrong[language][10100]
			return status, messages, "", 0, "", 0,0
		}
	}
	tx.Commit()
	go func() {
		if scoreConfig.IsGoods == 0 {
			messageConfig := GetMessageConfig(15)
			SendMessage(uid, 15,
				fmt.Sprintf(messageConfig.ContentZh, fmt.Sprintf("%.2f", scoreConfig.PayAmount), fmt.Sprintf("%.2f%s", scoreConfig.Amount, scoreConfig.Unit)),
				fmt.Sprintf(messageConfig.ContentHk, fmt.Sprintf("%.2f", scoreConfig.PayAmount), fmt.Sprintf("%.2f%s", scoreConfig.Amount, scoreConfig.Unit)),
				fmt.Sprintf(messageConfig.ContentEn, fmt.Sprintf("%.2f", scoreConfig.PayAmount), fmt.Sprintf("%.2f%s", scoreConfig.Amount, scoreConfig.Unit)))
		}
	}()
	return status, messages, users.TrueName, scoreConfig.PayAmount, goods, scoreDetail.Balance,scoreConfig.IsGoods
}
