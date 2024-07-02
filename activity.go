package abc

import (
	"fmt"
	golbal "github.com/chenqgp/abc/global"
	"gorm.io/gorm"
	"log"
	"sort"
	"strings"
	"time"
)

var (
	// LotteryLevel The deposit amount corresponds to the lottery level
	LotteryLevel = map[float64]int{
		200:   1,
		2000:  2,
		5000:  3,
		10000: 4,
		50000: 5,
	}
	// FirstPayIn The amount of the first deposit per day gets the corresponding number of lottery
	FirstPayIn = map[float64]int{
		200:   1,
		2000:  2,
		5000:  3,
		10000: 5,
	}
	// MaxLotteryCount Maximum number of lottery available
	MaxLotteryCount = map[string]int{
		"f": 5,
		"a": 5,
		"o": 1,
	}
	// InviteUser Invite how many users to get it lottery
	InviteUser = 2
)

func GetActivities(where string) (activities []Activities) {
	db.Debug().Where(where).Order("weight desc,create_time desc").Find(&activities)
	return activities
}

func FindLotteryConfig() (data []LotterySimple) {
	db.Debug().Model(LotteryConfig{}).Scan(&data)
	return data
}
func LotteryCount(uid int, t, where string) int64 {
	var count int64
	db.Debug().Model(LotteryDetail{}).Where(fmt.Sprintf("user_id=%v and level like '%v%%' %s", uid, t, where)).Count(&count)

	return count
}

func LotteryDetails(uid int, where, t string) []LotteryDetail {
	var lottery []LotteryDetail
	db.Debug().Where(fmt.Sprintf("user_id=%v %s and level like '%v%%'", uid, where, t)).
		Order("create_time").Find(&lottery)

	return lottery
}
func LotteryDetailsLimit(uid, page, size int, where, t string) (int64, []LotteryDetail) {
	var lottery []LotteryDetail
	search := fmt.Sprintf("user_id=%v %s and level like '%v%%'", uid, where, t)
	db.Debug().Where(search).Limit(size).Offset((page - 1) * size).Order("result_time desc").Find(&lottery)
	var count int64
	db.Debug().Model(LotteryDetail{}).Where(search).Count(&count)
	return count, lottery
}

func GetLotteryConfig(limit int, level string, id int) []LotteryConfigOne {
	var config []LotteryConfigOne
	where := ""
	if id == 0 {
		where = fmt.Sprintf("%s>0", level)
	} else {
		where = fmt.Sprintf("id=%d", id)
	}
	db.Debug().Select(fmt.Sprintf("%s as chance,id,name,type,value", level)).Model(LotteryConfig{}).
		Where(where).Order("id").Limit(limit).Scan(&config)
	return config
}

func CreateLottery(tx *gorm.DB, uid int, level, language string, amount float64) (int, string) {
	t := FormatNow()

	createCount, msg := GetLotteryCreateCount(tx, uid, level, language, amount)
	if createCount == -1 {
		return -1, msg
	}
	levelNum := ""
	if createCount > 0 {
		switch level {
		case "a":
			where := fmt.Sprintf(`status = 1 and (type = 'deposit' or type = 'withdraw') 
			and pay_time > '%s' and user_path like '%%,%d,%%'`, t[0:7], uid)
			sum := GetSumAmount(where)
			levelNum = GetLotteryLevelNum(sum)
		case "f":
			levelNum = GetLotteryLevelNum(amount)
		}
		level += levelNum
		sql := "insert into lottery_detail ( user_id, level, create_time) values "
		var values []string
		for i := 0; i < createCount; i++ {
			values = append(values, fmt.Sprintf(`(%d,'%s','%s')`, uid, level, t))
		}
		sql += strings.Join(values, ",")
		if _, err := tx.Debug().Raw(sql).Rows(); err != nil {
			log.Println(" abc CreateLottery ", err)
			return -1, fmt.Sprintf("user_id=%d 插入抽奖次数失败,等级:%s,次数:%d", uid, level, createCount)
		}
	}
	return 1, msg
}

