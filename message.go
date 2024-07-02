package abc

import (
	"fmt"
	"gorm.io/gorm"
)

func SendMessage(uid, templateId int, ContentZh, ContentHk, ContentEn string) {

	m := &UserMessage{
		UserId:     uid,
		CreateTime: FormatNow(),
		TemplateZh: ContentZh,
		TemplateHk: ContentHk,
		TemplateEn: ContentEn,
		TemplateId: templateId,
	}

	db.Debug().Create(&m)
}

func GetMessageConfig(id int) UserMessageConfig {
	var uc UserMessageConfig
	db.Debug().Where("id = ?", id).First(&uc)

	return uc
}

func GetUserMessageCount(where string) (count int64) {
	db.Debug().Model(UserMessage{}).Where(where).Count(&count)
	return count
}

func GetAnnouncementList(uid, page, size int, lang string) (int64, []Announcement) {
	user := GetUserById(uid)
	where := fmt.Sprintf("(who = '0' or who like '%%%s%%') and lang = '%s' and status = 1 and type=0", user.UserType, lang)
	var ann []Announcement
	db.Debug().Where(where).Limit(size).Offset((page - 1) * size).Order("create_time desc,id desc").Find(&ann)
	var count int64
	db.Debug().Model(Announcement{}).Where(where).Count(&count)
	return count, ann
}

func AnnouncementRead(id int) {
	db.Debug().Model(Announcement{}).Where("id=?", id).Updates(map[string]any{
		"clicks": gorm.Expr("clicks + (1)")})
}

func UserMessageList(uid, page, size int) (int64, []UserMessage) {
	var userMessage []UserMessage
	where := fmt.Sprintf("user_id=%d", uid)
	db.Debug().Where(where).Order("status asc,create_time desc").
		Limit(size).Offset((page - 1) * size).Find(&userMessage)

	return GetUserMessageCount(where), userMessage
}
