package abc

import (
	"log"
)

func (creditDetail CreditDetail) CreateCreditDetail() int {
	if err := db.Debug().Create(&creditDetail).Error; err != nil {
		log.Println(" abc CreateCreditDetail ", err)
		return 0
	}
	return 1
}

func GetCreditDetail(where string) (creditDetail []CreditDetail) {
	db.Debug().Where(where).Find(&creditDetail)
	return creditDetail
}

func GetCreditDetailList(page, size int, where string) (count int64, creditDetail []CreditDetail) {
	db.Debug().Where(where).Limit(size).Offset((page - 1) * size).Order("create_time desc").Find(&creditDetail)
	db.Debug().Model(CreditDetail{}).Where(where).Count(&count)
	return count, creditDetail
}

func GetCreditDetailOne(where string) (creditDetail CreditDetail) {
	db.Debug().Where(where).First(&creditDetail)
	return creditDetail
}
