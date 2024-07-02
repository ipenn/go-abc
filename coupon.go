package abc

import "log"

func (coupon Coupon) CreateCoupon() int {
	if err := db.Debug().Create(&coupon).Error; err != nil {
		log.Println(" abc CreateCoupon ", err)
		return 0
	}
	return 1
}

func GetCoupon(where string) (coupon []Coupon) {
	db.Debug().Where(where).Order("create_time desc").Find(&coupon)
	return coupon
}

func GetCouponOne(where string) (coupon Coupon) {
	db.Debug().Where(where).First(&coupon)
	return coupon
}