func GetLotteryCreateCount(tx *gorm.DB, uid int, level, language string, amount float64) (int, string) {
	createCount := 0
	nowCount := LotteryCount(uid, level, " and status=0")
	if int(nowCount) >= MaxLotteryCount[level] {
		return createCount, golbal.Wrong[language][10110]
	}
	switch level {
	case "a":
		act := GetUserActivity(uid)
		if act.ByUser >= InviteUser-1 {
			act.ByUser = 0
			createCount++
		} else {
			act.ByUser++
		}
		if act.Id == 0 {
			act.UserId = uid
			act.CreateUserActivity()
		} else {
			//UpdateSql(UserActivity{}.TableName(), fmt.Sprintf("user_id=%d", uid), map[string]interface{}{
			//	"by_user": act.ByUser,
			//})
			err := tx.Debug().Model(UserActivity{}).Where(fmt.Sprintf("user_id=%d", uid)).Updates(map[string]interface{}{
				"by_user": act.ByUser,
			}).Error
			if err != nil {
				log.Println("abc GetLotteryCreateCount ", err)
				return -1, fmt.Sprintf("user_id=%d 代理活跃用户增加失败", uid)
			}
		}
	case "f":
		addCount := 0
		var payin []float64
		for f, _ := range FirstPayIn {
			payin = append(payin, f)
		}
		sort.Float64s(payin)
		where := fmt.Sprintf("user_id = %d and status = 1 and type = 'deposit' and create_time>'%s'", uid, FormatNow()[0:10])
		if amount >= payin[len(payin)-1] {
			amount = payin[len(payin)-1]
			where += fmt.Sprintf(" and amount>%f", amount)
		} else {
			for i := len(payin) - 1; i >= 0; i-- {
				if amount >= payin[i] {
					amount = payin[i]
					where += fmt.Sprintf(" and amount BETWEEN %f and %f", amount, payin[i+1])
					break
				}
			}
		}

		payment := GetPayment(where)
		if len(payment) == 0 {
			addCount = FirstPayIn[amount]
		} else {
			return 0, ""
		}

		if int(nowCount)+addCount >= MaxLotteryCount[level] {
			createCount = MaxLotteryCount[level] - int(nowCount)
		} else {
			createCount = addCount
		}
	case "o":
		if o := LotteryCount(uid, level, ""); o > 0 {
			return 0, ""
		} else {
			createCount++
		}
	}
	return createCount, ""
}

func GetLotteryLevelNum(sum float64) (levelNum string) {
	var keys []float64
	for k, _ := range LotteryLevel {
		keys = append(keys, k)
	}
	sort.Float64s(keys)
	if sum >= keys[len(keys)-1] {
		levelNum = ToString(LotteryLevel[keys[len(keys)-1]])
	} else {
		for i, k := range keys {
			if sum >= k && sum < keys[i+1] {
				levelNum = ToString(LotteryLevel[k])
			}
		}
	}
	return levelNum
}

func LotteryDraw(limit int, lottery LotteryDetail) LotteryConfigOne {
	level := lottery.Level
	if len(lottery.Level) > 1 {
		level = fmt.Sprintf("s%d", ToInt(level[1:]))
	}else if lottery.Level == "a" {
		level = "s1"
	}

	config := GetLotteryConfig(limit, level, lottery.AssignId)

	var chances []float64
	for _, one := range config {
		chances = append(chances, one.Chance)
	}

	res := Lotteried(chances...)

	if res < 0 {
		res = 0
	}
	if config[res].Id == 0 {
		res = 0
	}
	return config[res]
}

func GetDateData(dateType string, dateNum int, selects, timeName, table, where string, isFloat bool) (int, []DateData) {
	dataType := "%Y-%m-%d"
	var e time.Time
	var s time.Time
	var times []string
	start, end := "", ""
	if dateType == "day" {
		e = time.Now().AddDate(0, 0, -1)
		s = e.AddDate(0, 0, -dateNum+1)
		start = s.Format("2006-01-02")
		end = e.Format("2006-01-02")
		limit := 0
		for {
			if limit != 0 {
				s = s.AddDate(0, 0, 1)
			}
			times = append(times, s.Format("2006-01-02"))
			if s == e {
				break
			} else if limit == 366 {
				break
			}
			limit++
		}
	} else if dateType == "month" {
		for i := 1; i <= 12; i++ {
			str := ""
			if i < 10 {
				str = fmt.Sprintf("0%s", ToString(i))
			} else {
				str = ToString(i)
			}
			times = append(times, fmt.Sprintf("%s-%s", ToString(dateNum), str))
		}
		start = fmt.Sprintf("%s-01", ToString(dateNum))
		end = fmt.Sprintf("%s-12", ToString(dateNum))
		dataType = "%Y-%m"
	}
	if start == "" || end == "" {
		return 0, nil
	}
	search := fmt.Sprintf(`where date_format(%s,'%s') BETWEEN '%s' and '%s' %s`, timeName, dataType, start, end, where)
	sql := fmt.Sprintf(`
	SELECT
		%s,
		DATE_FORMAT( %s, '%s' ) AS date 
	FROM
		%s 
		%s
	GROUP BY
		date 
	ORDER BY
		date DESC `, selects, timeName, dataType, table, search)
	res, _ := SqlOperators(sql)

	var data []DateData
	for _, t := range times {
		var c float64
		for _, r := range res {
			if t == PtoString(r, "date") {
				c += ToFloat64(PtoString(r, "value"))
			}
		}
		v := ""
		if !isFloat {
			v = ToString(ToInt(c))
		} else {
			v = ToString(fmt.Sprintf("%.2f", c))
		}
		data = append(data, DateData{
			Date:  t,
			Value: v,
		})
	}
	return 1, data
}
