package abc

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

func GetPaymentConfigOne(name string) PaymentConfig {
	var p PaymentConfig
	db.Debug().Where("name = ? ", name).First(&p)

	return p
}
func GetPaymentUnipayRate() []PaymentConfig {
	var p []PaymentConfig
	t := time.Now().Format("15:03:04")
	db.Debug().Where("status = 1 and open_time <= ? and close_time >= ?", t, t).Find(&p)

	return p
}
func StatisticalTransferAmount(account int) float64 {
	res, _ := SqlOperator(`select ifnull(sum(amount),0) as total_amount from payment where type = 'transfer' and status = 1 and transfer_login = ?`, account)

	if res != nil {
		return ToFloat64(PtoString(res, "total_amount"))
	}

	return 0
}

func CreatePaymentTransfer(p Payment) Payment {
	db.Debug().Create(&p)

	return p
}

func WithdrawalConfiguration() interface{} {
	res, _ := SqlOperator(`select GROUP_CONCAT(value SEPARATOR ',') as temp from config where id in (2,4,5,6,16)`)

	return res
}

func NumberWithdrawalsMonth(uid int) int64 {
	var count int64
	db.Debug().Table("payment").Where("status >= 0 and user_id = ? and create_time >= ? and type = 'withdraw' and find_in_set(pay_name,'UnionPay,Help2pay')", uid, time.Now().Format("2006-01")).Count(&count)

	return count
}

func GetPaymentOne(where string) (payment Payment) {
	db.Debug().Where(where).First(&payment)
	return
}

func UpdatePayment(where, update any) {
	db.Debug().Table("payment").Where(where).Updates(update)
}

func GetPaymentChannel() []PaymentConfig {
	var pSlice []PaymentConfig
	db.Debug().Where("status = 1 and type = 0").Find(&pSlice)

	return pSlice
}

func GetPayment(where string) (payment []Payment) {
	db.Debug().Where(where).Order("create_time desc").Find(&payment)
	return payment
}

func CheckUserIsDeposit(uid int) bool {
	var count int64
	db.Debug().Table("payment").Where("user_id = ? and status = 1 and type = 'deposit'", uid).Count(&count)

	if count == 0 {
		return false
	}

	return true
}

func GetSumAmount(where string) float64 {
	data := struct {
		Amount float64 `json:"amount"`
	}{}
	db.Debug().Select("sum(amount+pay_fee) as amount").Model(Payment{}).
		Where(where).Scan(&data)
	return data.Amount
}

func GetPaymentList(page, size int, where string) (int, int64, []PaymentSimple2) {
	var pay []PaymentSimple2
	db.Debug().Model(Payment{}).Where(where).Limit(size).Offset((page - 1) * size).Order("create_time desc").Scan(&pay)
	for i, payment := range pay {
		if payment.Type == "withdraw" {
			pay[i].Amount = math.Abs(payment.Amount)
			if payment.Status == 0 {
				if payment.AStatus == 0 || payment.BStatus == 0 {
					pay[i].Status = 4
				} else if payment.AStatus == 1 && payment.BStatus == 1 {
					pay[i].Status = 5
				}
			}
		} else if payment.Type == "deposit" {
			if payment.Status == 0 {
				if payment.PayName == "USDT" {
					if payment.CStatus == 1 {
						pay[i].Status = 4
					}
				} else if payment.PayName == "Wire" {
					if payment.WaterNumber != "" && payment.WireDoc != "" {
						pay[i].Status = 4
					}
				}
			}
		}
		//pay[i].Amount = math.Abs(payment.Amount)
		//pay[i].PayFee = math.Abs(payment.PayFee)
		pay[i].WireDoc = ""
		pay[i].WaterNumber = ""
	}
	var count int64
	db.Debug().Model(Payment{}).Where(where).Count(&count)
	return 1, count, pay
}

