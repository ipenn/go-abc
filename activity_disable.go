package abc

import (
	"fmt"
	"strings"
)

func FindActivityDisableOne(right, path string) bool {
	ps := strings.Trim(path, ",")
	p := strings.Split(ps, ",")
	userId := p[len(p)-1]

	var data []ActivityDisable
	db.Debug().Model(&ActivityDisable{}).Where(fmt.Sprintf("type like '%%%s%%' and user_path = '%s' and match_type = 0", right, userId)).Find(&data)
	if len(data) != 0 {
		return true
	}

	db.Debug().Model(&ActivityDisable{}).Where(fmt.Sprintf("type like '%%%s%%' and user_path in (%s) and match_type = 1", right, ps)).Find(&data)
	for _, d := range data {
		if strings.Contains(path, d.UserPath) {
			return true
		}
	}
	return false
}

func FindActivityDisableAllUser(right, path string) []int {
	ps := strings.Trim(path, ",")
	p := strings.Split(ps, ",")
	var pathId []int
	for _, s := range p {
		pathId = append(pathId, ToInt(s))
	}
	var data []ActivityDisable
	var disableUser []int
	db.Debug().Model(&ActivityDisable{}).Where(fmt.Sprintf("type like '%%%s%%' and user_path in (%s)", right, ps)).
		Order(fmt.Sprintf("FIND_IN_SET(id,'%s')", ps)).Find(&data)
	for i, id := range pathId {
		for _, d := range data {
			if d.UserPath == ToString(id) && d.MatchType == 0 {
				disableUser = append(disableUser, id)
			} else if d.UserPath == ToString(id) && d.MatchType == 1 {
				disableUser = append(disableUser, pathId[i:]...)
				break
			}
		}
	}
	return disableUser
}

func GetDisableFeature(uid int, path string) string {
	result, _ := SqlOperator(`select group_concat(type) disable from activity_disable where user_path = ? and match_type = 0`, uid)
	res, _ := SqlOperator(`select group_concat(type) disable from activity_disable where find_in_set(user_path,?) and match_type = 1`, path)

	if res == nil && result == nil {
		return ""
	}
	r1 := ""
	r2 := ""
	if result != nil {
		r1 = PtoString(result, "disable")
	}

	if res != nil {
		r2 = PtoString(res, "disable")
	}

	r := r1 + "," + r2

	return strings.Join(RemoveDuplicatesAndEmpty(strings.Split(strings.Trim(r, ","), ",")), ",")
}
