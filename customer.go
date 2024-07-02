package abc

import "gorm.io/gorm"

func UpdateCustomer(tx *gorm.DB, username string) error {
	var c Customer
	tx.Debug().Where("email = ?", username).First(&c)

	if c.Id == 0 {
		return nil
	}

	if err := tx.Debug().Model(&Customer{}).Where("id = ?", c.Id).Updates(map[string]interface{}{
		"status": 1,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func CreateCustomer(uid int, email string) {
	var customer Customer
	db.Debug().Where("email = ?", email).First(&customer)

	if customer.Id == 0 {
		c := Customer{
			UserId:     uid,
			Email:      email,
			CreateTime: FormatNow(),
		}

		db.Debug().Create(&c)
		return
	}

	db.Debug().Table("customer").Where("id = ?", customer.Id).Updates(map[string]interface{}{
		"user_id":     uid,
		"create_time": FormatNow(),
	})
}
