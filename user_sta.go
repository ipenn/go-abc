package abc

import (
	"gorm.io/gorm"
)

type UserSta struct {
	UserId	int `json:"user_id"`
	ParentId	int `json:"parent_id"`
	Mobile	string `json:"mobile"`
	Email	string `json:"email"`
	TrueName	string `json:"true_name"`
	RegTime	string `json:"reg_time"`
	UserPath	string `json:"user_path"`
	Status	int `json:"status"`
	AuthStatus	int `json:"auth_status"`
}

func CreateUserSta(tx *gorm.DB, i UserSta) error {
	if err := db.Debug().Create(&i).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