func GetPaymentSimple(re bool, uid, page, size int, t, where string) (count int64, data []PaymentSimple, total Total) {
	var redis RedisListData
	var result string
	if re {
		redisData := RDB.Get(Rctx, fmt.Sprintf("GetPayment-%d-%d-%s", uid, size, t))
		result = redisData.Val()
	}
	if result == "" {
		db.Debug().Select(`p.id,p.order_no,p.user_id,u.true_name,p.create_time,p.amount,p.pay_name,p.pay_time,p.comment,p.status,p.type,p.pay_fee
			,p.pay_company_amount,u.user_type`).
			Table("payment p").Joins("left join user u on p.user_id=u.id").Where(where).Limit(size).Offset((page - 1) * size).
			Order("create_time desc").Scan(&data)
		db.Debug().Table("payment p").Joins("left join user u on p.user_id=u.id").Where(where).Count(&count)

		if t == "deposit" {
			db.Debug().Select("sum(pay_company_amount) as amount").Table("payment p").Joins("left join user u on p.user_id=u.id").Where(where).Scan(&total)
		} else {
			db.Debug().Select("sum(amount+pay_fee) as amount").Table("payment p").Joins("left join user u on p.user_id=u.id").Where(where).Scan(&total)
		}

		if re {
			redis.Total = total
			redis.List = ToJSON(data)
			redis.Count = count
			str := ToJSON(redis)
			RDB.Set(Rctx, fmt.Sprintf("GetPayment-%d-%d-%s", uid, size, t), str, 10*time.Second)
		}
	} else {
		json.Unmarshal([]byte(result), &redis)
		json.Unmarshal(redis.List, &data)
		count = redis.Count
		total = redis.Total
	}
	return count, data, total
}

func QueryUnpaidOrders(uid int) Payment {
	var p Payment
	db.Debug().Where("user_id = ? and type = 'deposit' AND `status` = 0 AND pay_name = 'USDT'", uid).First(&p)

	return p
}

func GetPaymentByOrderId(orderNo string) Payment {
	var p Payment
	db.Debug().Where("order_no = ?", orderNo).First(&p)

	return p
}

func UpdatePaymentState(tx *gorm.DB, p Payment, amount float64) bool {
	if err := tx.Debug().Model(&Payment{}).Where("id = ?", p.Id).Updates(map[string]interface{}{
		"status":             1,
		"pay_time":           FormatNow(),
		"amount":             amount,
		"pay_company_amount": amount,
	}).Error; err != nil {
		tx.Rollback()

		return false
	}

	if err := tx.Debug().Model(&User{}).Where("id = ?", p.UserId).Updates(map[string]interface{}{
		"wallet_balance": gorm.Expr(fmt.Sprintf(`wallet_balance + %.2f`, amount+p.PayFee)),
	}).Error; err != nil {
		tx.Rollback()
		return false
	}

	return true
}

