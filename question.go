package abc

import (
	"fmt"
	"log"
)

func GetQuestion(id int) (question Question) {
	db.Debug().Where("id=?", id).First(&question)

	return question
}

func GetQuestionDetail(qid int) (questionDetail []QuestionDetail) {
	db.Debug().Where("q_id=?", qid).Order("create_time").Find(&questionDetail)
	return questionDetail
}

func SaveQuestion(question Question) int {
	if err := db.Debug().Save(&question).Error; err != nil {
		return 0
	}
	return 1
}

func ReadQuestion(qid int) {
	db.Debug().Raw("update question_detail set is_read = 1,read_time = ? where q_id = ?", FormatNow(), qid).Rows()
}

// CreateQuestion 创建问题
func CreateQuestion(uid int, title, content, tag string) (int, Question) {
	time := FormatNow()
	question := Question{
		UserId:     uid,
		CreateTime: time,
		UpdateTime: time,
		Title:      title,
		Content:    content,
		Status:     1,
		Weight:     1,
		Tag:        tag,
	}
	if err := db.Debug().Create(&question).Error; err != nil {
		log.Println("abc CreateQuestion ", err)
		return 0, Question{}
	}
	return 1, question
}

// CreateQuestionDetail 创建问题详情
func CreateQuestionDetail(qid int, content, file string) int {
	detail := QuestionDetail{
		QId:        qid,
		AdminId:    0,
		Content:    content,
		CreateTime: FormatNow(),
		ReadTime:   "",
		IsRead:     1,
		FilePath:   file,
		FileName:   "",
		TelegramId: 0,
	}
	if err := db.Debug().Create(&detail).Error; err != nil {
		log.Println("abc CreateQuestionDetail ", err)
		return 0
	}
	return 1
}

func QuestionList(uid, status, page, size int) (int, string, int64, any) {
	var question []Question
	where := fmt.Sprintf("user_id=%d and status=%d", uid, status)
	db.Debug().Where(where).Limit(size).Offset((page - 1) * size).Order("update_time desc").Find(&question)
	for i, q := range question {
		question[i].Tag = ReplaceTag(q.Tag)
	}
	var count int64
	db.Debug().Model(Question{}).Where(where).Count(&count)
	return 1, "", count, question
}

func SwitchTag(tag string) string {
	switch tag {
	case "1":
		return "存款问题"
	case "2":
		return "取款问题"
	case "3":
		return "开户问题"
	case "4":
		return "佣金问题"
	case "5":
		return "其他问题"
	default:
		return ""
	}
}

func ReplaceTag(tag string) string {
	switch tag {
	case "存款问题":
		return "1"
	case "取款问题":
		return "2"
	case "开户问题":
		return "3"
	case "佣金问题":
		return "4"
	case "其他问题":
		return "5"
	default:
		return ""
	}
}
