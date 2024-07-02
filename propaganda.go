package abc

import "fmt"

func PropagandaList(uid, t int, language string) (int, string, any) {
	user := GetUserById(uid)
	var pro []Propaganda
	where := fmt.Sprintf("cate=%d and status=1 and (user_type = '0' or user_type like '%%%s%%') and langs='%v'", t, user.UserType, language)
	db.Debug().Where(where).Order("weight desc, id desc").Find(&pro)

	data := make(map[string][]Propaganda, 0)
	for _, p := range pro {
		data[p.DocType] = append(data[p.DocType], p)
	}
	return 1, "", data
}
