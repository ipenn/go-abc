package abc

import (
	golbal "github.com/chenqgp/abc/global"
	"gorm.io/gorm"
)

const MarginLevel = 150

func GetMyAccount(uid int) []Account {
	var a []Account
	db.Debug().Where("user_id = ? and enable = 1 and login > 0", uid).Find(&a)

	return a
}

func UpdateAccountStatus(where string, m map[string]interface{}) error {
	err := db.Debug().Model(&Account{}).Where(where).Updates(m).Error

	return err
}

func IsExperienceAccount(uid int) bool {
	var count int64
	db.Debug().Table("account").Where("user_id = ? and experience != 0", uid).Count(&count)

	if count == 0 {
		return true
	}

	return false
}

func GetAccountOne(where string) (account Account) {
	db.Debug().Where(where).Order("reg_time desc").First(&account)
	return
}

func TxGetAccountOne(tx *gorm.DB, where string) (account Account) {
	tx.Debug().Where(where).First(&account)
	return
}

func GetAccounts(where string) (acc []Account) {
	db.Debug().Where(where).Find(&acc)
	return acc
}

func GetUserAccount(uid int, login string) Account {
	var a Account
	db.Debug().Where("user_id = ? and login = ?", uid, login).First(&a)

	return a
}
func CheckUserApplying(uid int) bool {
	var count int64
	db.Debug().Table("account").Where("user_id = ? and (enable = -1 or enable=-2) and login < 0", uid).Count(&count)

	if count == 0 {
		return false
	}

	return true
}

func InterestAccount(uid, readOnly int, groupName string, enable int, trueName, country, city, address, mobile, email string, rebateId int, leverage, language, path string) (int, string) {
	var a Account
	db.Debug().Order("login asc").First(&a)

	account := Account{
		Login:          a.Login - 1,
		UserId:         uid,
		RegTime:        FormatNow(),
		Balance:        0,
		GroupName:      groupName,
		Enable:         enable,
		Name:           trueName,
		Country:        country,
		City:           city,
		Address:        address,
		Phone:          mobile,
		Email:          email,
		Leverage:       leverage,
		RebateId:       rebateId,
		LeverageStatus: 1,
		ApplyLeverage:  0,
		AB:             "b",
		UserPath:       path,
		ReadOnly:       readOnly,
	}

	err := db.Debug().Create(&account).Error

	if err != nil {
		return 0, golbal.Wrong[language][10100]
	}

	return 1, ""
}

func swap(arr []OrdersData, i, j int) []OrdersData {
	temp := arr[j]
	arr[j] = arr[i]
	arr[i] = temp
	return arr
}

// 排序
func SortIntList(profit string, arr []OrdersData) []OrdersData {
	if len(arr) <= 1 {
		return arr
	}
	for i := 1; i < len(arr); i++ {
		for j := i - 1; j >= 0; j-- {
			switch profit {
			case "1":
				if ToFloat64(arr[j].Profit) > ToFloat64(arr[j+1].Profit) {
					swap(arr, j, j+1)
				}
			case "2":
				if ToFloat64(arr[j].Profit) < ToFloat64(arr[j+1].Profit) {
					swap(arr, j, j+1)
				}
			}
		}
	}
	return arr
}