func ActivateCashCoupon(tx *gorm.DB, uid int) error {
	var cSlice []CashVoucher
	tx.Debug().Where("end_time > ? and user_id = ? and status = 0", FormatNow(), uid).Find(&cSlice)

	if len(cSlice) > 0 {
		if err := tx.Debug().Model(&CashVoucher{}).Where("end_time > ? and user_id = ? and status = 0", FormatNow(), uid).Updates(map[string]interface{}{
			"status": 1,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, v := range cSlice {
			credit := CreditDetail{
				UserId:     v.UserId,
				Login:      0,
				CreateTime: FormatNow(),
				OverTime:   ToAddDay(30),
				Balance:    v.Amount,
				Source:     3,
				Equity:     0,
				Status:     0,
				Comment:    v.Comment,
				Volume:     v.Volume,
				CouponNo:   v.CashNo,
				Deposit:    0,
			}

			if err := tx.Debug().Create(&credit).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func CreatePayment(orderNo, t, payName, comment, file, path, url, waterNumber string, amount, fee, exRate float64, uid, login, cStatus int) int {
	if comment == "7Star" {
		url = ""
	}

	payment := Payment{
		OrderNo:      orderNo,
		UserId:       uid,
		Login:        login,
		CreateTime:   t,
		Amount:       amount,
		PayName:      payName,
		Intro:        comment,
		PayFee:       fee,
		Status:       0,
		CStatus:      ToFloat64(cStatus),
		Type:         "deposit",
		WireDoc:      file,
		ExchangeRate: exRate,
		UserPath:     path,
		PayUrl:       url,
		WaterNumber:  waterNumber,
	}

	if err := db.Debug().Create(&payment).Error; err != nil {
		return 0
	}

	return payment.Id
}

func CreateWithdrawPayment(OrderNo, CreateTime, PayName, Type, Intro, Path, comment string, UserId, Login, Status, TransferLogin int, Amount, AmountRmb, PayFee float64) (int, Payment) {
	tx := db.Begin()

	p := Payment{
		OrderNo:       OrderNo,
		UserId:        UserId,
		Login:         Login,
		CreateTime:    CreateTime,
		Amount:        Amount,
		AmountRmb:     AmountRmb,
		PayName:       PayName,
		PayFee:        PayFee,
		Status:        Status,
		Type:          Type,
		Intro:         Intro,
		UserPath:      Path,
		TransferLogin: TransferLogin,
		Comment:       comment,
	}

	if err := tx.Debug().Create(&p).Error; err != nil {
		tx.Rollback()
		return 0, Payment{}
	}

	if err := tx.Debug().Model(&User{}).Where("id = ?", UserId).Updates(map[string]interface{}{
		"wallet_balance": gorm.Expr(fmt.Sprintf("wallet_balance + (%.2f)", Amount+PayFee)),
	}).Error; err != nil {
		tx.Rollback()
		return 0, Payment{}
	}

	tx.Commit()

	return 1, p
}

func GetPayConfigByName(name string) PaymentConfig {
	var p PaymentConfig
	db.Debug().Where("name = ?", name).First(&p)

	return p
}

func UserIsSuccessDeposited(uid int) Payment {
	var p Payment
	db.Debug().Where("user_id = ? and type = 'deposit' and status = 1 and pay_name = 'UnionPay'", uid).First(&p)

	return p
}

func UserIsSuccessWithdraw(uid int) Payment {
	var p Payment
	db.Debug().Where("user_id = ? and type = 'withdraw' and status = 1", uid).First(&p)

	return p
}

func VerySignWithSha256RSA(message string, signature string, pubKey string) error {
	k, _ := base64.StdEncoding.DecodeString(pubKey)
	publicKey, err := LoadPublicKeyRSA(k)
	if err != nil {
		return err
	}
	h := crypto.Hash.New(crypto.SHA256)
	io.WriteString(h, message)
	hashed := h.Sum(nil)
	k2, err := base64.StdEncoding.DecodeString(signature)
	fmt.Println(err)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, k2)
}

// LoadPublicKeyRSA 从秘钥字节流(秘钥必须被格式化后的)中解析出公钥.
func LoadPublicKeyRSA(pubKey []byte) (*rsa.PublicKey, error) {
	//block, _ := pem.Decode(pubKey)
	//if block == nil {
	//	return nil, fmt.Errorf("block is nil")
	//}

	publicKey, err := x509.ParsePKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	return publicKey.(*rsa.PublicKey), nil
}

func NumberOfWithdrawals(uid int) int64 {
	var count int64
	db.Debug().Table("payment").Where("status >= 0 and user_id = ? and create_time >= ? and type = 'withdraw' and find_in_set(pay_name,'UnionPay,Help2pay')", uid, time.Now().Format("2006-01")).Count(&count)

	return count
}

func TransferTotalAmount(uid int) float64 {
	res, _ := SqlOperator(`select sum(amount+pay_fee) as sums from payment where status >= 0 and create_time >= ? and type = 'transfer' and user_id = ? and login > 0`, time.Now().Format("2006-01"), uid)

	if res != nil {
		return ToFloat64(PtoString(res, "sums"))
	}

	return 0.0
}

func GetPaymentAmount(where string) (data AmountInfo) {
	db.Debug().Model(Payment{}).Select("sum(amount) as amount,sum(pay_fee) as fee").Where(where).Scan(&data)
	return data
}

func UpdatePaymentStatus(id int) error {
	if err := db.Debug().Model(&Payment{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   1,
		"pay_time": FormatNow(),
	}).Error; err != nil {
		return err
	}

	return nil
}

func UpdateUserWallet(amount float64, uid int) error {
	if err := db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"wallet_balance": gorm.Expr(fmt.Sprintf("wallet_balance + (%.2f)", amount)),
	}).Error; err != nil {
		return err
	}

	return nil
}

func RefundCoupon(id, cType int) {
	var uv UserVipCash
	db.Debug().Where("pay_id = ?", id).First(&uv)

	if uv.Id == 0 {
		return
	}

	db.Debug().Model(&UserVipCash{}).Where("id = ?", uv.Id).Updates(map[string]interface{}{
		"status": 0,
		"pay_id": 0,
	})

	if cType == 1 {
		db.Debug().Model(&Payment{}).Where("id = ?", id).Updates(map[string]interface{}{
			"pay_fee": 0,
		})
	}
}

func SendRequest(method, urls string, payload io.Reader, header map[string][]string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest(method, urls, payload)
	if err != nil {
		panic("DoRequest 初始化网络失败 " + err.Error())
	}

	//req.Header.Set("Content-Type", "application/json")
	req.Header = header

	res, err := client.Do(req)
	if err != nil {
		panic("DoRequest network connection failed " + err.Error())
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic("DoRequest EOF " + err.Error())
	}

	return body
}

func PaymentLimit(cType int, path string) []interface{} {
	res, _ := SqlOperators(`SELECT name, max_amount, min_amount, quick_pay, weight FROM payment_config WHERE type = ? AND status = 1 group by name ORDER BY weight DESC`, cType)

	flag := FindActivityDisableOne("open7Star", path)

	if !flag {
		var mSlice []interface{}
		for _, v := range res {
			if PtoString(v, "name") != "open7Star" {
				mSlice = append(mSlice, v)
			}
		}

		return mSlice
	}
	return res
}

func FinalDepositUsdt(uid int, flag bool) bool {
	u := GetUserById(uid)

	var p Payment
	db.Debug().Where("type = 'deposit' AND `status` = 1 AND user_id = ?", uid).Order("pay_time DESC").First(&p)

	if p.PayName == "USDT" || flag == true {
		return false
	}

	start := time.Now().AddDate(0, 0, -30).Format("2006-01-02") + " 00:00:00"
	var pSlice []Payment
	db.Debug().Where("pay_time <= ? AND pay_name != 'USDT' AND type = 'deposit' AND `status` = 1 AND user_id = ?", start, uid).Order("pay_time DESC").Find(&pSlice)
	var o Orders
	db.Debug().Where("user_id = ? and cmd < 2", u.Path).Order("open_time ASC").First(&o)

	var count int64
	if len(pSlice) != 0 {
		db.Debug().Table("orders").Where("user_id = ? and cmd < 2", u.Path).Count(&count)
	}

	if !flag {
		if len(pSlice) != 0 && time.Now().AddDate(0, 0, -90).Format("2006-01-02 15:04:05") >= o.OpenTime && count >= 100 {
			return false
		} else {
			return true
		}
	}

	return false
}

func CancelDeposit(id int) {
	db.Debug().Model(Payment{}).Where(fmt.Sprintf("id=%d", id)).Updates(map[string]interface{}{
		"status": -1,
	})
}



func ReturnDepositCoupon(orderId int) {
	var uv UserVipCash
	db.Debug().Where("pay_id = ?", orderId).First(&uv)

	if uv.Id == 0 {
		return
	}

	db.Debug().Model(&UserVipCash{}).Where("id = ?", uv.Id).Updates(map[string]interface{}{
		"status": 0,
		"pay_id": 0,
	})
}

func DetermineWithdrawalVolume(uid int, amount float64, cType int) (bool, float64, float64) {
	var need float64
	u := GetUserById(uid)
	x := 0.0
	res, _ := SqlOperator(`SELECT SUM(volume) volume FROM orders WHERE user_id = (SELECT path FROM user WHERE id = ?) AND symbol_type != 2`, uid)

	if res != nil {
		x = ToFloat64(PtoString(res, "volume"))
	}

	res1, _ := SqlOperators(`SELECT IFNULL(SUM(o.volume),0) volume, SUM(md2.user_amount) total_amount, md.user_amount amount, md.mam_id FROM mam_detail md
								 LEFT JOIN mam_project mp
								 ON md.mam_id = mp.id
								 LEFT JOIN orders o
								 ON o.login = mp.login
								 LEFT JOIN mam_detail md2
								 ON md.mam_id = md2.mam_id
								 WHERE md.user_id = ? GROUP BY md.mam_id`, uid)

	m := 0.0
	for _, v := range res1 {
		value := ToFloat64(fmt.Sprintf("%.2f", ToFloat64(PtoString(v, "volume"))*(ToFloat64(PtoString(v, "amount"))/ToFloat64(PtoString(v, "total_amount")))))
		m += value
	}

	f := x + m

	a, b := 0.0, 0.0

	var res2 interface{}

	if strings.Contains(u.UserType, "L") {
		res2, _ = SqlOperator(`SELECT SUM(amount+pay_fee) total_amount, SUM(IF(pay_name != 'USDT',amount+pay_fee,0)) amount FROM payment WHERE user_id = ? AND type = 'deposit' AND status = 1 AND create_time >= '2023-12-31 23:59:59'`, uid)
	} else {
		res2, _ = SqlOperator(`SELECT SUM(amount+pay_fee) total_amount, SUM(IF(pay_name != 'USDT',amount+pay_fee,0)) amount FROM payment WHERE user_id = ? AND type = 'deposit' AND status = 1`, uid)
	}

	if res2 != nil {
		a = ToFloat64(PtoString(res2, "total_amount"))
		b = ToFloat64(PtoString(res2, "amount"))
	}

	y := amount

	fmt.Println(fmt.Sprintf("a=%vb=%vx=%vneed=%vf=%vy=%v", a, b, x, need, f, y))

	if cType == 1 {
		if y <= a {
			need = y / 2000
		}

		if y > a {
			need = a / 2000
		}

		if y <= a && f >= y/2000 {
			return true, need, f
		}

		if y > a && f > a/2000 {
			return true, need, f
		}
	} else {
		if b > 0 {
			if y <= b {
				need = y / 100
			}

			if y > b {
				need = b/100 + (y-b)/2000
			}

			if y <= b && f >= y/100 {
				return true, need, f
			}

			if y > b && f >= (b/100)+((y-b)/2000) {
				return true, need, f
			}
		}

		if b == 0 {
			if y <= a {
				need = y / 2000
			}

			if y > a {
				need = a / 2000
			}

			if y <= a && f >= y/2000 {
				return true, need, f
			}

			if y > a && f > a/2000 {
				return true, need, f
			}
		}
	}
	return false, need, f
}

func DepositTimeJudgment(uid int, path string) bool {
	var p Payment
	db.Debug().Where("user_id = ? and type = 'deposit' and status = 1", uid).Order("create_time DESC").First(&p)

	var o Orders
	db.Debug().Where("user_id = ? and cmd < 2", path).Order("close_time DESC").First(&o)

	if o.CommissionTime > p.CreateTime {
		return true
	}

	return false
}
