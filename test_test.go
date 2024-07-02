package abc

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

const (
	uid = 16672
)

func TestLottery(t *testing.T) {
	res := make(map[int]int, 0)
	res[0] = 0
	res[1] = 0
	res[2] = 0
	res[3] = 0
	for i := 0; i < 100000; i++ {
		r := Lotteried(0.3, 0.5, 0.1, 0.1)
		res[r]++
	}
	fmt.Println(fmt.Sprintf("%+v", res))
}


func TestGetInviteCount(T *testing.T) {

	myinfo := GetUserById(uid)
	var inviteData []struct {
		Id        int    `json:"id"`
		UserType  string `json:"user_Type"`
		SalesType string `json:"sales_type"`
		GroupType string `json:"group_type"`
	}
	db.Debug().Raw(fmt.Sprintf(`SELECT
	u.id,
	u.user_type,
	u.sales_type,
	r.group_type 
FROM
	user u
	LEFT JOIN rebate_config r ON u.rebate_id = r.id 
WHERE
	u.path LIKE '%s%%'
ORDER BY
	u.id
`, myinfo.Path)).Scan(&inviteData)

	var count InviteCount
	for _, data := range inviteData {
		if data.Id != myinfo.Id {
			switch data.UserType {
			case "user":
				switch data.GroupType {
				case "STD":
					count.STD++
				case "DMA":
					count.DMA++
				}
			case "sales":
				switch data.SalesType {
				case "Level Ⅰ":
					count.Level1++
				case "Level Ⅱ":
					count.Level2++
				case "Level Ⅲ":
					count.Level3++
				}
			default:
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
	fmt.Println(fmt.Sprintf("%+v", count))
}

func TestGetUserPath(t *testing.T) {
	user := GetUserById(uid)
	fmt.Println(strings.Count(user.Path, ","))
}

func TestAddLotteryCount(t *testing.T) {
	tx := Tx()
	fmt.Println(CreateLottery(tx, 16697, "f", "CN", 10200))
}
func TestAddLotteryCount2(t *testing.T) {
	tx := Tx()
	fmt.Println(CreateLottery(tx, 16659, "o", "CN", 0))
}

func TestActivityDisable(t *testing.T) {
	path := ",0,1,16293,2287,5929,15136,13892,16672,16713,16714,16715,"
	ps := strings.Split(strings.Trim(path, ","), ",")
	userId := ps[len(ps)-1]
	fmt.Println(userId)
}

func TestFindActivityDisableAllUser(t *testing.T) {
	fmt.Println(fmt.Sprintf("%+v", FindActivityDisableAllUser("score",
		",0,1,16293,2287,5929,15136,13892,16672,16713,16714,16715,")))
}
func TestFindActivityDisableOne(t *testing.T) {
	fmt.Println(fmt.Sprintf("%+v", FindActivityDisableOne("CloseUnionPay",
		",0,1,16293,2287,5929,15136,13892,16672,")))
}

func TestUserVipUpgrade(t *testing.T) {
	UserVipUpgrade(16720)
}

func TestToJson(t *testing.T) {
	var pa []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	str := `[{"key":"mid","value":"3500639"},{"key":"secret","value":"QxgAEhlDaEW0VLUUowRA1HenXk3ZN2GF"},{"key":"url","value":"https://api.maxpay666.com/coin/pay/order/pay/checkout/counter"},{"key":"callback_key","value":"w0JSiaeuAN45pQw6mozRKAtsP4vJaloa"}]`
	json.Unmarshal([]byte(str), &pa)
	t.Log(fmt.Sprintf("%+v", pa))
}

func TestSalesGroup(t *testing.T) {
	var users []User
	// 我是三级代理
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,%' AND user_type != 'user'").Find(&users)
	a := SalesGroup([]User{}).AllSalesGroups(users)
	fmt.Println(len(a))
	for _, aa := range a {
		fmt.Println(aa)
	}
	fmt.Println("===================================================")
	// 我是三级业务员
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,%' AND user_type != 'user'").Find(&users)
	Ib := User{}
	Ib.Id = 16957
	users = append([]User{Ib}, users...)
	b := SalesGroup([]User{}).AllSalesGroups(users)
	b = b[1:]
	fmt.Println(len(b))
	for _, bb := range b {
		fmt.Println(bb)
	}
	fmt.Println("===================================================")
	// 我是二级代理
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,16959,%' AND user_type != 'user'").Find(&users)
	c := SalesGroup([]User{}).AllSalesGroups(users)
	fmt.Println(len(c))
	for _, cc := range c {
		fmt.Println(cc)
	}
	fmt.Println("===================================================")
	// 我是二级业务员
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,16959,16979,%' AND user_type != 'user'").Find(&users)
	Ib = User{}
	Ib.Id = 16959
	users = append([]User{Ib}, users...)
	d := SalesGroup([]User{}).AllSalesGroups(users)
	d = d[1:]
	fmt.Println(len(d))
	for _, dd := range d {
		fmt.Println(dd)
	}
	fmt.Println("===================================================")
	// 我是一级代理
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,16959,16961,%' AND user_type != 'user'").Find(&users)
	e := SalesGroup([]User{}).AllSalesGroups(users)
	fmt.Println(len(e))
	for _, ee := range e {
		fmt.Println(ee)
	}
	fmt.Println("===================================================")
	// 我是一级业务员
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,16959,16979,16980,16981,%' AND user_type != 'user'").Find(&users)
	Ib = User{}
	Ib.Id = 16980
	users = append([]User{Ib}, users...)
	f := SalesGroup([]User{}).AllSalesGroups(users)
	f = f[1:]
	fmt.Println(len(f))
	for _, ff := range f {
		fmt.Println(ff)
	}
}

func TestMatch(t *testing.T) {
	trueName := strings.ReplaceAll("JIN", " ", "") + " " + strings.ReplaceAll("DING DING", " ", "")
	match, err := regexp.MatchString(`^[A-Z a-z]+$`, trueName)
	if err != nil {
		fmt.Println("MatchString:", err.Error())
	}

}
func TestNodeGroup(t *testing.T) {
	var users []User
	// 我是三级代理
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,%' AND user_type != 'user'").Find(&users)
	a := NodeGroups(users)
	fmt.Println(len(a))
	for _, aa := range a {
		fmt.Println(aa)
	}
	fmt.Println("===================================================")
	// 我是三级业务员
	users = []User{}
	db.Debug().Select([]string{"id, email, user_type, sales_type, parent_id, sales_id, path"}).Where("path LIKE ',0,1,16955,16957,16958,%' AND user_type != 'user'").Find(&users)
	Ib := User{}
	Ib.Id = 16957
	users = append([]User{Ib}, users...)
	b := NodeGroups(users)
	b = b[1:]
	fmt.Println(len(b))
	for _, bb := range b {
		fmt.Println(bb)
	}
}
