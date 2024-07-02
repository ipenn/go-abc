package abc

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

func DeleteToken(uid int) {
	db.Debug().Where("uid = ?", uid).Delete(&Token{})
}

func FetchFromToken(t string) Token {
	var token Token
	if len(t) != 60 {
		return token
	}

	db.Debug().Where("token = ?", t).First(&token)
	return token
}

func UserHandleToken(uid, role int) string {
	var t Token
	db.Debug().Where("uid = ?", uid).First(&t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(fmt.Sprintf("%x", uid)), bcrypt.DefaultCost)
	expire, _ := strconv.ParseInt(strconv.FormatInt(time.Now().Add(+5*time.Hour).Unix(), 10), 10, 64)

	if t.ID == 0 {
		t.Uid = uid
		t.Role = role
		t.UpdateTime = &sql.NullString{
			String: time.Now().Format("2006-01-02 15:04:05"),
			Valid:  true,
		}
		t.Expire = expire
		t.Token = string(hashedPassword)

		db.Debug().Create(&t)

		return t.Token
	}

	db.Debug().Model(&Token{}).Where("uid = ?", uid).Updates(map[string]interface{}{
		"update_time": &sql.NullString{
			String: time.Now().Format("2006-01-02 15:04:05"),
			Valid:  true,
		},
		"expire": expire,
		"token":  string(hashedPassword),
		"role":   role,
	})

	return string(hashedPassword)
}
