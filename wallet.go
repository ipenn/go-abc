package abc

import "log"

func GetResidueWalletCount() int64 {
	var count int64

	db.Debug().Table("user_wallet").Where("user_id = 0").Count(&count)

	return count
}

func GetWalletAddress(uid int) []interface{} {
	res, _ := SqlOperators("SELECT id, address, tag FROM wallet WHERE user_id = ?", uid)

	return res
}

func GetAuditedWallet(uid, id int) Wallet {
	var w Wallet

	db.Debug().Where("id = ? and status = 1 and user_id = ?", id, uid).First(&w)

	return w
}

func GetWalletById(id int) (wallet Wallet) {
	db.Debug().Where("id=?", id).First(&wallet)
	return wallet
}

func GetWalletList(uid int) []Wallet {
	var wallet []Wallet
	db.Debug().Where("user_id=?", uid).Find(&wallet)
	return wallet
}

func (wallet Wallet) CreateWallet() int {
	if err := db.Debug().Create(&wallet).Error; err != nil {
		log.Println(" abc CreateWallet ", err)
		return 0
	}
	return 1
}

func (wallet Wallet) SaveWallet(where string) int {
	if err := db.Debug().Model(wallet).Where(where).Updates(&wallet).Error; err != nil {
		log.Println(" abc SaveWallet ", err)
		return 0
	}
	return 1
}

func DeleteWallet(where string) int {
	if err := db.Debug().Where(where).Delete(&Wallet{}).Error; err != nil {
		log.Println(" abc DeleteWallet ", err)
		return 0
	}
	return 1
}

func GetWallet(uid any) (UserWallet, bool) {
	var wallet UserWallet
	db.Debug().Where("user_id = ?", uid).First(&wallet)

	return wallet, wallet.Id != 0
}

func WalletInsert(wallet UserWallet) {
	db.Debug().Create(&wallet)
}

func WalletCount(where any) (count int64) {
	var wallet UserWallet
	db.Debug().Model(&wallet).Where(where).Count(&count)
	return
}
