package abc

import (
	"fmt"
	golbal "github.com/chenqgp/abc/global"
	"log"
)

func GetAuditedBankOne(uid int) Bank {
	var b Bank
	db.Debug().Where("user_id = ? and status = 1", uid).First(&b)

	return b
}

func GetAuditedBank(id, uid int) Bank {
	var b Bank
	db.Debug().Where("id = ? and user_id = ? and status = 1", id, uid).First(&b)

	return b
}

func GetMyBank(uid int) []interface{} {
	res, _ := SqlOperators("SELECT bank_name, bank_no, id, `default` FROM bank WHERE user_id = ? AND status = 1", uid)

	return res
}

func CheckBankNo(bankNo string) Bank {
	var bank Bank
	db.Debug().Where("bank_no = ?", bankNo).First(&bank)

	return bank
}

func GetBankCardById(id int) Bank {
	var bank Bank
	db.Debug().Where("id=?", id).First(&bank)
	return bank
}

func GetBankCardList(uid int) []Bank {
	var bank []Bank
	db.Debug().Where("user_id=?", uid).Find(&bank)
	return bank
}

func (bank Bank) SaveBank(where string) int {
	if err := db.Debug().Model(bank).Where(where).Save(&bank).Error; err != nil {
		log.Println(" abc SaveBank ", err)
		return 0
	}
	return 1
}

func DeleteBank(where string) int {
	if err := db.Debug().Where(where).Delete(&Bank{}).Error; err != nil {
		log.Println(" abc DeleteBank ", err)
		return 0
	}
	return 1
}

func SaveBankInformation(bankType int, user User, status int, name, bankName, bankNo, bankAddress, bankFile, swift, iban, language string, area, bankCardType int, chineseIdentity, bank_code string) (int, string, Bank) {
	var b Bank
	bank := CheckBankNo(bankNo)
	ui := GetUserInfoForId(user.Id)

	if ui.IdentityType != "Identity card" && bankType == 1 && user.UserType != "sales" && user.SalesType == "admin" {
		if chineseIdentity == "" {
			return 0, golbal.Wrong[language][10000], b
		} else {
			UpdateSql("user_info", fmt.Sprintf("user_id = %v", user.Id), map[string]interface{}{
				"chinese_identity": chineseIdentity,
			})
		}
	}
	if bank.Id > 0 {
		return 0, golbal.Wrong[language][10051], b
	}

	if bankType == 0 {
		return 0, golbal.Wrong[language][10000], b
	}

	switch bankType {
	case 1:
		if name == "" || bankName == "" || bankNo == "" || bankAddress == "" {
			return 0, golbal.Wrong[language][10000], b
		}
		b = Bank{
			UserId:       user.Id,
			BankName:     bankName,
			BankNo:       bankNo,
			TrueName:     name,
			CreateTime:   FormatNow(),
			Default:      0,
			BankAddress:  bankAddress,
			Area:         area,
			BankCardType: bankCardType,
			Status:       status,
			BankCode:     bank_code,
		}

		db.Debug().Create(&b)

	case 2:
		if name == "" || bankName == "" || bankNo == "" || bankAddress == "" || bankFile == "" || bankFile == "[]" {
			return 0, golbal.Wrong[language][10000], b
		}
		b = Bank{
			UserId:       user.Id,
			BankName:     bankName,
			BankNo:       bankNo,
			TrueName:     name,
			CreateTime:   FormatNow(),
			Default:      1,
			BankAddress:  bankAddress,
			Area:         area,
			BankCardType: bankCardType,
			Status:       status,
			Files:        bankFile,
			BankCode:     bank_code,
		}

		db.Debug().Create(&b)
	case 3:
		if bankName == "" || bankNo == "" || bankAddress == "" || swift == "" ||
			iban == "" || bankFile == "" || bankFile == "[]" {
			return 0, golbal.Wrong[language][10000], b
		}
		b = Bank{
			UserId:      user.Id,
			BankName:    bankName,
			BankNo:      bankNo,
			TrueName:    name,
			CreateTime:  FormatNow(),
			Default:     2,
			BankAddress: bankAddress,
			Status:      status,
			Files:       bankFile,
			Iban:        iban,
			Swift:       swift,
			BankCode:    bank_code,
		}

		db.Debug().Create(&b)
	default:
		return 0, golbal.Wrong[language][10000], b
	}

	return 1, "", b
}
