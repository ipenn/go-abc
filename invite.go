package abc

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

func CreateInvite(tx *gorm.DB, i InviteCode) error {
	if err := db.Debug().Create(&i).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetMyInvite(uid int) []InviteCode {
	var iSlice []InviteCode

	db.Debug().Where("user_id = ?", uid).Find(&iSlice)

	return iSlice
}

func GetInviteIsExit(code string) InviteCode {
	var i InviteCode

	db.Debug().Where("code = ?", code).First(&i)

	return i

}

func DeleteInvite(tx *gorm.DB, uid int) error {
	if err := tx.Debug().Where("user_id = ?", uid).Delete(&InviteCode{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func CheckUserHaveCode(uid int) bool {
	var count int64
	db.Debug().Table("invite_code").Where("user_id = ?", uid).Count(&count)

	if count == 0 {
		return false
	}

	return true
}

func GetInviteCodeOne(where string) (invite InviteCode) {
	db.Debug().Where(where).First(&invite)
	return invite
}

func GetInviteCodeSet(userType string) (ics []InviteCodeSet) {
	db.Debug().Where(fmt.Sprintf("user_type = '%s' and is_admin = 0", userType)).Find(&ics)
	return ics
}

func GetInviteCode(uid int) (inviteCode []InviteCode) {
	db.Debug().Select(`i.id,i.user_id,i.code,i.comment,r.display_name AS name,i.rights,i.type`).
		Table("invite_code i").Joins("LEFT JOIN rebate_config r ON i.name = r.group_name").
		Where("i.user_id = ?", uid).Scan(&inviteCode)
	return inviteCode
}

func GetInviteInfo(uid int) (info InviteInfo) {
	db.Debug().Raw(`SELECT
	u.true_name,
	u.user_type,
	u.ib_no,
	ui.agreement,
	ui.agreement_fee,
	ui.agreement_time
FROM
	user u
	LEFT JOIN user_info ui ON u.id = ui.user_id 
WHERE
	u.id = ?`, uid).Scan(&info)
	return info
}

func GetInviteCount(uid int) InviteCount {

	where := " u.user_type!='sales' and u.auth_status=1"
	group := ""
	myinfo := GetUserById(uid)
	if strings.Contains(myinfo.UserType, "Level") {
		where += fmt.Sprintf(" and u.parent_id=%d", myinfo.Id)
	} else if strings.Contains(myinfo.UserType, "sales") || strings.Contains(myinfo.UserType, "user") {
		where += fmt.Sprintf(" and LEFT ( path, %d)= '%s'", len(myinfo.Path), myinfo.Path)
		group = fmt.Sprintf(" group by SUBSTRING_INDEX( path ,',', FIND_IN_SET(%d,path) + 1)", myinfo.Id)
	}

	inviteData := GetDirectInvitation(where, group)
	var count InviteCount
	for _, data := range inviteData {
		if data.Id != myinfo.Id {
			if data.UserType == "user" {
				switch data.GroupType {
				case "STD":
					count.STD++
				case "DMA":
					count.DMA++
				}
			} else if data.UserType[:1] == "L" {
				switch data.UserType {
				case "Level Ⅰ":
					count.Level1++
				case "Level Ⅱ":
					count.Level2++
				case "Level Ⅲ":
					count.Level3++
				}
			}
		}
	}
	return count
}

func CreateInviteCode(user User) {
	inviteCode := GetInviteCode(user.Id)

	up := GetUser(fmt.Sprintf("id=%d", user.ParentId))

	if len(inviteCode) > 0 {
		var ids []string
		for _, code := range inviteCode {
			ids = append(ids, ToString(code.Id))
		}
		db.Debug().Where(fmt.Sprintf("id in (%s) and rights!='user'", strings.Join(ids, ","))).Delete(&InviteCode{})
	}

	sets := GetInviteCodeSet(user.UserType)

	var in []InviteCodeSet
	if up.UserType[0:1] == "L" {
		upInvite := GetMyInvite(user.ParentId)
		for _, set := range sets {
			for _, code := range upInvite {
				if code.Name != "" && set.CodeType == code.Name {
					in = append(in, set)
				}else if code.Name == "" && set.CodeType == code.Type {
					in = append(in, set)
				}
			}
		}
	} else {
		for _, set := range sets {
			if set.Status == 1 {
				in = append(in, set)
			}
		}
	}
	insert := "insert into invite_code (user_id, code, name, rights, type) VALUE "
	var codeList []string
	for _, item := range in {
		name := ""
		t := item.CodeType
		if item.CodeType[0:1] != "L" {
			t = strings.Split(item.CodeType, "-")[1]
			name = item.CodeType
		}
		code := ""
		for {
			code = RandStr(6)
			if invite := GetInviteCodeOne(fmt.Sprintf("code='%s'", code)); invite.Id == 0 {
				break
			}
			time.Sleep(100 * time.Microsecond)
		}
		codeList = append(codeList, fmt.Sprintf(`(%d,'%s','%s','%s','%s')`, user.Id, code, name, user.UserType, t))
	}
	value := strings.Join(codeList, ",")
	_, err := db.Debug().Raw(fmt.Sprintf(`%s %s`, insert, value)).Rows()
	if err != nil {
		log.Println(" abc CreateInviteCode ", err)
	}
}
