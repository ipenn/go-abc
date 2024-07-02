package payNotice

import (
	"fmt"
	"github.com/chenqgp/abc"
	"github.com/chenqgp/abc/third/brevo"
	"github.com/chenqgp/abc/third/telegram"
	"strings"
)

func SuccessPayNotice(payId, state int, amount float64) {
	p := abc.GetPaymentOne(fmt.Sprintf("id =%v", payId))
	u := abc.GetUserById(p.UserId)
	u1 := abc.GetUserById(u.ParentId)

	//发送邮件
	mail := abc.MailContent(6)
	content := fmt.Sprintf(mail.Content, u.TrueName, fmt.Sprintf("%.2f", p.Amount))

	brevo.Send(mail.Title, content, u.Email)

	//用户存款给代理发邮件
	mail2 := abc.MailContent(71)
	content2 := fmt.Sprintf(mail2.Content, u.TrueName, fmt.Sprintf("%.2f", p.Amount))
	brevo.Send(mail.Content, content2, u1.Email)

	//发送站内信
	message := abc.GetMessageConfig(19)

	abc.SendMessage(u.Id, 19, fmt.Sprintf(message.ContentZh, abc.ToString(amount/p.ExchangeRate)), fmt.Sprintf(message.ContentHk, abc.ToString(amount/p.ExchangeRate)), fmt.Sprintf(message.ContentEn, abc.ToString(amount/p.ExchangeRate)))

	res, err := abc.SqlOperator(`select GROUP_CONCAT(CONCAT(true_name,'(',user_type,')') order by find_in_set(id,?) asc) as field1 from user where find_in_set(id,?)`, u.Path, u.Path)
	newPath := ""
	if err == nil {
		path := abc.PtoString(res, "field1")
		ps := strings.Split(path, ",")
		curIndex := -2
		for index, item := range ps {
			if curIndex == -2 && (strings.Index(item, "Level") > -1) {
				curIndex = index - 1
			}
			if curIndex > -2 {
				newPath += "," + item
			}
		}
		if curIndex > 0 {
			newPath = ps[curIndex] + newPath
		}
	}

	res2, err := abc.SqlOperator(`select group_concat(login) as logins from account where user_id = ? and login > 0`, p.UserId)
	accounts := ""
	if err == nil {
		accounts = abc.PtoString(res2, "logins")
	}

	teleData := fmt.Sprintf("NO.=%s 存款成功通知：%.2f (实际金额：%.2f) %s 。 %s [%s]", p.OrderNo, p.Amount+p.PayFee, amount/p.ExchangeRate, p.PayName+p.Intro, newPath, accounts)

	telegram.SendMsg(telegram.TEXT, 10, teleData)
	telegram.SendMsg(telegram.TEXT, 21, teleData)

	if abc.FindActivityDisableOne("SendHKDspMsg", u.Path) {
		telegram.SendMsg(telegram.TEXT, 26, teleData)
	}
	if abc.FindActivityDisableOne("SendTZDspMsg", u.Path) {
		telegram.SendMsg(telegram.TEXT, 25, teleData)
	}
	if abc.FindActivityDisableOne("SendTBDspMsg", u.Path) {
		telegram.SendMsg(telegram.TEXT, 24, teleData)
	}
	if abc.FindActivityDisableOne("SendHMDspMsg", u.Path) {
		telegram.SendMsg(telegram.TEXT, 23, teleData)
	}
	if state == 0 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单号:%v成功支付更新订单状态失败", p.OrderNo))
	}
}

func SuccessPay(orderNo string, amount float64) (int, string) {
	tx := abc.Tx()
	p := abc.GetPaymentByOrderId(orderNo)
	u := abc.GetUserById(p.UserId)

	if p.Id == 0 || p.Status == 1 {
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单号:%v原因%v", p.OrderNo, "订单不存在或状态已完成更新"))
		tx.Commit()
		return 0, "订单不存在或状态已完成更新"
	}

	//修改我的金额
	if !abc.UpdatePaymentState(tx, p, amount) {
		tx.Rollback()
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单号:%v原因%v", p.OrderNo, "修改金额错误"))
		return 0, "修改金额错误"
	}

	//激活现金券
	if err := abc.ActivateCashCoupon(tx, p.UserId); err != nil {
		tx.Rollback()
		telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("订单号:%v原因%v", p.OrderNo, "激活现金券失败"))
		return 0, "激活现金券失败"
	}

	if !abc.FindActivityDisableOne("lottery", u.Path) {
		status, msg := abc.CreateLottery(tx, p.UserId, "f", "CN", amount+p.PayFee)

		if status == -1 {
			tx.Rollback()
			telegram.SendMsg(telegram.TEXT, 0, msg)
			return 0, "增加抽奖次数失败"
		}
	}

	//用户是否是第一次存款
	var count int64
	tx.Debug().Table("payment").Where("user_id = ? and type = 'deposit' and status = 1", p.UserId).Count(&count)

	if count == 1 {
		u1 := abc.GetUserById(u.ParentId)

		if !abc.FindActivityDisableOne("lottery", u1.Path) {
			status, msg := abc.CreateLottery(tx, u.ParentId, "a", "CN", p.Amount)

			if status == -1 {
				tx.Rollback()
				telegram.SendMsg(telegram.TEXT, 0, msg)

				return 0, "增加抽奖次数失败"
			}
		}

		res, _ := abc.SqlOperator(`select GROUP_CONCAT(CONCAT(id,',',user_type) order by find_in_set(id,?) desc) as pathFull from user where find_in_set(id,?)`, u.Path, u.Path)
		idNum := 0
		arr := strings.Split(abc.PtoString(res, "pathFull"), ",")
		if res != nil {
			for i := 3; i < len(arr); i += 2 {
				if arr[i] != "sales" {
					idNum = abc.ToInt(i - 1)
					break
				}
			}

			if idNum != 0 && !abc.FindActivityDisableOne("score", u.Path) {
				uv := abc.GetUserVipById(abc.ToInt(arr[idNum]))
				var vipConfig abc.UserVipConfig
				tx.Debug().Where("grade_id = ?", uv.Grade).First(&vipConfig)

				s := &abc.ScoreLog{
					Symbol:     "Referral",
					UserId:     abc.ToInt(arr[idNum]),
					Type:       1,
					CreateTime: abc.FormatNow(),
					Amount:     abc.ToFloat64(vipConfig.Invite),
				}

				if err := tx.Debug().Create(&s).Error; err != nil {
					tx.Rollback()
					telegram.SendMsg(telegram.TEXT, 0, fmt.Sprintf("用户id:%v原因:%v", abc.ToInt(arr[idNum]), err))

					return 0, "用户增加积分失败"
				}
			}
		}
	}
	//messageConfig := GetMessageConfig(19)
	//SendMessage(p.UserId)

	tx.Commit()
	return 1, ""
}
