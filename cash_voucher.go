package abc

import "log"

var VolAmountMap = map[float64]float64{
	50:   10,
	100:  20,
	200:  40,
	500:  125,
	1000: 300,
	2000: 650,
	5000: 1650,
}

func (cashVoucher CashVoucher) CreateCashVoucher() int {
	if err := db.Debug().Create(&cashVoucher).Error; err != nil {
		log.Println(" abc CreateCashVoucher ", err)
		return 0
	}
	return 1
}

func GetCashVoucher(where string) (cashVoucher []CashVoucher) {
	db.Debug().Where(where).Order("create_time desc").Find(&cashVoucher)
	return cashVoucher
}
