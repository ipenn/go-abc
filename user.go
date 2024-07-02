package abc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	file "github.com/chenqgp/abc/third/uFile"
	ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
	"github.com/xuri/excelize"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	golbal "github.com/chenqgp/abc/global"
	"gorm.io/gorm"
)

func GetUserCountByIdentity(identity string) int64 {
	var count int64
	db.Debug().Table("user_info").Where("identity = ?", identity).Count(&count)

	return count
}

func GetPathIb(uid int) string {
	res, _ := SqlOperator(`SELECT GROUP_CONCAT(id) path FROM user WHERE FIND_IN_SET(id,(SELECT path FROM user WHERE id = ?))`, uid)

	return strings.Trim(PtoString(res, "path"), ",")
}

func TxUpdateUserWallet(tx *gorm.DB, where, update any) error {
	if err := tx.Debug().Model(&User{}).Where(where).Updates(update).Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func TxCoefficientCommissionByUser(tx *gorm.DB, where string) (userType []struct {
	Id          int     `json:"id"`
	UserType    string  `json:"user_type"`
	RebateCate  int     `json:"rebate_cate"`
	SomeId      int     `json:"some_id"`
	ParentId    int     `json:"parent_id"`
	RebateMulti float64 `json:"rebate_multi"`
}) {
	tx.Debug().
		Table("user").
		Select([]string{"id", "user_type", "rebate_cate", "some_id", "parent_id", "rebate_multi"}).
		Where(where).Order(where + " DESC").Scan(&userType)
	return
}

func CheckUserPhone(phone, area_code string) User {
	var u User
	db.Debug().Where("mobile = ? and phonectcode = ?", phone, area_code).First(&u)

	return u
}

func GetUserFile(uid int) []UserFile {
	var uf []UserFile
	db.Debug().Where("user_id = ?", uid).Find(&uf)

	return uf
}

func GetUserById(uid int) (u User) {
	db.Debug().Where("id = ?", uid).First(&u)
	return u
}
func GetUser(where string) (u User) {
	db.Debug().Where(where).First(&u)
	return u
}
func GetUsers(where string) (u []User) {
	db.Debug().Where(where).Find(&u)
	return u
}

func SaveUser(user User) int {
	if err := db.Debug().Model(User{}).Where("id=?", user.Id).Updates(&user).Error; err != nil {
		log.Println(" abc SaveUser ", err)
		return 0
	}
	return 1
}

func GetUserInfoById(uid int) (u UserInfo) {
	db.Debug().Where("user_id = ?", uid).First(&u)
	return u
}

func SaveUserInfo(userInfo UserInfo) int {
	if err := db.Debug().Model(UserInfo{}).Where("user_id = ?", userInfo.UserId).Updates(&userInfo).Error; err != nil {
		log.Println(" abc SaveUserInfo ", err)
		return 0
	}
	return 1
}

func GetUserIdAndUserTypeByPath(path string) (data []UserIdAndUserType) {
	db.Debug().Raw(fmt.Sprintf(`select id,user_type
	from user where find_in_set(id,'%s')`, path)).Scan(&data)
	return data
}

func GetUserMore(uid int) (more UserMore) {
	db.Debug().Where("user_id=?", uid).First(&more)
	return more
}

func SaveUserMore(more UserMore) int {
	if err := db.Debug().Model(UserMore{}).Where("user_id = ?", more.UserId).Updates(&more).Error; err != nil {
		log.Println(" abc SaveUserMore ", err)
		return 0
	}
	return 1
}

func GetUserActivity(uid int) (act UserActivity) {
	act.UserId = uid
	db.Debug().Where("user_id=?", uid).FirstOrCreate(&act)
	return act
}

func (userActivity UserActivity) CreateUserActivity() int {
	if err := db.Debug().Create(&userActivity).Error; err != nil {
		log.Println(" abc CreateUserActivity ", err)
		return 0
	}
	return 1
}

func (userAddress UserAddress) CreateUserAddress() int {
	if err := db.Debug().Create(&userAddress).Error; err != nil {
		log.Println(" abc CreateUserAddress ", err)
		return 0
	}
	return userAddress.Id
}

func GetUserAddress(where string) (userAddress []UserAddress) {
	db.Debug().Where(where).Find(&userAddress)
	return userAddress
}

func GetUserAddressOne(where string) (userAddress UserAddress) {
	db.Debug().Where(where).First(&userAddress)
	return userAddress
}

func GetUsersSimple(where, order string) (users []UserSimple) {
	if order == "" {
		order = "id desc"
	}
	db.Debug().Model(User{}).Where(where).Order(order).Scan(&users)
	return users
}

// GetDirectInvitation Get the someone that belong to me
func GetDirectInvitation(where, group string) (data []DirectInvitation) {
	redisData := RDB.Get(Rctx, fmt.Sprintf("GetDirectInvitation-%s", where))
	result := redisData.Val()
	if result == "" {
		db.Debug().Raw(fmt.Sprintf(`SELECT
		u.id,
		u.user_type,
		u.create_time,
		u.true_name,
		u.email,
		u.mobile,
		u.auth_status,
		u.status,
		r.group_type
	FROM
		user u
	LEFT JOIN 
		rebate_config r ON u.rebate_id = r.id
	WHERE
	%s
	%s 
	ORDER BY
		u.id`, where, group)).Find(&data)
		str := ToJSON(data)
		RDB.Set(Rctx, fmt.Sprintf("GetDirectInvitation-%s", where), str, 10*time.Second)
	} else {
		json.Unmarshal([]byte(result), &data)
	}
	return data
}

func GetDirectInvitationList(re bool, page, size, uid int, path, where string) (c int64, data []DirectInvitation) {
	var redis RedisListData
	var result string
	if re {
		redisData := RDB.Get(Rctx, fmt.Sprintf("invite-%d-%d", uid, size))
		result = redisData.Val()
	}

	if result == "" {
		db.Debug().Raw(fmt.Sprintf(`SELECT
	u.id,
	u.user_type,
	u.create_time,
	u.true_name,
	u.email,
	u.mobile,
	u.auth_status,
	u.STATUS,
	u.path,
	IFNULL( p.count, 0 ) AS count 
FROM
	user u
	LEFT JOIN (
	SELECT
		user_id,
		count(id) as count
	FROM
		payment 
	WHERE
		STATUS = 1 
		AND type = 'deposit' 
		AND amount > 0 
		AND LEFT ( user_path, %d )= '%s' 
		group by SUBSTRING_INDEX( user_path ,',', FIND_IN_SET(%d,user_path) + 1)
	) p ON p.user_id = u.id 
WHERE
	LEFT ( path, %d )= '%s'  and id!=%d
	%s
	group by SUBSTRING_INDEX( path ,',', FIND_IN_SET(%d,path) + 1)
ORDER BY
	u.id 
	LIMIT %d,%d`, len(path), path, uid, len(path), path, uid, where, uid, (page-1)*size, size)).Scan(&data)
		operators, _ := SqlOperator(fmt.Sprintf(`SELECT count(c.count) as count FROM (SELECT
	COUNT(u.id) as count
FROM
	user u
	LEFT JOIN (
	SELECT
		user_id,
		count(id) as count
	FROM
		payment 
	WHERE
		STATUS = 1 
		AND type = 'deposit' 
		AND amount > 0 
		AND LEFT ( user_path, %d )= '%s' 
		group by SUBSTRING_INDEX( user_path ,',', FIND_IN_SET(%d,user_path) + 1)
	) p ON p.user_id = u.id 
WHERE
	LEFT ( path, %d )= '%s' and id!=%d
	%s
	group by SUBSTRING_INDEX( path ,',', FIND_IN_SET(%d,path) + 1)) c`, len(path), path, uid, len(path), path, uid, where, uid))
		c = ToInt64(PtoString(operators, "count"))

		for i, datum := range data {
			data[i].UserStatus = SwitchStatus(datum.Status, datum.AuthStatus, datum.Count)
		}

		if re {
			redis.List = ToJSON(data)
			redis.Count = c
			str := ToJSON(redis)
			RDB.Set(Rctx, fmt.Sprintf("invite-%d-%d", uid, size), str, 10*time.Second)
		}
	} else {
		json.Unmarshal([]byte(result), &redis)
		json.Unmarshal(redis.List, &data)
	}
	return c, data
}

func CheckUserExists(username string) User {
	var u User
	db.Debug().Where("email = ?", username).First(&u)

	return u
}

func GetUserRole(userType, salesType string) int {
	switch userType {
	case "user":
		return 1
	case "Level Ⅰ":
		return 3
	case "Level Ⅱ":
		return 5
	case "Level Ⅲ":
		return 7
	case "sales":
		switch salesType {
		case "admin":
			return 8
		case "Level Ⅰ":
			return 2
		case "Level Ⅱ":
			return 4
		case "Level Ⅲ":
			return 6

		}
	}

	return 0
}

func GetUserAccounts(uid int) string {
	rows, err := db.Debug().Raw(`SELECT IFNULL(GROUP_CONCAT(login),'') login FROM account WHERE user_id = ?`, uid).Rows()

	if err != nil {
		log.Println("api GetUserAccount ", err)
	}

	result := HandleRawSQL(rows)

	return PtoString(result[0], "login")
}

func UpdateLoginNum(uid int) {
	db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"login_time": &sql.NullString{
			time.Now().Format("2006-01-02 15:04:05"),
			true,
		},
		"login_times": gorm.Expr("login_times + 1"),
	})
}

// 生成用户的path_full
func GetUserPathFull(tx *gorm.DB, mode, path string) string {
	if mode == "" {
		mode = "id,',',user_type"
	}
	arr := strings.Trim(path, ",")
	rows, err := tx.Debug().Raw(fmt.Sprintf("select GROUP_CONCAT(CONCAT(%s) order by find_in_set(id,'%s') asc) as path_full from user where find_in_set(id,'%s')", mode, arr, arr)).Rows()
	defer rows.Close()

	if err != nil {
		log.Println("abc GetUserPathFull ", err)
	}

	return PtoString(HandleRawSQL(rows)[0], "path_full")
}

// 获取用户组
func GetAccountGroup(name string) RebateConfig {
	var r RebateConfig
	db.Debug().Where("group_name = ?", name).First(&r)

	return r
}

// 获取佣金等级
func GetCommissionLevel(userType string) int {
	switch userType {
	case "Level Ⅰ":
		return 1
	case "Level Ⅱ":
		return 4
	case "Level Ⅲ":
		return 5
	}

	return 0
}

func SaveUserInformation(uid, cType int, language, surname, lastname, nationality, identityType, identity, birthday, title, currencyType, accountType, platform, forexp, investfreq, incomesource, employment, business, position, company, idFront, idBack, other, birthcountry, country, address, address_date, income, netasset, ispolitic, istax, isusa, isforusa, isearnusa, otherexp, investaim string, idType int) (int, string, interface{}) {
	switch cType {
	case 1:
		if surname == "" || lastname == "" || nationality == "" || identityType == "" || identity == "" || title == "" || birthday == "" || birthcountry == "" || address == "" {
			return 0, golbal.Wrong[language][10000], ""
		}
		trueName := strings.ReplaceAll(surname, " ", "") + " " + strings.ReplaceAll(lastname, " ", "")
		match, _ := regexp.MatchString(`^[A-Z a-z]+$`, trueName)
		if !match {
			return 0, golbal.Wrong[language][10008], nil
		}
		db.Debug().Model(&UserInfo{}).Where("user_id = ? ", uid).Updates(map[string]interface{}{
			"surname":       surname,
			"lastname":      lastname,
			"nationality":   nationality,
			"identity_type": identityType,
			"identity":      identity,
			"title":         title,
			"birthday":      birthday,
			"birthcountry":  birthcountry,
			"country":       country,
			"address":       address,
			"address_date":  address_date,
		})

		db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"account_status": 1,
			"true_name":      strings.ToUpper(surname) + " " + strings.ToUpper(lastname),
		})
	case 2:
		if currencyType == "" || platform == "" || forexp == "" || investfreq == "" || otherexp == "" || accountType == "" || investaim == "" {
			return 0, golbal.Wrong[language][10000], ""
		}
		db.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
			"currency_type": currencyType,
			"account_type":  accountType,
			"platform":      platform,
			"forexp":        forexp,
			"investfreq":    investfreq,
			"otherexp":      otherexp,
		})

		db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
			"investaim": investaim,
		})

		db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"transaction_status": 1,
		})
	case 3:
		if incomesource == "" || employment == "" || business == "" || position == "" || isusa == "" || isforusa == "" || isearnusa == "" || istax == "" || ispolitic == "" || income == "" || netasset == "" {
			return 0, golbal.Wrong[language][10000], ""
		}
		db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
			"incomesource": incomesource,
			"employment":   employment,
			"business":     business,
			"position":     position,
			"income":       income,
			"netasset":     netasset,
			"ispolitic":    ispolitic,
			"istax":        istax,
			"isusa":        isusa,
			"isforusa":     isforusa,
			"isearnusa":    isearnusa,
		})

		db.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
			"company": company,
		})

		db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"financial_status": 1,
		})
	case 4:
		if idType == 0 {
			return 0, golbal.Wrong[language][10000], ""
		}

		switch idType {
		//身份证
		case 1:
			if idFront == "" || idBack == "" {
				return 0, golbal.Wrong[language][10000], ""
			}
			uf := UserFile{
				FileName:   idFront,
				FileType:   "ID",
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				UserId:     uid,
				Status:     0,
				Front:      "0",
			}
			db.Debug().Create(&uf)
			uf2 := UserFile{
				FileName:   idBack,
				FileType:   "ID",
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				UserId:     uid,
				Status:     0,
				Front:      "1",
			}
			db.Debug().Create(&uf2)
			//护照
		case 2:
			if idFront == "" {
				return 0, golbal.Wrong[language][10000], ""
			}
			uf3 := UserFile{
				FileName:   idFront,
				FileType:   "ID",
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				UserId:     uid,
				Status:     0,
				Front:      "2",
			}
			db.Debug().Create(&uf3)
			//驾照
		case 3:
			if idFront == "" {
				return 0, golbal.Wrong[language][10000], ""
			}
			uf3 := UserFile{
				FileName:   idFront,
				FileType:   "ID",
				CreateTime: time.Now().Format("2006-01-02 15:04:05"),
				UserId:     uid,
				Status:     0,
				Front:      "2",
			}
			db.Debug().Create(&uf3)
			//银行卡
		case 4:

		}
		db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"documents_status": 1,
		})
	default:
		return 0, golbal.Wrong[language][10000], ""
	}
	return 1, "", nil
}

// 身份验证成功，修改用户KYC状态
func ModifyUserProfileStatus(uid int, chineseName, address, language string) (int, string) {
	tx := db.Begin()

	if err := tx.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"chinese_name":     chineseName,
		"identity_address": address,
		"info_status":      2,
	}).Error; err != nil {
		tx.Rollback()
		return 0, golbal.Wrong[language][10100]
	}

	if err := tx.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"auth_status": 1,
	}).Error; err != nil {
		tx.Rollback()
		return 0, golbal.Wrong[language][10100]
	}

	if err := tx.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"account_status":     2,
		"transaction_status": 2,
		"financial_status":   2,
		"documents_status":   2,
	}).Error; err != nil {
		tx.Rollback()
		return 0, golbal.Wrong[language][10100]
	}

	if err := tx.Debug().Model(&UserFile{}).Where("user_id= ? and status != 1 and file_type = 'ID'", uid).Updates(map[string]interface{}{
		"status": 1,
	}).Error; err != nil {
		tx.Rollback()
		return 0, golbal.Wrong[language][10100]
	}

	tx.Commit()
	return 1, ""
}

// 帮用户账号申请
func AddUserAccountApplication(uid int, ui UserInfo, language string) {
	u := GetUserById(uid)
	rc := RebateConfig{}
	db.Debug().Where("id = ?", u.RebateId).First(&rc)

	code := ""
	for true {
		time.Sleep(100 * time.Microsecond)
		code = RandStr(6)

		if GetInviteIsExit(code).Id == 0 {
			break
		}
	}

	i := InviteCode{
		UserId: u.Id,
		Code:   code,
		Name:   rc.GroupName,
		Rights: "user",
		Type:   rc.GroupType,
	}
	db.Debug().Create(&i)

	var account Account
	db.Debug().Order("login asc").First(&account)
	readOnly := 0

	if strings.Contains(rc.GroupName, "DMA") {
		readOnly = 1
	}

	a := Account{
		Login:          account.Login - 1,
		UserId:         u.Id,
		RegTime:        time.Now().Format("2006-01-02 15:04:05"),
		Balance:        0,
		GroupName:      rc.GroupName,
		Enable:         -1,
		Name:           u.TrueName,
		Country:        ui.Country,
		City:           ui.City,
		Address:        ui.Address,
		Phone:          u.Mobile,
		Email:          u.Email,
		Comment:        "",
		Leverage:       "100",
		RebateId:       rc.Id,
		IsMam:          0,
		AB:             "b",
		LeverageStatus: 1,
		ApplyLeverage:  0,
		UserPath:       u.Path,
		ReadOnly:       readOnly,
	}

	db.Debug().Create(&a)
}

func WriteUserLog(userId int, logType string, ip string, content string) {
	userLog := UserLog{
		UserId:     userId,
		Type:       logType,
		CreateTime: FormatNow(),
		Ip:         strings.Trim(ip, " "),
		Content:    content,
	}
	db.Debug().Create(&userLog)
}

func UploadIdDocument(uid int, language string, i IdDocument, u User, oldIdFront, oldIdBack, oldOther string) string {
	switch i.IdType {
	case 1:
		if i.IdFront == "" || i.IdBack == "" {
			return golbal.Wrong[language][10524]
		}
	case 2:
		if i.IdFront == "" {
			return golbal.Wrong[language][10524]
		}
	case 3:
		if i.IdFront == "" {
			return golbal.Wrong[language][10524]
		}
	case 4:
		if i.IdFront == "" {
			return golbal.Wrong[language][10524]
		}
	default:
		return golbal.Wrong[language][10524]
	}

	res, _ := SqlOperator(`select ifnull(group_concat(file_name),'') name from user_file where user_id = ? and file_type = 'ID'`, uid)
	if res != nil {
		if PtoString(res, "name") != "" {
			arr := strings.Split(PtoString(res, "name"), ",")
			s := strings.Trim(strings.Trim(oldIdFront+","+oldIdBack, ",")+","+oldOther, ",")

			if s != "" {
				oldArr := strings.Split(s, ",")
				DelFile(RemoveDuplicateData(arr, oldArr))
			} else {
				DelFile(arr)
			}
		}
	}

	db.Debug().Where("user_id = ? AND file_type = 'ID'", uid).Delete(&UserFile{})

	if i.IdFront != "" {
		uf := UserFile{
			FileName:   i.IdFront,
			FileType:   "ID",
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
			UserId:     uid,
			Status:     0,
			Front:      "1",
		}
		db.Debug().Create(&uf)
	}

	if i.IdBack != "" {
		uf2 := UserFile{
			FileName:   i.IdBack,
			FileType:   "ID",
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
			UserId:     uid,
			Status:     0,
			Front:      "2",
		}
		db.Debug().Create(&uf2)
	}

	if i.Other != "" {
		uf3 := UserFile{
			FileName:   i.Other,
			FileType:   "ID",
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
			UserId:     uid,
			Status:     0,
			Front:      "3",
		}
		db.Debug().Create(&uf3)
	}

	db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"documents_status": 1,
	})

	fmt.Println("i.Surname:", i.Surname)
	fmt.Println("i.Lastname:", i.Lastname)
	fmt.Println("u.TrueName:", u.TrueName)

	if strings.ToUpper(i.Surname)+" "+strings.ToUpper(i.Lastname) != u.TrueName {
		trueName := strings.ReplaceAll(i.Surname, " ", "") + " " + strings.ReplaceAll(i.Lastname, " ", "")
		match, err := regexp.MatchString(`^[A-Z a-z]+$`, trueName)
		if err != nil {
			fmt.Println("MatchString:", err.Error())
		}
		if !match {
			fmt.Println("match fail:", match)
			return golbal.Wrong[language][10524]
		}

		if err = db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
			"true_name": strings.ToUpper(i.Surname) + " " + strings.ToUpper(i.Lastname),
		}).Error;err != nil {
			fmt.Println("===保存kyc第四步姓名失败===err2===",err)
		}

		if err = db.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
			"surname":  i.Surname,
			"lastname": i.Lastname,
		}).Error;err != nil {
			fmt.Println("===保存kyc第四步姓名失败===err===",err)
		}
	}

	return ""
}

func CheckUsernameAndPassword(username, password string) User {
	var u User
	db.Debug().Where("email = ? and password = ?", username, password).First(&u)

	return u
}

func GetRebateForId(id int) RebateConfig {
	var r RebateConfig
	db.Debug().Where("id = ?", id).First(&r)

	return r
}

func CheckInviteCode(code string) InviteCode {
	var i InviteCode
	db.Debug().Where("code = ?", code).First(&i)

	return i
}

func CreateUser(tx *gorm.DB, u User) (User, error) {
	if err := tx.Debug().Create(&u).Error; err != nil {
		tx.Rollback()
		return User{}, err
	}

	return u, nil
}

func UpdateUserPath(tx *gorm.DB, uid int, path string) error {
	path =  path + strconv.Itoa(uid) + ","
	if err := tx.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"path": path,
		"path_full": GetUserPathFull(tx, "id,',',user_type,',',true_name", path),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func CreateUserRelatedTables(tx *gorm.DB, uid int) error {
	ui := &UserInfo{
		UserId: uid,
		AgreementTime: &sql.NullString{
			"",
			false,
		},
	}
	if err := tx.Debug().Create(&ui).Error; err != nil {
		tx.Rollback()
		return err
	}

	um := &UserMore{
		UserId: uid,
	}
	if err := tx.Debug().Create(&um).Error; err != nil {
		tx.Rollback()
		return err
	}

	uv := &UserVip{
		UserId: uid,
		Grade:  1,
	}
	if err := tx.Debug().Create(&uv).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func CreateMoveReward(uid int) {
	moveReward := &MoveRewards{
		UserId:     uid,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	db.Debug().Create(&moveReward)
}

func SaveAccountInformation(uid int, a AccountInformation, language string) (int, string, interface{}) {
	//if surname == "" || lastname == "" || nationality == "" || identityType == "" || identity == "" || title == "" || birthday == "" || birthcountry == "" || address == "" {
	//	return 0, golbal.Wrong[language][10000], ""
	//}
	if err := db.Debug().Model(&UserInfo{}).Where("user_id = ? ", uid).Updates(map[string]interface{}{
		"nationality":   a.Nationality,
		"identity_type": a.IdentityType,
		"identity":      a.Identity,
		"title":         a.Title,
		"birthday":      a.Birthday,
		"birthcountry":  a.Birthcountry,
		"country":       a.Country,
		"address":       a.Address,
		"address_date":  a.AddressDate,
	}).Error; err != nil {
		fmt.Println("======保存kyc第一步====err=",err)
	}

	//db.Debug().Model(&User{}).Where("id = ?", uid).Updates(map[string]interface{}{
	//	"true_name": strings.ToUpper(a.Surname) + " " + strings.ToUpper(a.Lastname),
	//})

	if err := db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"account_status": 1,
		"phonectcode":    a.OldPhonectcode,
		"mobile":         a.OldMobile,
	}).Error; err != nil {
		fmt.Println("======保存kyc第一步==222222==err=",err)
	}

	return 1, "", nil
}

func SaveTransaction(uid int, language string, t Transaction) (int, string, interface{}) {
	//if currencyType == "" || platform == "" || forexp == "" || investfreq == "" || otherexp == "" || accountType == "" || investaim == "" {
	//	return 0, golbal.Wrong[language][10000], nil
	//}
	db.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"currency_type": t.CurrencyType,
		"account_type":  t.AccountType,
		"platform":      t.Platform,
		"forexp":        t.Forexp,
		"investfreq":    t.Investfreq,
		"otherexp":      t.Otherexp,
	})

	db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"investaim":          t.Investaim,
		"transaction_status": 1,
	})

	return 1, "", nil
}

func SaveFinancialInformation(uid int, language string, f FinancialInformation) (int, string, interface{}) {
	//if incomesource == "" || employment == "" || business == "" || position == "" || isusa == "" || isforusa == "" || isearnusa == "" || istax == "" || ispolitic == "" || income == "" || netasset == "" {
	//	return 0, golbal.Wrong[language][10000], nil
	//}
	db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"incomesource": f.Incomesource,
		"employment":   f.Employment,
		"business":     f.Business,
		"position":     f.Position,
		"income":       f.Income,
		"netasset":     f.Netasset,
		"ispolitic":    f.Ispolitic,
		"istax":        f.Istax,
		"isusa":        f.Isusa,
		"isforusa":     f.Isforusa,
		"isearnusa":    f.Isearnusa,
	})

	db.Debug().Model(&UserInfo{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"company": f.Company,
	})

	db.Debug().Model(&UserMore{}).Where("user_id = ?", uid).Updates(map[string]interface{}{
		"financial_status": 1,
	})

	return 1, "", nil
}

func GetUserInfoForId(uid int) UserInfo {
	var ui UserInfo
	db.Debug().Where("user_id = ?", uid).First(&ui)

	return ui
}

func UpdateSql(table, where string, m map[string]interface{}) error {
	err := db.Debug().Table(table).Where(where).Updates(m).Error

	return err
}

func ObtainAnnualIncome(uid int, language string) ([]interface{}, []interface{}, []interface{}) {
	start := fmt.Sprintf("%d-01-01 00:00:00", time.Now().Year()-2)
	end := fmt.Sprintf("%d-12-31 23:59:59", time.Now().Year())
	res, _ := SqlOperators(`SELECT SUM(fee) fee, a.m months, a.y years FROM (SELECT IFNULL(SUM(fee),0) fee,MONTH(create_time) m, 0 type, YEAR(create_time) y FROM interest WHERE user_id = ? AND create_time BETWEEN ? AND ? GROUP BY YEAR(create_time), MONTH(create_time)
								 UNION ALL
								 SELECT SUM(fee) fee,MONTH(create_time) m, 1 type, YEAR(create_time) y FROM commission WHERE ib_id = ? AND create_time BETWEEN ? AND ? GROUP BY YEAR(create_time), MONTH(create_time)) a GROUP BY  a.y, a.m`, uid, start, end, uid, start, end)

	var res1, res2, res3 []interface{}
	for _, v := range res {
		if time.Now().Year() == ToInt(PtoString(v, "years")) {
			res1 = append(res1, v)
		}
		if time.Now().Year()-1 == ToInt(PtoString(v, "years")) {
			res2 = append(res2, v)
		}
		if time.Now().Year()-2 == ToInt(PtoString(v, "years")) {
			res3 = append(res3, v)
		}
	}

	return res1, res2, res3
}

func ObtainAnnualIncome1(uid int, language string) (map[int]float64, map[int]float64, map[int]float64) {
	start := fmt.Sprintf("%d-01-01 00:00:00", time.Now().Year()-2)
	end := fmt.Sprintf("%d-12-31 23:59:59", time.Now().Year())

	m := make(map[int]float64)
	m1 := make(map[int]float64)
	m2 := make(map[int]float64)

	var res []interface{}
	res1, _ := SqlOperators(`SELECT IFNULL(fee,0) fee,MONTH(create_time) m, YEAR(create_time) y FROM interest WHERE user_id = ? AND create_time BETWEEN ? AND ?`, uid, start, end)
	res2, _ := SqlOperators(`SELECT IFNULL(fee,0) fee,MONTH(create_time) m, YEAR(create_time) y FROM commission WHERE ib_id = ? AND create_time BETWEEN ? AND ?`, uid, start, end)
	res = append(res, res1...)
	res = append(res, res2...)

	for _, v := range res {
		if time.Now().Year() == ToInt(PtoString(v, "y")) {
			for i := 1; i <= 12; i++ {
				if ToInt(PtoString(v, "m")) == i {
					m[i] += ToFloat64(PtoString(v, "fee"))
				}
			}
		}
		if time.Now().Year()-1 == ToInt(PtoString(v, "y")) {
			for i := 1; i <= 12; i++ {
				if ToInt(PtoString(v, "m")) == i {
					m1[i] += ToFloat64(PtoString(v, "fee"))
				}
			}
		}
		if time.Now().Year()-2 == ToInt(PtoString(v, "y")) {
			for i := 1; i <= 12; i++ {
				if ToInt(PtoString(v, "m")) == i {
					m2[i] += ToFloat64(PtoString(v, "fee"))
				}
			}
		}
	}

	return m, m1, m2
}

func ConvertMonthlyIncome1(res map[int]float64, num int) []float64 {
	var arr []float64
	flag := false

	for i := 1; i <= num; i++ {
		for k, v := range res {
			if i == k {
				flag = true
				arr = append(arr, ToFloat64(fmt.Sprintf("%.2f", v)))
				continue
			}
		}
		if !flag {
			arr = append(arr, 0)
		}
		flag = false
	}

	return arr
}

func ConvertMonthlyIncome(res []interface{}, num int) []float64 {
	var arr []float64
	flag := false

	for i := 1; i <= num; i++ {
		for k, v := range res {
			if i == ToInt(PtoString(res[k], "months")) {
				flag = true
				arr = append(arr, ToFloat64(PtoString(v, "fee")))
				continue
			}
		}
		if !flag {
			arr = append(arr, 0)
		}
		flag = false
	}

	return arr
}

func CreateCaptcha(mail, code string) {
	c := Captcha{
		Address:    mail,
		Code:       code,
		CreateTime: FormatNow(),
		CreateAt:   time.Now().Unix(),
	}

	db.Debug().Create(&c)
}

func GetUserMoreById(uid int) UserMore {
	var um UserMore
	db.Debug().Where("user_id = ?", uid).First(&um)

	return um
}

func MonthExchangeDay(res []interface{}) interface{} {
	mSlice := make([]map[string]int, 0)
	month := time.Now().Month()
	day := 30

	if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		day = 31
	}

	flag := false
	for i := 1; i <= day; i++ {
		m := make(map[string]int)
		for _, v := range res {
			if i == ToInt(PtoString(v, "d")) {
				m[fmt.Sprintf(`%s-%s`, fmt.Sprintf(`%02d`, month), fmt.Sprintf(`%02d`, i))] = ToInt(PtoString(v, "num"))
				mSlice = append(mSlice, m)
				flag = true
				break
			}
		}

		if !flag {
			m[fmt.Sprintf(`%s-%s`, fmt.Sprintf(`%02d`, month), fmt.Sprintf(`%02d`, i))] = 0
			mSlice = append(mSlice, m)
		}

		flag = false
	}

	return mSlice
}

func GetUserAudit(uid int) UserAuditLog {
	var a UserAuditLog

	db.Debug().Where("user_id = ? and old = 0", uid).Order("create_time DESC").First(&a)

	return a
}

func GetUserVip(uid int) UserVip {
	var uv UserVip
	db.Debug().Where("user_id = ?", uid).First(&uv)

	return uv
}

func CreateCapture(address, code, createTime string, createAt int64, cType int) {
	c := Captcha{
		Address:    address,
		Code:       code,
		CreateTime: createTime,
		CreateAt:   createAt,
		Type:       cType,
	}

	db.Debug().Create(&c)
}

func GetFilterCriteria(where, name string, grade int, state int, email string, vipGrade int, inviter, groupName string) string {
	if name != "" {
		if !strings.Contains(name, "'") {
			where += fmt.Sprintf(" AND REPLACE(u.true_name,' ','') = '%v'", strings.ReplaceAll(name, " ", ""))
		}
	}

	if grade != 0 {
		switch grade {
		case 1:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅰ")
		case 2:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅱ")
		case 3:
			where += fmt.Sprintf(` AND u.user_type = '%v'`, "Level Ⅲ")
		default:

		}
	}

	if state != 0 {
		switch state {
		case 1:
			where += fmt.Sprintf(` AND u.auth_status = 0`)
		case 2:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn != 0`)
		case 3:
			where += fmt.Sprintf(` AND u.auth_status = 1 AND p1.walletIn IS NULL`)
		case 4:
			where += fmt.Sprintf(` AND u.status = -1`)
		case 5:
			where += fmt.Sprintf(` AND u.status = 1`)
		}
	}

	if email != "" {
		if !strings.Contains(name, "'") {
			where += fmt.Sprintf(` AND u.email = '%v'`, email)
		}
	}

	if vipGrade != 0 {
		where += fmt.Sprintf(` AND uv.grade = %v`, vipGrade)
	}

	if inviter != "" {
		if !strings.Contains(name, "'") {
			where += fmt.Sprintf(` AND u1.true_name = '%v'`, inviter)
		}
	}

	if groupName != "" {
		if !strings.Contains(name, "'") {
			where += fmt.Sprintf(` AND rc.group_name = '%v'`, groupName)
		}
	}

	return where
}

func ExportList(sql string, userType int) []interface{} {
	rows, err := db.Debug().Raw(sql).Rows()
	defer rows.Close()
	if err != nil {
		log.Println("abc UserList ", err)
	}

	result := HandleRawSQL(rows)

	for _, v := range result {
		if PtoString(v, "mobile") == "" {
			v.(map[string]interface{})["mobile"] = PtoString(v, "old_mobile")
		}

		if PtoString(v, "phonectcode") == "" {
			v.(map[string]interface{})["phonectcode"] = PtoString(v, "old_phonectcode")
		}
	}

	//0伙伴管理  1业务员管理  2客户管理
	if userType == 1 {
		for _, v := range result {
			if ToInt(PtoString(v, "status")) == -1 {
				v.(map[string]interface{})["user_status"] = 4
				continue
			}

			if ToInt(PtoString(v, "status")) == 1 {
				v.(map[string]interface{})["user_status"] = 5
				continue
			}
		}
	} else {
		for _, v := range result {
			if ToInt(PtoString(v, "status")) == -1 {
				v.(map[string]interface{})["user_status"] = 4
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 1 && ToFloat64(PtoString(v, "walletIn")) != 0 {
				v.(map[string]interface{})["user_status"] = 2
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 1 {
				v.(map[string]interface{})["user_status"] = 3
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 0 {
				v.(map[string]interface{})["user_status"] = 1
				continue
			}
		}
	}

	return result
}

func UserList(sql, totalSql, countSql string, userType int) ([]interface{}, interface{}, int64) {

	rows, err := db.Debug().Raw(sql).Rows()
	defer rows.Close()
	if err != nil {
		log.Println("abc UserList ", err)
	}

	result := HandleRawSQL(rows)

	for _, v := range result {
		if PtoString(v, "mobile") == "" {
			v.(map[string]interface{})["mobile"] = PtoString(v, "old_mobile")
		}

		if PtoString(v, "phonectcode") == "" {
			v.(map[string]interface{})["phonectcode"] = PtoString(v, "old_phonectcode")
		}
	}

	//0伙伴管理  1业务员管理  2客户管理
	if userType == 1 {
		for _, v := range result {
			if ToInt(PtoString(v, "status")) == -1 {
				v.(map[string]interface{})["user_status"] = 4
				continue
			}

			if ToInt(PtoString(v, "status")) == 1 {
				v.(map[string]interface{})["user_status"] = 5
				continue
			}
		}
	} else {
		for _, v := range result {
			if ToInt(PtoString(v, "status")) == -1 {
				v.(map[string]interface{})["user_status"] = 4
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 1 && ToFloat64(PtoString(v, "walletIn")) != 0 {
				v.(map[string]interface{})["user_status"] = 2
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 1 {
				v.(map[string]interface{})["user_status"] = 3
				continue
			}

			if ToInt(PtoString(v, "auth_status")) == 0 || ToInt(PtoString(v, "auth_status")) == -1 {
				v.(map[string]interface{})["user_status"] = 1
				continue
			}
		}
	}

	rows1, err1 := db.Debug().Raw(totalSql).Rows()
	defer rows1.Close()

	if err1 != nil {
		log.Println("abc UserList 1", err1)
	}

	result1 := HandleRawSQL(rows1)

	var count int64
	rows2, err2 := db.Debug().Raw(countSql).Rows()
	if err2 != nil {
		log.Println("abc UserList 2", err2)
	}

	result2 := HandleRawSQL(rows2)

	if len(result2) > 0 {
		count = ToInt64(PtoString(result2[0], "count"))
	}

	return result, result1[0], count
}

func PartnerList(uid, page, size int, where, startTime, endTime, order string, userType int) ([]interface{}, interface{}, int64) {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode,u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.mobile, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user_more um
                                  ON u.id = um.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v ORDER BY %v LIMIT %v,%v`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where, order, (page-1)*size, size)
	totalSql := fmt.Sprintf(`SELECT SUM(temp.balance) balance_total, SUM(temp.equity) equity_total, SUM(temp.forex) forex_total, SUM(temp.metal) metal_total, SUM(temp.stockCommission) stock_commission_total, SUM(temp.silver) silver_total, SUM(temp.dma) dma_total, SUM(temp.deposit) deposit_total, SUM(temp.withdraw) withdraw_total, SUM(temp.walletIn) walletIn_total, SUM(temp.walletOut) walletOut_total, SUM(temp.wallet_balance) wallet_balance_total 
									   FROM (SELECT u.id, u.user_type, u.email, u.true_name, u.auth_status, u.mobile, rc.group_name, uv.grade, u1.true_name Inviter, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v ORDER BY %v) temp`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where, order)
	countSql := fmt.Sprintf(`SELECT COUNT(u.id) count FROM user u 
	   								  LEFT JOIN rebate_config rc
									  ON u.rebate_id = rc.id
									  LEFT JOIN user_vip uv
									  ON u.id = uv.user_id
									  LEFT JOIN user u1
								      ON u.parent_id = u1.id
									  LEFT JOIN
									 (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  	 ON u.id = p1.user_path
									 WHERE %v`, uid, startTime, endTime, len(u.Path), u.Path, uid, where)

	return UserList(sql, totalSql, countSql, userType)

}

func ExportPartner(uid int, where, startTime, endTime string) []interface{} {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode,u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.mobile, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user_more um
                                  ON u.id = um.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where)

	return ExportList(sql, 0)
}

func SalesmanList(uid, page, size int, where, startTime, endTime, order string, userType int) ([]interface{}, interface{}, int64) {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.id, u.user_type, u.email, u.true_name, u.auth_status, u.mobile, rc.group_name, uv.grade, u1.true_name Inviter, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user_more um
                                  ON u.id = um.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma 
                                   FROM 
                                   (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v ORDER BY %v LIMIT %v,%v`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where, order, (page-1)*size, size)
	totalSql := fmt.Sprintf(`SELECT SUM(temp.balance) balance_total, SUM(temp.equity) equity_total, SUM(temp.forex) forex_total, SUM(temp.metal) metal_total, SUM(temp.stockCommission) stock_commission_total, SUM(temp.silver) silver_total, SUM(temp.dma) dma_total, SUM(temp.deposit) deposit_total, SUM(temp.withdraw) withdraw_total, SUM(temp.walletIn) walletIn_total, SUM(temp.walletOut) walletOut_total, SUM(temp.wallet_balance) wallet_balance_total 
									   FROM (SELECT u.id, u.user_type, u.email, u.true_name, u.auth_status, u.mobile, rc.group_name, uv.grade, u1.true_name Inviter, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v ORDER BY %v) temp`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where, order)
	countSql := fmt.Sprintf(`SELECT COUNT(u.id) count FROM user u 
	   								  LEFT JOIN rebate_config rc
									  ON u.rebate_id = rc.id
									  LEFT JOIN user_vip uv
									  ON u.id = uv.user_id
									  LEFT JOIN user u1
								      ON u.parent_id = u1.id
									  LEFT JOIN
									  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  	  ON u.id = p1.user_path
									  WHERE %v`, uid, startTime, endTime, len(u.Path), u.Path, uid, where)

	return UserList(sql, totalSql, countSql, userType)
}

func ExportSalesman(uid int, where, startTime, endTime string) []interface{} {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.id, u.user_type, u.email, u.true_name, u.auth_status, u.mobile, rc.group_name, uv.grade, u1.true_name Inviter, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN rebate_config rc
                                  ON u.rebate_id = rc.id
                                  LEFT JOIN user_vip uv
                                  ON u.id = uv.user_id
                                  LEFT JOIN user_more um
                                  ON u.id = um.user_id
                                  LEFT JOIN user u1
                                  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a 
								  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
								  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma 
                                   FROM 
                                   (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
								  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
                                  ON u.id = p1.user_path
                                  WHERE %v`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where)

	return ExportList(sql, 1)
}

func CustomList(uid, page, size int, where, startTime, endTime, order string, userType int, otherWhere string) ([]interface{}, interface{}, int64) {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	//u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
	  							  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal,
	                             IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
	 							  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
	  							  LEFT JOIN rebate_config rc
								  ON u.rebate_id = rc.id
								  LEFT JOIN user_vip uv
								  ON u.id = uv.user_id
								  LEFT JOIN user u1
								  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
									LEFT JOIN (SELECT u.wallet_balance, u.id FROM user u WHERE %v) u2
									ON u2.id = u.id
	 							    LEFT JOIN user_more um
									ON u.id = um.user_id
									 LEFT JOIN 
									 (SELECT IFNULL(SUM(balance),0) balance, IFNULL(SUM(equity),0) equity, user_id FROM account WHERE login > 0 GROUP BY user_id) a2
							     ON u.id = a2.user_id
									 LEFT JOIN
								  (SELECT com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume, sum(fee) fee, symbol_type FROM commission  WHERE commission_type = 0 and close_time between '%v' and '%v' group by uid,symbol) com GROUP BY com.uid) c1
									 ON u.Id = c1.uid
									LEFT JOIN
								  (select SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1) as user_path, SUM(IF(profit > 0,profit,0)) deposit, SUM(IF(profit < 0,profit,0)) withdraw from orders where cmd = 6 and close_time between '%v' and '%v' group by SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1)) ord
								  ON u.Id = ord.user_path
									LEFT JOIN
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id 								 from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v ORDER BY %v LIMIT %v,%v`, otherWhere, startTime, endTime, startTime, endTime, startTime, endTime, where, order, (page-1)*size, size)
	totalSql := fmt.Sprintf(`SELECT SUM(temp.balance) balance_total, SUM(temp.equity) equity_total, SUM(temp.forex) forex_total, SUM(temp.metal) metal_total, SUM(temp.stockCommission) stock_commission_total, SUM(temp.silver) silver_total, SUM(temp.dma) dma_total, SUM(temp.deposit) deposit_total, SUM(temp.withdraw) withdraw_total, SUM(temp.walletIn) walletIn_total, SUM(temp.walletOut) walletOut_total, SUM(temp.wallet_balance) wallet_balance_total 
									FROM (SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
	  							  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal,
	                             IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
	 							  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
	  							  LEFT JOIN rebate_config rc
								  ON u.rebate_id = rc.id
								  LEFT JOIN user_vip uv
								  ON u.id = uv.user_id
								  LEFT JOIN user u1
								  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
									LEFT JOIN (SELECT u.wallet_balance, u.id FROM user u WHERE %v) u2
									ON u2.id = u.id
	 							    LEFT JOIN user_more um
									ON u.id = um.user_id
									 LEFT JOIN 
									 (SELECT IFNULL(SUM(balance),0) balance, IFNULL(SUM(equity),0) equity, user_id FROM account WHERE login > 0 GROUP BY user_id) a2
							     ON u.id = a2.user_id
									 LEFT JOIN
								  (SELECT com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume, sum(fee) fee, symbol_type FROM commission  WHERE commission_type = 0 and close_time between '%v' and '%v' group by uid,symbol) com GROUP BY com.uid) c1
									 ON u.Id = c1.uid
									LEFT JOIN
								  (select SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1) as user_path, SUM(IF(profit > 0,profit,0)) deposit, SUM(IF(profit < 0,profit,0)) withdraw from orders where cmd = 6 and close_time between '%v' and '%v' group by SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1)) ord
								  ON u.Id = ord.user_path
									LEFT JOIN
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v) temp`, otherWhere, startTime, endTime, startTime, endTime, startTime, endTime, where)
	countSql := fmt.Sprintf(`SELECT COUNT(u.id) count FROM user u 
	   								  LEFT JOIN rebate_config rc
									  ON u.rebate_id = rc.id
									  LEFT JOIN user_vip uv
									  ON u.id = uv.user_id
									  LEFT JOIN user u1
								      ON u.parent_id = u1.id
								      LEFT JOIN 
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v`, startTime, endTime, where)

	return UserList(sql, totalSql, countSql, userType)
}

func ExportCustom(where, startTime, endTime string, otherWhere string) []interface{} {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	//u := GetUserById(uid)
	sql := fmt.Sprintf(`SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
	  							  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal,
	                             IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
	 							  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
	  							  LEFT JOIN rebate_config rc
								  ON u.rebate_id = rc.id
								  LEFT JOIN user_vip uv
								  ON u.id = uv.user_id
								  LEFT JOIN user u1
								  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
									LEFT JOIN (SELECT u.wallet_balance, u.id FROM user u WHERE %v) u2
									ON u2.id = u.id
	 							    LEFT JOIN user_more um
									ON u.id = um.user_id
									 LEFT JOIN 
									 (SELECT IFNULL(SUM(balance),0) balance, IFNULL(SUM(equity),0) equity, user_id FROM account WHERE login > 0 GROUP BY user_id) a2
							     ON u.id = a2.user_id
									 LEFT JOIN
								  (SELECT com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume, sum(fee) fee, symbol_type FROM commission  WHERE commission_type = 0 and close_time between '%v' and '%v' group by uid,symbol) com GROUP BY com.uid) c1
									 ON u.Id = c1.uid
									LEFT JOIN
								  (select SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1) as user_path, SUM(IF(profit > 0,profit,0)) deposit, SUM(IF(profit < 0,profit,0)) withdraw from orders where cmd = 6 and close_time between '%v' and '%v' group by SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1)) ord
								  ON u.Id = ord.user_path
									LEFT JOIN
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id 								 from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v`, otherWhere, startTime, endTime, startTime, endTime, startTime, endTime, where)

	return ExportList(sql, 2)
}

func SortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance int) (string, bool) {
	order := ""
	if deposit != 0 {
		switch deposit {
		case 1:
			order = "ord.deposit ASC"
		case 2:
			order = "ord.deposit DESC"
		}
		return order, true
	}
	if withdraw != 0 {
		switch withdraw {
		case 1:
			order = "ord.withdraw ASC"
		case 2:
			order = "ord.withdraw DESC"
		}
		return order, true
	}
	if walletIn != 0 {
		switch walletIn {
		case 1:
			order = "p1.walletIn ASC"
		case 2:
			order = "p1.walletIn DESC"
		}
		return order, true
	}
	if walletOut != 0 {
		switch walletOut {
		case 1:
			order = "p1.walletOut ASC"
		case 2:
			order = "p1.walletOut DESC"
		}
		return order, true
	}
	if walletOut != 0 {
		switch walletOut {
		case 1:
			order = "p1.walletOut ASC"
		case 2:
			order = "p1.walletOut DESC"
		}
		return order, true
	}
	if balance != 0 {
		switch balance {
		case 1:
			order = "a2.balance ASC"
		case 2:
			order = "a2.balance DESC"
		}
		return order, true
	}
	if equity != 0 {
		switch equity {
		case 1:
			order = "a2.equity ASC"
		case 2:
			order = "a2.equity DESC"
		}
		return order, true
	}
	if forex != 0 {
		switch forex {
		case 1:
			order = "c1.forex ASC"
		case 2:
			order = "c1.forex DESC"
		}
		return order, true
	}
	if metal != 0 {
		switch metal {
		case 1:
			order = "c1.metal ASC"
		case 2:
			order = "c1.metal DESC"
		}
		return order, true
	}
	if stockCommission != 0 {
		switch stockCommission {
		case 1:
			order = "c1.stockCommission ASC"
		case 2:
			order = "c1.stockCommission DESC"
		}
		return order, true
	}
	if silver != 0 {
		switch silver {
		case 1:
			order = "c1.silver ASC"
		case 2:
			order = "c1.silver DESC"
		}
		return order, true
	}
	if dma != 0 {
		switch dma {
		case 1:
			order = "c1.dma ASC"
		case 2:
			order = "c1.dma DESC"
		}
		return order, true
	}
	if walletBalance != 0 {
		switch walletBalance {
		case 1:
			order = "u2.wallet_balance ASC"
		case 2:
			order = "u2.wallet_balance DESC"
		}
		return order, true
	}

	return "u.id DESC", false
}

func NewSortCriteria(deposit, withdraw, walletIn, walletOut, equity, balance, forex, metal, stockCommission, silver, dma, walletBalance int) (string, bool) {
	order := ""
	if deposit != 0 {
		switch deposit {
		case 1:
			order = "deposit ASC"
		case 2:
			order = "deposit DESC"
		}
		return order, true
	}
	if withdraw != 0 {
		switch withdraw {
		case 1:
			order = "withdraw ASC"
		case 2:
			order = "withdraw DESC"
		}
		return order, true
	}
	if walletIn != 0 {
		switch walletIn {
		case 1:
			order = "walletIn ASC"
		case 2:
			order = "walletIn DESC"
		}
		return order, true
	}
	if walletOut != 0 {
		switch walletOut {
		case 1:
			order = "walletOut ASC"
		case 2:
			order = "walletOut DESC"
		}
		return order, true
	}
	if walletOut != 0 {
		switch walletOut {
		case 1:
			order = "walletOut ASC"
		case 2:
			order = "walletOut DESC"
		}
		return order, true
	}
	if balance != 0 {
		switch balance {
		case 1:
			order = "balance ASC"
		case 2:
			order = "balance DESC"
		}
		return order, true
	}
	if equity != 0 {
		switch equity {
		case 1:
			order = "equity ASC"
		case 2:
			order = "equity DESC"
		}
		return order, true
	}
	if forex != 0 {
		switch forex {
		case 1:
			order = "forex ASC"
		case 2:
			order = "forex DESC"
		}
		return order, true
	}
	if metal != 0 {
		switch metal {
		case 1:
			order = "metal ASC"
		case 2:
			order = "metal DESC"
		}
		return order, true
	}
	if stockCommission != 0 {
		switch stockCommission {
		case 1:
			order = "stockCommission ASC"
		case 2:
			order = "stockCommission DESC"
		}
		return order, true
	}
	if silver != 0 {
		switch silver {
		case 1:
			order = "silver ASC"
		case 2:
			order = "silver DESC"
		}
		return order, true
	}
	if dma != 0 {
		switch dma {
		case 1:
			order = "dma ASC"
		case 2:
			order = "dma DESC"
		}
		return order, true
	}
	if walletBalance != 0 {
		switch walletBalance {
		case 1:
			order = "wallet_balance ASC"
		case 2:
			order = "wallet_balance DESC"
		}
		return order, true
	}

	return "id DESC", false
}

func UpdateUserParentId(tx *gorm.DB, userId int, path string) error {
	if err := tx.Debug().Model(&User{}).Where("id != ? and path like ?", userId, path+"%s").Updates(map[string]interface{}{
		"parent_id": userId,
		"sales_id":  0,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func ReplaceUserPath(tx *gorm.DB, newPath, path, language string) (int, string) {

	err := tx.Debug().Exec(fmt.Sprintf(`update user set path = REPLACE(path,'%s','%s') where path like '%s%%'`, path, newPath, path)).Error

	if err != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	err1 := tx.Debug().Exec(fmt.Sprintf(`update orders set user_id = REPLACE(user_id,'%s','%s') where user_id like '%s%%'`, path, newPath, path)).Error

	if err1 != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	err2 := tx.Debug().Exec(fmt.Sprintf(`update account set user_path = REPLACE(user_path,'%s','%s') where user_path like '%s%%'`, path, newPath, path)).Error

	if err2 != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	err3 := tx.Debug().Exec(fmt.Sprintf(`update commission set user_path = REPLACE(user_path,'%s','%s') where user_path like '%s%%'`, path, newPath, path)).Error

	if err3 != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	err4 := tx.Debug().Exec(fmt.Sprintf(`update payment set user_path = REPLACE(user_path,'%s','%s') where user_path like '%s%%'`, path, newPath, path)).Error

	if err4 != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	err5 := tx.Debug().Exec(fmt.Sprintf(`update commission_set_custom set user_path = REPLACE(user_path,'%s','%s') where user_path like '%s%%'`, path, newPath, path)).Error

	if err5 != nil {
		log.Println("abc ReplaceUserPath ", err)
		tx.Rollback()
		return 0, golbal.Wrong[language][10048]
	}

	return 1, ""
}

func EditUserPth(tx *gorm.DB, uid int, path string) error {
	if err := tx.Debug().Model(&User{}).Where("id = ? ", uid).Updates(map[string]interface{}{
		"path": path + ToString(uid) + ",",
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func UnregisteredList(uid, page, size int, where string) ([]interface{}, int64) {
	var count int64
	res, _ := SqlOperators(fmt.Sprintf(`SELECT id, create_time, email FROM customer WHERE user_id = %v AND status = 0 %v LIMIT %v,%v`, uid, where, (page-1)*size, size))
	db.Debug().Table("customer").Where("user_id = ? AND `status` = 0", uid).Count(&count)

	return res, count

}

func GetMyData(uid int) interface{} {
	res, _ := SqlOperator(`SELECT u.id, u.true_name, uv.grade, u.wallet_balance FROM user u
					 LEFT JOIN user_vip uv 
					 ON u.id = uv.user_id
					 WHERE u.id = ?`, uid)

	return res
}

func SwitchStatus(status, authStatus, count int) int {
	if status == -1 {
		return 4
	}
	if authStatus == 1 && count >= 1 {
		return 2
	}
	if authStatus == 1 {
		return 3
	}
	if authStatus == 0 {
		return 1
	}
	if status == 1 {
		return 5
	}
	return 0
}

func SaveUnRegister(email string, uid int) {
	c := Customer{
		UserId:     uid,
		Email:      email,
		CreateTime: FormatNow(),
	}

	db.Debug().Create(&c)
}

func VerifyPasswordFormat(password string, cType int) bool {
	num := 0
	arr := []string{`[0-9]{1}`, `[a-z]{1}`, `[A-Z]{1}`}

	for _, v := range arr {
		match, _ := regexp.MatchString(v, password)
		if match {
			num++
		}
	}

	if cType == 0 {
		if num < 3 || len(password) < 8 {
			return false
		}
	} else {
		flag, _ := regexp.MatchString(`[.@$!%*#_~?&^]+`, password)
		if num < 2 || flag == true || len(password) < 5 {
			return false
		}

	}

	return true
}

func ClearUserProtocol(tx *gorm.DB, userid int) error {
	if err := tx.Debug().Model(&UserInfo{}).Where("user_id = ?", userid).Updates(map[string]interface{}{
		"agreement":     "",
		"agreement_fee": "",
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func ModifyUserParentID(tx *gorm.DB, userId, someId, ParentId, SalesId, cate int, userType string) error {
	if err := tx.Debug().Model(&User{}).Where("id = ?", userId).Updates(map[string]interface{}{
		"some_id":      someId,
		"parent_id":    ParentId,
		"sales_id":     SalesId,
		"user_type":    userType,
		"rebate_cate":  cate,
		"rebate_multi": 1,
		"ib_no":        "",
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetIdFile(uid int) UserFile {
	var f UserFile
	db.Debug().Where("user_id = ? and front = 1", uid).First(&f)

	return f
}

func GetSystemConfiguration() []interface{} {
	res, _ := SqlOperators(`select name, content from config2`)

	return res
}

func GetUserIdNumber(identity string) bool {
	var ui UserInfo
	db.Debug().Where("identity = ?", identity).First(&ui)

	if ui.UserId == 0 {
		return false
	}

	u := GetUserById(ui.UserId)

	if u.UserType == "sales" && u.SalesType == "admin" {
		return true
	}

	return false
}

func PhoneIsExit(phone, phonectcode string) bool {
	var u User
	db.Debug().Where("mobile = ? and phonectcode = ?", phone, phonectcode).First(&u)

	if u.Id == 0 {
		return true
	}

	return false
}

func DisableIDNumber(identity string) bool {
	var rSlice []RegionBlackHouse
	db.Debug().Find(&rSlice)

	for _, v := range rSlice {
		if strings.HasPrefix(identity, v.Code) {
			return false
		}
	}

	return true
}

func GetLiveChat() string {
	var l LiveChat
	db.Debug().Where("status = 1").Order("id DESC").First(&l)

	return l.Content
}

func DelFile(arr []string) {
	for _, v := range arr {
		p := strings.ReplaceAll(v, "https://xpub.cn-sh2.ufileos.com/", "")
		a := strings.Split(p, "/")
		RemoveFile("upload/" + a[len(a)-1])

		config := &ufsdk.Config{
			PublicKey:       "TOKEN_7d9baa9b-903d-4f60-b20a-05d869fa469e",
			PrivateKey:      "001d0145-2137-48ad-b87d-1d726ae76782",
			BucketName:      "xpub",
			FileHost:        "cn-sh2.ufileos.com",
			VerifyUploadMD5: false,
		}

		UsReq, err := ufsdk.NewFileRequest(config, nil)
		if err != nil {
			log.Println("abc DelFile 1", err)
		}

		err = UsReq.DeleteFile(p)
		if err != nil {
			log.Println(fmt.Sprintf("删除文件 %s 失败，错误信息为：%s", p, err.Error()))
		}
	}
}

func RemoveDuplicateData(arr []string, oldArr []string) []string {
	flag := false
	var newArr []string
	for _, v := range arr {
		for _, vv := range oldArr {
			if v == vv {
				flag = true
				break
			}
		}
		if !flag {
			newArr = append(newArr, v)
		}
		flag = false
	}

	return newArr
}

func ExportReport(uid int, startTime, endTime string) ([]Export, []CustomExport) {
	var exportList []Export
	u := GetUserById(uid)
	path := u.Path + "%"
	w1 := fmt.Sprintf("find_in_set(u.id,(select group_concat(id) ids from user where path like '%v' and user_type != 'user'))", path)
	w2 := fmt.Sprintf("find_in_set(u.id,(select group_concat(id) ids from user where path like '%v' and user_type = 'user'))", path)

	res1, res2 := StatisticalTransactionData(w1, w2, startTime, endTime)

	var users []User
	var temp []User
	db.Debug().Select([]string{"id, user_type, sales_type, parent_id, sales_id, path, true_name"}).Where("path LIKE ? AND user_type != 'user'", path).Find(&users)

	var myUser []User
	if u.UserType == "sales" {
		db.Debug().Select([]string{"id, user_type, sales_type, parent_id, sales_id, path, true_name"}).Where("parent_id = ? and sales_id = ? and user_type = 'user'", u.ParentId, u.Id).Find(&myUser)
	} else {
		db.Debug().Select([]string{"id, user_type, sales_type, parent_id, sales_id, path, true_name"}).Where("parent_id = ? and sales_id = ? and user_type = 'user'", u.Id, u.SalesId).Find(&myUser)
		var user1 []User
		db.Debug().Select([]string{"id, user_type, sales_type, parent_id, sales_id, path, true_name"}).Where("parent_id = ? and FIND_IN_SET(sales_id,(SELECT GROUP_CONCAT(id) FROM `user` WHERE parent_id = ? AND user_type = 'sales')) and user_type = 'user'", u.Id, u.Id).Find(&user1)
		myUser = append(myUser, user1...)
	}

	var salesUser []User
	if strings.Contains(u.UserType, "Level") {
		db.Debug().Select([]string{"id, user_type, sales_type, parent_id, sales_id, path, true_name"}).Where("parent_id = ? and sales_id = ? and user_type = 'sales'", u.Id, u.SalesId).Find(&salesUser)
	}

	var customIds []string
	var customList []CustomExport
	for _, v := range myUser {
		customIds = append(customIds, ToString(v.Id))
	}

	res3, _ := StatisticalTransactionData(fmt.Sprintf("FIND_IN_SET(u.id,'%v')", strings.Join(customIds, ",")), "", startTime, endTime)

	for _, v := range myUser {
		var custom CustomExport
		custom.Name = v.TrueName
		custom.Role = v.UserType
		for _, vv := range res3 {
			if v.Id == ToInt(PtoString(vv, "id")) {
				custom.My = vv
				break
			}
		}

		for _, vvv := range salesUser {
			if v.SalesId == vvv.Id {
				custom.SuperiorName = vvv.TrueName
				break
			}
		}

		if custom.SuperiorName == "" {
			custom.SuperiorName = u.TrueName
		}

		customList = append(customList, custom)
	}

	if strings.Contains(u.UserType, "Level") {
		a := SalesGroup([]User{}).AllSalesGroups(users)
		temp = a
	} else {
		Ib := User{}
		Ib.Id = u.ParentId
		users = append([]User{Ib}, users...)
		b := SalesGroup([]User{}).AllSalesGroups(users)
		b = b[1:]
		temp = b
	}

	for _, v := range temp {
		var export Export
		export.Name = v.TrueName
		export.Role = v.UserType + v.SalesType
		export.ParentId = v.ParentId
		export.Path = v.Path
		for _, vv := range res1 {
			if v.ParentId == ToInt(PtoString(vv, "parent_id")) && v.SalesId == ToInt(PtoString(vv, "sales_id")) {
				export.SuperiorName = PtoString(vv, "Inviter")
				export.My = vv
			}
		}
		for _, vvv := range res2 {
			if v.ParentId == ToInt(PtoString(vvv, "parent_id")) && v.SalesId == ToInt(PtoString(vvv, "sales_id")) {
				export.SuperiorName = PtoString(vvv, "Inviter")
				export.Customer = vvv
			}
		}

		exportList = append(exportList, export)
	}

	return exportList, customList
}

func StatisticalTransactionData(where, otherWhere, startTime, endTime string) ([]interface{}, []interface{}) {
	var res, res1 []interface{}
	//统计自己的
	if where != "" {
		sql1 := fmt.Sprintf(`SELECT u.id, u.parent_id, u.sales_id, IFNULL(u2.wallet_balance,0) wallet_balance, u1.true_name Inviter,
	  							  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal,
	                             IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
	 							  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
	  							  LEFT JOIN rebate_config rc
								  ON u.rebate_id = rc.id
								  LEFT JOIN user_vip uv
								  ON u.id = uv.user_id
								  LEFT JOIN user u1
								  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
									LEFT JOIN (SELECT u.wallet_balance, u.id FROM user u WHERE %v) u2
									ON u2.id = u.id
	 							    LEFT JOIN user_more um
									ON u.id = um.user_id
									 LEFT JOIN 
									 (SELECT IFNULL(SUM(balance),0) balance, IFNULL(SUM(equity),0) equity, user_id FROM account WHERE login > 0 GROUP BY user_id) a2
							     ON u.id = a2.user_id
									 LEFT JOIN
								  (SELECT com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume, sum(fee) fee, symbol_type FROM commission  WHERE commission_type = 0 and close_time between '%v' and '%v' group by uid,symbol) com GROUP BY com.uid) c1
									 ON u.Id = c1.uid
									LEFT JOIN
								  (select SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1) as user_path, SUM(IF(profit > 0,profit,0)) deposit, SUM(IF(profit < 0,profit,0)) withdraw from orders where cmd = 6 and close_time between '%v' and '%v' group by SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1)) ord
								  ON u.Id = ord.user_path
									LEFT JOIN
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id 								 from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v`, where, startTime, endTime, startTime, endTime, startTime, endTime, where)

		res, _ = SqlOperators(sql1)
	}

	//统计直客的
	if otherWhere != "" {
		sql2 := fmt.Sprintf(`SELECT temp.parent_id, temp.sales_id, SUM(temp.balance) balance, SUM(temp.equity) equity, SUM(temp.forex) forex, SUM(temp.metal) metal, SUM(temp.stockCommission) stock_commission, SUM(temp.silver) silver, SUM(temp.dma) dma, SUM(temp.deposit) deposit, SUM(temp.withdraw) withdraw, SUM(temp.walletIn) walletIn, SUM(temp.walletOut) walletOut, SUM(temp.wallet_balance) wallet_balance, temp.Inviter, temp.id
									FROM (SELECT um.mobile old_mobile, um.phonectcode old_phonectcode, u.parent_id, u.sales_id, u.id, u.user_type, u.email, u.true_name, u.auth_status, rc.group_name, uv.grade, u1.true_name Inviter, u.mobile, IFNULL(u2.wallet_balance,0) wallet_balance, u.phonectcode, u.status, u.create_time,
	  							  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal,
	                             IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
	 							  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
	  							  LEFT JOIN rebate_config rc
								  ON u.rebate_id = rc.id
								  LEFT JOIN user_vip uv
								  ON u.id = uv.user_id
								  LEFT JOIN user u1
								  ON SUBSTRING_INDEX(SUBSTRING_INDEX(u.path,',', -3),',',1) = u1.id
									LEFT JOIN (SELECT u.wallet_balance, u.id FROM user u) u2
									ON u2.id = u.id
	 							    LEFT JOIN user_more um
									ON u.id = um.user_id
									 LEFT JOIN 
									 (SELECT IFNULL(SUM(balance),0) balance, IFNULL(SUM(equity),0) equity, user_id FROM account WHERE login > 0 GROUP BY user_id) a2
							     ON u.id = a2.user_id
									 LEFT JOIN
								  (SELECT com.uid, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select uid, sum(volume) volume, sum(fee) fee, symbol_type FROM commission  WHERE commission_type = 0 and close_time between '%v' and '%v' group by uid,symbol) com GROUP BY com.uid) c1
									 ON u.Id = c1.uid
									LEFT JOIN
								  (select SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1) as user_path, SUM(IF(profit > 0,profit,0)) deposit, SUM(IF(profit < 0,profit,0)) withdraw from orders where cmd = 6 and close_time between '%v' and '%v' group by SUBSTRING_INDEX(SUBSTRING_INDEX(user_id,',', -2),',',1)) ord
								  ON u.Id = ord.user_path
									LEFT JOIN
									(SELECT SUM(IF(pay_fee+amount > 0,pay_fee+amount,0)) walletIn, SUM(IF(pay_fee+amount < 0,pay_fee+amount,0)) walletOut, user_id from payment
									where status = 1 and transfer_login = 0 and pay_time between '%v' and '%v' group by user_id) p1
									ON u.id = p1.user_id
									WHERE %v) temp group by temp.parent_id, temp.sales_id`, startTime, endTime, startTime, endTime, startTime, endTime, otherWhere)

		res1, _ = SqlOperators(sql2)
	}
	return res, res1
}

//func NewExportReport(uid int, where, startTime, endTime, order string, flag bool, name string, where1 string) []interface{} {
//	startTime = startTime + " 00:00:00"
//	endTime = endTime + " 23:59:59"
//	u := GetUserById(uid)
//	newPath := u.Path + "%"
//	//sql := fmt.Sprintf(`SELECT u.id, u.user_type, u.true_name, IFNULL(u2.wallet_balance,0) wallet_balance,
//	//                              IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
//	//                              IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
//	//                              LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE left(path,%v)='%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
//	//                              ON u.id = u2.user_id
//	//                              LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a
//	//							  WHERE a.login > 0 and left(a.user_path,%v)='%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
//	//							  GROUP BY a1.user_id) a2
//	//                              ON u.id = a2.user_id
//	//                              LEFT JOIN
//	//                              (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
//	//							  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and left(com1.user_path,%v)='%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
//	//                              ON u.Id = c1.user_path
//	//                              LEFT JOIN
//	//                              (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
//	//							  where o.cmd = 6 and o.close_time between '%v' and '%v' and left(o.user_id,%v)='%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
//	//                              ON u.Id = ord.user_path
//	//                              LEFT JOIN
//	//                              (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and left(p.user_path,%v)='%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
//	//                              ON u.id = p1.user_path
//	//                              WHERE %v ORDER BY %v`, len(u.Path), u.Path, uid, uid, uid, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, uid, startTime, endTime, len(u.Path), u.Path, uid, where, order)
//	//
//	//rows, err := db.Debug().Raw(sql).Rows()
//	//defer rows.Close()
//	//if err != nil {
//	//	log.Println("abc UserList ", err)
//	//}
//	//
//	//result := HandleRawSQL(rows)
//	//
//	////统计自己的数据
//	//
//	//res1, _ := SqlOperator(fmt.Sprintf(`SELECT u.id, u.user_type, u.true_name, IFNULL(u2.wallet_balance,0) wallet_balance,
//	//                              IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
//	//                              IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
//	//                              LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE id = %v) uu GROUP BY uu.user_id) u2
//	//                              ON u.id = u2.user_id
//	//                              LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, user_id FROM account a
//	//                                                              WHERE a.login > 0 and user_id = %v) a1
//	//                                                              GROUP BY a1.user_id) a2
//	//                              ON u.id = a2.user_id
//	//                              LEFT JOIN
//	//                              (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.uid as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
//	//                                                              WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and uid = %v group by com1.symbol_type) com GROUP BY com.user_path) c1
//	//                              ON u.Id = c1.user_path
//	//                              LEFT JOIN
//	//                              (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select substring_index(substring_index('%v',',',-2),',',1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
//	//                                                              where o.cmd = 6 and o.close_time between '%v' and '%v' and user_id ='%v') o1) ord
//	//                              ON u.Id = ord.user_path
//	//                              LEFT JOIN
//	//                              (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT user_id as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and user_id = %v group by p.type) pa) p1
//	//                              ON u.id = p1.user_path
//	//                              WHERE u.id = %v ORDER BY u.id DESC`, uid, uid, startTime, endTime, uid, u.Path, startTime, endTime, u.Path, startTime, endTime, uid, uid))
//	sql := fmt.Sprintf(`(SELECT u.email, u.id, u.user_type, u.true_name, IFNULL(u2.wallet_balance,0) wallet_balance, account.login,
//                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
//                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
//                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, SUBSTRING_INDEX(SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1),',',-1) user_id FROM user WHERE path like '%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id) u2
//                                  ON u.id = u2.user_id
//                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a
//								  WHERE a.login > 0 and a.user_path like '%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
//								  GROUP BY a1.user_id) a2
//                                  ON u.id = a2.user_id
//                                  LEFT JOIN
//                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
//								  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and com1.user_path like '%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path) c1
//                                  ON u.Id = c1.user_path
//                                  LEFT JOIN
//                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(o.user_id,',', FIND_IN_SET(%v,o.user_id) + 1),',',-1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
//								  where o.cmd = 6 and o.close_time between '%v' and '%v' and o.user_id like '%v' group by SUBSTRING_INDEX(o.user_id ,',', FIND_IN_SET(%v,o.user_id) + 1)) o1 GROUP BY o1.user_path) ord
//                                  ON u.Id = ord.user_path
//                                  LEFT JOIN
//                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and p.user_path like '%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path) p1
//                                  ON u.id = p1.user_path
//                                  LEFT JOIN
//							      (SELECT IFNULL(GROUP_CONCAT(a.login),'') login, a.user_id FROM user u
//								  LEFT JOIN account a
//								  ON u.id = a.user_id
//							      WHERE %v GROUP BY a.user_id) account
//							      ON u.id = account.user_id
//                                  WHERE %v)
//                                  UNION ALL
//                                  (SELECT u.email, u.id, u.user_type, u.true_name, IFNULL(u2.wallet_balance,0) wallet_balance, account.login,
//                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
//                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
//                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE id = %v) uu GROUP BY uu.user_id) u2
//                                  ON u.id = u2.user_id
//                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, user_id FROM account a
//                                                                  WHERE a.login > 0 and user_id = %v) a1
//                                                                  GROUP BY a1.user_id) a2
//                                  ON u.id = a2.user_id
//                                  LEFT JOIN
//                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.uid as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
//                                                                  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and ib_id = %v group by com1.symbol_type) com GROUP BY com.user_path) c1
//                                  ON u.Id = c1.user_path
//                                  LEFT JOIN
//                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select substring_index(substring_index('%v',',',-2),',',1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
//                                                                  where o.cmd = 6 and o.close_time between '%v' and '%v' and user_id ='%v') o1) ord
//                                  ON u.Id = ord.user_path
//                                  LEFT JOIN
//                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT user_id as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and user_id = %v group by p.type) pa) p1
//                                  ON u.id = p1.user_path
//								  LEFT JOIN
//								  (SELECT IFNULL(GROUP_CONCAT(a.login),'') login, a.user_id FROM user u
//								  LEFT JOIN account a
//								  ON u.id = a.user_id
//							      WHERE u.id = %v GROUP BY a.user_id) account
//							      ON u.id = account.user_id
//                                  WHERE u.id = %v)
//   								  ORDER BY %v`, uid, newPath, uid, uid, uid, newPath, uid, uid, startTime, endTime, newPath, uid, uid, startTime, endTime, newPath, uid, uid, startTime, endTime, newPath, uid, where1, where, uid, uid, startTime, endTime, uid, u.Path, startTime, endTime, u.Path, startTime, endTime, uid, uid, uid, order)
//
//	result, _ := SqlOperators(sql)
//
//	var temp []interface{}
//
//	if !flag {
//		for i := 0; i <= 5; i++ {
//			for _, v := range result {
//				if i == 0 && ToInt(PtoString(v, "id")) == uid {
//					temp = append(temp, v)
//				}
//
//				if i == 1 && PtoString(v, "user_type") == "sales" && ToInt(PtoString(v, "id")) != uid {
//					temp = append(temp, v)
//				}
//
//				if i == 2 && PtoString(v, "user_type") == "Level Ⅲ" && ToInt(PtoString(v, "id")) != uid {
//					temp = append(temp, v)
//				}
//
//				if i == 3 && PtoString(v, "user_type") == "Level Ⅱ" && ToInt(PtoString(v, "id")) != uid {
//					temp = append(temp, v)
//				}
//
//				if i == 4 && PtoString(v, "user_type") == "Level Ⅰ" && ToInt(PtoString(v, "id")) != uid {
//					temp = append(temp, v)
//				}
//
//				if i == 5 && PtoString(v, "user_type") == "user" && ToInt(PtoString(v, "id")) != uid {
//					temp = append(temp, v)
//				}
//			}
//		}
//	} else {
//		temp = result
//	}
//
//	for _, v := range temp {
//		email := strings.Split(PtoString(v, "email"), "@")
//
//		if len(email) == 2 {
//			if len(email[0]) <= 4 {
//				v.(map[string]interface{})["email"] = "****" + "@" + email[1]
//			} else {
//				v.(map[string]interface{})["email"] = email[0][0:len(email[0])-4] + "****" + "@" + email[1]
//			}
//		}
//	}
//	return temp
//}

type ExportData struct {
	Balance         string `json:"balance"`
	Deposit         string `json:"deposit"`
	Dma             string `json:"dma"`
	Email           string `json:"email"`
	Equity          string `json:"equity"`
	Forex           string `json:"forex"`
	Id              string `json:"id"`
	Login           string `json:"login"`
	Metal           string `json:"metal"`
	Silver          string `json:"silver"`
	StockCommission string `json:"stockCommission"`
	TrueName        string `json:"true_name"`
	UserType        string `json:"user_type"`
	WalletIn        string `json:"walletIn"`
	WalletOut       string `json:"walletOut"`
	WalletBalance   string `json:"wallet_balance"`
	Withdraw        string `json:"withdraw"`
}

func NewExportReport(uid int, where, startTime, endTime, order string, flag bool, name string, where1 string) []ExportData {
	startTime = startTime + " 00:00:00"
	endTime = endTime + " 23:59:59"
	u := GetUserById(uid)
	newPath := u.Path + "%"

	var eSlice []ExportData

	res, _ := SqlOperators(fmt.Sprintf(`SELECT u.email, u.id, u.user_type, u.true_name, account.login FROM user u
					  LEFT JOIN
					  (SELECT IFNULL(GROUP_CONCAT(a.login),'') login, a.user_id FROM user u
					  LEFT JOIN account a
					  ON u.id = a.user_id
					  WHERE %v GROUP BY a.user_id) account
					  ON u.id = account.user_id
					  WHERE %v ORDER BY u.id DESC`, where1, where))

	for _, v := range res {
		var e ExportData
		e.Email = PtoString(v, "email")
		e.Id = PtoString(v, "id")
		e.UserType = PtoString(v, "user_type")
		e.TrueName = PtoString(v, "true_name")
		e.Login = PtoString(v, "login")

		eSlice = append(eSlice, e)
	}

	res1, _ := SqlOperators(fmt.Sprintf(`SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, SUBSTRING_INDEX(SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1),',',-1) user_id FROM user WHERE path like '%v' AND id != %v group by SUBSTRING_INDEX(path ,',', FIND_IN_SET(%v,path) + 1)) uu GROUP BY uu.user_id`, uid, newPath, uid, uid))

	for _, v := range res1 {
		for kk, vv := range eSlice {
			if PtoString(v, "user_id") == vv.Id {
				vv.WalletBalance = PtoString(v, "wallet_balance")
				eSlice[kk] = vv
				break
			}
		}
	}

	res2, _ := SqlOperators(fmt.Sprintf(`SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, SUBSTRING_INDEX(SUBSTRING_INDEX(a.user_path,',', FIND_IN_SET(%v,a.user_path) + 1),',',-1) as user_id FROM account a
                                                                  WHERE a.login > 0 and a.user_path like '%v' group by SUBSTRING_INDEX(a.user_path ,',', FIND_IN_SET(%v,a.user_path) + 1)) a1
                                                                  GROUP BY a1.user_id`, uid, newPath, uid))

	for _, v := range res2 {
		for kk, vv := range eSlice {
			if PtoString(v, "user_id") == vv.Id {
				vv.Balance = PtoString(v, "balance")
				vv.Equity = PtoString(v, "equity")
				eSlice[kk] = vv
				break
			}
		}
	}

	res3, _ := SqlOperators(fmt.Sprintf(`SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(com1.user_path,',', FIND_IN_SET(%v,com1.user_path) + 1),',',-1) as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
                                                                  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and com1.user_path like '%v' group by SUBSTRING_INDEX(com1.user_path ,',', FIND_IN_SET(%v,com1.user_path) + 1),com1.symbol_type) com GROUP BY com.user_path`, uid, startTime, endTime, newPath, uid))

	for _, v := range res3 {
		for kk, vv := range eSlice {
			if PtoString(v, "user_path") == vv.Id {
				vv.Forex = PtoString(v, "forex")
				vv.Metal = PtoString(v, "metal")
				vv.StockCommission = PtoString(v, "stockCommission")
				vv.Silver = PtoString(v, "silver")
				vv.Dma = PtoString(v, "dma")
				eSlice[kk] = vv
				break
			}
		}
	}

	res4, _ := SqlOperators(fmt.Sprintf(`SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and p.user_path like '%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1), p.type) pa GROUP BY pa.user_path`, uid, startTime, endTime, newPath, uid))

	for _, v := range res4 {
		for kk, vv := range eSlice {
			if PtoString(v, "user_path") == vv.Id {
				vv.WalletIn = PtoString(v, "walletIn")
				vv.WalletOut = PtoString(v, "walletOut")

				eSlice[kk] = vv
				break
			}
		}
	}

	res5, _ := SqlOperators(fmt.Sprintf(`SELECT p1.user_path, SUM(IFNULL(p1.deposit,0)) deposit, SUM(IFNULL(p1.withdraw,0)) withdraw	FROM (select SUBSTRING_INDEX(SUBSTRING_INDEX(p.user_path,',', FIND_IN_SET(%v,p.user_path) + 1),',',-1) as user_path, SUM(IF(p.amount > 0,p.amount,0))*-1 withdraw, SUM(IF(p.amount < 0,p.amount,0))*-1 deposit from payment p
									  			where p.type = 'transfer' AND p.comment != 'salary' and p.status = 1 and p. pay_time between '%v' and '%v' and p.user_path like '%v' group by SUBSTRING_INDEX(p.user_path ,',', FIND_IN_SET(%v,p.user_path) + 1)) p1 GROUP BY p1.user_path`, uid, startTime, endTime, newPath, uid))

	for _, v := range res5 {
		for kk, vv := range eSlice {
			if PtoString(v, "user_path") == vv.Id {
				vv.Deposit = PtoString(v, "deposit")
				vv.Withdraw = PtoString(v, "withdraw")

				eSlice[kk] = vv
				break
			}
		}
	}
	var my ExportData
	db.Debug().Raw(fmt.Sprintf(`SELECT u.email, u.id, u.user_type, u.true_name, IFNULL(u2.wallet_balance,0) wallet_balance, account.login,
                                  IFNULL(a2.balance,0) balance, IFNULL(a2.equity,0) equity, IFNULL(c1.forex,0) forex, IFNULL(c1.metal,0) metal, IFNULL(c1.stockCommission,0) stockCommission,IFNULL(c1.silver,0) silver,IFNULL(c1.dma,0) dma,
                                  IFNULL(ord.deposit,0) deposit, IFNULL(ord.withdraw,0) withdraw, IFNULL(p1.walletIn,0) walletIn, IFNULL(p1.walletOut,0) walletOut FROM user u
                                  LEFT JOIN (SELECT SUM(uu.wallet_balance) wallet_balance, uu.user_id FROM (SELECT SUM(wallet_balance) wallet_balance, id user_id FROM user WHERE id = %v) uu GROUP BY uu.user_id) u2
                                  ON u.id = u2.user_id
                                  LEFT JOIN (SELECT a1.user_id, SUM(IFNULL(a1.balance,0)) balance, SUM(IFNULL(a1.equity,0)) equity FROM (SELECT IFNULL(SUM(a.balance),0) balance, IFNULL(SUM(a.equity),0) equity, user_id FROM account a
                                                                  WHERE a.login > 0 and user_id = %v) a1
                                                                  GROUP BY a1.user_id) a2
                                  ON u.id = a2.user_id
                                  LEFT JOIN
                                  (SELECT com.user_path, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (select com1.uid as user_path,sum(com1.volume) volume,sum(com1.fee) fee,com1.symbol_type FROM commission com1
                                                                  WHERE com1.commission_type = 0 and com1.close_time between '%v' and '%v' and ib_id = %v group by com1.symbol_type) com GROUP BY com.user_path) c1
                                  ON u.Id = c1.user_path
                                  LEFT JOIN
                                  (SELECT o1.user_path, SUM(IFNULL(o1.deposit,0)) deposit, SUM(IFNULL(o1.withdraw,0)) withdraw FROM (select substring_index(substring_index('%v',',',-2),',',1) as user_path, SUM(IF(o.profit > 0,o.profit,0)) deposit, SUM(IF(o.profit < 0,o.profit,0)) withdraw from orders o
                                                                  where o.cmd = 6 and o.comment = 'transfer' and o.close_time between '%v' and '%v' and user_id ='%v') o1) ord
                                  ON u.Id = ord.user_path
                                  LEFT JOIN
                                  (SELECT SUM(pa.walletIn) walletIn, SUM(pa.walletOut) walletOut, pa.user_path FROM (SELECT user_id as user_path, SUM(IF(amount > 0,amount+pay_fee,0)) walletIn, SUM(IF(amount < 0 ,amount+pay_fee,0)) walletOut FROM payment p where p.status = 1 and p.transfer_login = 0 and p.pay_time between '%v' and '%v' and user_id = %v group by p.type) pa) p1
                                  ON u.id = p1.user_path
                                                                  LEFT JOIN
                                                                  (SELECT IFNULL(GROUP_CONCAT(a.login),'') login, a.user_id FROM user u
                                                                  LEFT JOIN account a
                                                                  ON u.id = a.user_id
                                                              WHERE u.id = %v GROUP BY a.user_id) account
                                                              ON u.id = account.user_id
                                  WHERE u.id = %v`, uid, uid, startTime, endTime, uid, u.Path, startTime, endTime, u.Path, startTime, endTime, uid, uid, uid)).Scan(&my)

	eSlice = append(eSlice, my)
	var temp []ExportData
	if !flag {
		for i := 0; i <= 5; i++ {
			for _, v := range eSlice {
				if i == 0 && ToInt(v.Id) == uid {
					temp = append(temp, v)
				}

				if i == 1 && v.UserType == "sales" && ToInt(v.Id) != uid {
					temp = append(temp, v)
				}

				if i == 2 && v.UserType == "Level Ⅲ" && ToInt(v.Id) != uid {
					temp = append(temp, v)
				}

				if i == 3 && v.UserType == "Level Ⅱ" && ToInt(v.Id) != uid {
					temp = append(temp, v)
				}

				if i == 4 && v.UserType == "Level Ⅰ" && ToInt(v.Id) != uid {
					temp = append(temp, v)
				}

				if i == 5 && v.UserType == "user" && ToInt(v.Id) != uid {
					temp = append(temp, v)
				}
			}
		}
	} else {
		temp = eSlice
	}

	for _, v := range temp {
		email := strings.Split(v.Email, "@")

		if len(email) == 2 {
			if len(email[0]) <= 4 {
				v.Email = "****" + "@" + email[1]
			} else {
				v.Email = email[0][0:len(email[0])-4] + "****" + "@" + email[1]
			}
		}
	}
	return temp
}

func ProxyRelationshipList(where string, uid int) []interface{} {
	result, _ := SqlOperators(fmt.Sprintf(`select id, true_name, user_type from user u WHERE %v`, where))

	var temp []interface{}
	for i := 0; i <= 5; i++ {
		for _, v := range result {
			if i == 0 && ToInt(PtoString(v, "id")) == uid {
				temp = append(temp, v)
			}

			if i == 1 && PtoString(v, "user_type") == "sales" && ToInt(PtoString(v, "id")) != uid {
				temp = append(temp, v)
			}

			if i == 2 && PtoString(v, "user_type") == "Level Ⅲ" && ToInt(PtoString(v, "id")) != uid {
				temp = append(temp, v)
			}

			if i == 3 && PtoString(v, "user_type") == "Level Ⅱ" && ToInt(PtoString(v, "id")) != uid {
				temp = append(temp, v)
			}

			if i == 4 && PtoString(v, "user_type") == "Level Ⅰ" && ToInt(PtoString(v, "id")) != uid {
				temp = append(temp, v)
			}

			if i == 5 && PtoString(v, "user_type") == "user" && ToInt(PtoString(v, "id")) != uid {
				temp = append(temp, v)
			}
		}
	}

	return temp
}

func CreateUserAuditLog(ud UserAuditLog) {
	db.Debug().Create(&ud)
}

func EditLength()  {
	db.Debug().Exec(`SET SESSION group_concat_max_len=204800`)
}

func IsUserWhiteHouse(uid int) bool {
	var uwh UserWhiteHouse
	db.Debug().Where("user_id = ?", uid).First(&uwh)

	return uwh.Id == 0
}

func AddPaymentLog(pId int, comment string) {
	pl := &PaymentLog{
		PaymentId:  pId,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Comment:    comment,
	}

	db.Debug().Create(&pl)
}

func GetConfigOne(where string) Config {
	var c Config
	db.Debug().Where(where).First(&c)

	return c
}

func AddUserLog(uid int, cType, email, createTime, ip, content string) {
	var ul UserLog
	ul.UserId = uid
	ul.Type = cType
	ul.Email = email
	ul.CreateTime = createTime
	ul.Ip = ip
	ul.Content = content

	db.Debug().Create(&ul)
}

func InvalidVerificationCode(phone string) bool {
	var count int64
	db.Debug().Model(&Captcha{}).Where("address = ? and used = 0 and type = 1 and create_time >= ? and create_time <= ?", phone, time.Now().Add(-30*time.Minute).Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")).Count(&count)

	if count >= 3 {
		return true
	}

	return false
}

func CreateSalesMonthCache(c Cache, t1, t2 string) {
	u := GetUserById(c.Uid)
	var uSlice []User
	db.Debug().Where("parent_id = ? and sales_id = ? and user_type = 'sales'", u.Id, u.SalesId).Find(&uSlice)

	for _, v := range uSlice {
		res, _ := SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m') t, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (SELECT sum(volume) volume,sum(fee) fee, symbol_type, close_time FROM commission WHERE FIND_IN_SET(uid,(SELECT GROUP_CONCAT(id) FROM user WHERE parent_id = ? AND sales_id = ? AND user_type = 'user')) AND ib_id = ? AND commission_type = 0 and close_time between ? and ? GROUP BY symbol_type, DATE_FORMAT(close_time,'%Y-%m')) com GROUP BY DATE_FORMAT(com.close_time,'%Y-%m')`, v.ParentId, v.Id, u.Id, t2, t1)
		res1, _ := SqlOperator(`SELECT GROUP_CONCAT(id) ids FROM user WHERE parent_id = ? AND sales_id = ? AND LEFT(user_type,1) = 'L'`, v.ParentId, v.SalesId)
		if len(res) != 0 {
			for _, vv := range res {
				var salesCache SalesCache
				salesCache.Uid = v.Id
				salesCache.Time = PtoString(vv, "t")
				salesCache.Type = 1
				salesCache.Forex = ToFloat64(PtoString(vv, "forex"))
				salesCache.Metal = ToFloat64(PtoString(vv, "metal"))
				salesCache.StockCommission = ToFloat64(PtoString(vv, "stockCommission"))
				salesCache.Silver = ToFloat64(PtoString(vv, "silver"))
				salesCache.Dma = ToFloat64(PtoString(vv, "dma"))
				if res1 != nil {
					salesCache.DirectlyInviteAgents = PtoString(res1, "ids")
				}

				db.Debug().Create(&salesCache)
			}
		}
	}
}

func CreateSalesDayCache(c Cache, t1, t2 string) {
	u := GetUserById(c.Uid)
	var uSlice []User
	db.Debug().Where("parent_id = ? and sales_id = ? and user_type = 'sales'", u.Id, u.SalesId).Find(&uSlice)

	for _, v := range uSlice {
		res, _ := SqlOperators(`SELECT DATE_FORMAT(com.close_time,'%Y-%m-%d') t, SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (SELECT sum(volume) volume,sum(fee) fee, symbol_type, close_time FROM commission WHERE FIND_IN_SET(uid,(SELECT GROUP_CONCAT(id) FROM user WHERE parent_id = ? AND sales_id = ? AND user_type = 'user')) AND ib_id = ? AND commission_type = 0 and close_time between ? and ? GROUP BY symbol_type, DATE_FORMAT(close_time,'%Y-%m-%d')) com GROUP BY DATE_FORMAT(com.close_time,'%Y-%m-%d')`, v.ParentId, v.Id, u.Id, t1, t2)
		res1, _ := SqlOperator(`SELECT GROUP_CONCAT(id) ids FROM user WHERE parent_id = ? AND sales_id = ? AND LEFT(user_type,1) = 'L'`, v.ParentId, v.Id)
		if len(res) != 0 {
			for _, vv := range res {
				var salesCache SalesCache
				salesCache.Uid = v.Id
				salesCache.Time = PtoString(vv, "t")
				salesCache.Type = 2
				salesCache.Forex = ToFloat64(PtoString(vv, "forex"))
				salesCache.Metal = ToFloat64(PtoString(vv, "metal"))
				salesCache.StockCommission = ToFloat64(PtoString(vv, "stockCommission"))
				salesCache.Silver = ToFloat64(PtoString(vv, "silver"))
				salesCache.Dma = ToFloat64(PtoString(vv, "dma"))
				if res1 != nil {
					salesCache.DirectlyInviteAgents = PtoString(res1, "ids")
				}

				db.Debug().Create(&salesCache)
			}
		}

		if PtoString(res1, "ids") != "" && len(res) == 0 {
			var salesCache SalesCache
			salesCache.Uid = v.Id
			salesCache.DirectlyInviteAgents = PtoString(res1, "ids")
			db.Debug().Create(&salesCache)
		}
	}
}

func CreateSalesTotalCache(c Cache, t1 string) {
	u := GetUserById(c.Uid)
	var uSlice []User
	db.Debug().Where("parent_id = ? and sales_id = ? and user_type = 'sales'", u.Id, u.SalesId).Find(&uSlice)

	for _, v := range uSlice {
		res, _ := SqlOperators(`SELECT SUM(IF(com.symbol_type = 0,com.volume,0)) forex, SUM(IF(com.symbol_type = 1,com.volume,0)) metal, SUM(IF(com.symbol_type = 2,fee,0)) stockCommission,SUM(IF(com.symbol_type = 3,com.volume,0)) silver,SUM(IF(com.symbol_type = 4,com.volume,0)) dma FROM (SELECT sum(volume) volume,sum(fee) fee, symbol_type, close_time FROM commission WHERE FIND_IN_SET(uid,(SELECT GROUP_CONCAT(id) FROM user WHERE parent_id = ? AND sales_id = ? AND user_type = 'user')) AND ib_id = ? AND commission_type = 0 and close_time <= ? GROUP BY symbol_type) com`, v.ParentId, v.Id, u.Id, t1)
		res1, _ := SqlOperator(`SELECT GROUP_CONCAT(id) ids FROM user WHERE parent_id = ? AND sales_id = ? AND LEFT(user_type,1) = 'L'`, v.ParentId, v.Id)
		if len(res) != 0 {
			for _, vv := range res {
				var salesCache SalesCache
				salesCache.Uid = v.Id
				salesCache.Type = 3
				salesCache.Forex = ToFloat64(PtoString(vv, "forex"))
				salesCache.Metal = ToFloat64(PtoString(vv, "metal"))
				salesCache.StockCommission = ToFloat64(PtoString(vv, "stockCommission"))
				salesCache.Silver = ToFloat64(PtoString(vv, "silver"))
				salesCache.Dma = ToFloat64(PtoString(vv, "dma"))
				if res1 != nil {
					salesCache.DirectlyInviteAgents = PtoString(res1, "ids")
				}

				db.Debug().Create(&salesCache)
			}
		}

		if PtoString(res1, "ids") != "" && len(res) == 0 {
			var salesCache SalesCache
			salesCache.Uid = v.Id
			salesCache.DirectlyInviteAgents = PtoString(res1, "ids")
			db.Debug().Create(&salesCache)
		}
	}
}

func GenerateLastMonthData(uid int) {
	u := GetUserById(uid)
	t1 := time.Date(time.Now().Year(), time.Now().Month()-1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	t2 := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1).Format("2006-01-02")
	var cache Cache
	db.Debug().Where("uid = ? and time = ? and type = 2", uid, time.Now().AddDate(0, -1, 0).Format("2006-01")).First(&cache)

	if cache.Id == 0 {
		res, _ := SqlOperator(`SELECT sum(quantity) quantity, sum(commission) commission, sum(commission_difference) commission_difference, sum(fee) fee, sum(volume) volume, sum(forex) forex, sum(metal) metal, sum(stock_commission) stock_commission, sum(silver) silver, sum(dma) dma FROM cache WHERE uid = ? AND time >= ? AND time <= ?`, uid, t1, t2)

		if res != nil {
			var c1 Cache
			c1.Uid = uid
			c1.Commission = ToFloat64(PtoString(res, "commission"))
			c1.CommissionDifference = ToFloat64(PtoString(res, "commission_difference"))
			c1.Fee = ToFloat64(PtoString(res, "fee"))
			c1.Volume = ToFloat64(PtoString(res, "volume"))
			c1.Forex = ToFloat64(PtoString(res, "forex"))
			c1.Metal = ToFloat64(PtoString(res, "metal"))
			c1.StockCommission = ToFloat64(PtoString(res, "stock_commission"))
			c1.Silver = ToFloat64(PtoString(res, "silver"))
			c1.Dma = ToFloat64(PtoString(res, "dma"))
			c1.Type = 2
			c1.Time = time.Now().AddDate(0, -1, 0).Format("2006-01")
			c1.Quantity = ToInt(PtoString(res, "quantity"))
			db.Debug().Create(&c1)
		}

		db.Debug().Where("time >= ? and time <= ? and uid = ?", t1, t2, uid).Delete(&cache)
	}

	var uSlice []User
	db.Debug().Where("parent_id = ? and sales_id = ? and user_type = 'sales'", u.Id, u.SalesId).Find(&uSlice)

	for _, v := range uSlice {
		var salesCache SalesCache
		db.Debug().Where("uid = ? and time = ? and type = 1", v.Id, time.Now().AddDate(0, -1, 0).Format("2006-01")).First(&salesCache)

		if salesCache.Id == 0 {
			res, _ := SqlOperator(`SELECT sum(forex) forex, sum(metal) metal, sum(stock_commission) stock_commission, sum(silver) silver, sum(dma) dma FROM sales_cache WHERE uid = ? AND time >= ? AND time <= ?`, v.Id, t1, t2)

			if res != nil {
				var s1 SalesCache
				s1.Uid = v.Id
				s1.Forex = ToFloat64(PtoString(res, "forex"))
				s1.Metal = ToFloat64(PtoString(res, "metal"))
				s1.StockCommission = ToFloat64(PtoString(res, "stock_commission"))
				s1.Silver = ToFloat64(PtoString(res, "silver"))
				s1.Dma = ToFloat64(PtoString(res, "dma"))
				s1.Time = time.Now().AddDate(0, -1, 0).Format("2006-01")
				s1.Type = 1
				db.Debug().Create(&s1)
			}

			db.Debug().Where("time >= ? and time <= ? and uid = ?", t1, t2, uid).Delete(&salesCache)
		}
	}
}

func IsValueExist(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func CreateCache(c Cache) {
	db.Debug().Create(&c)
}

func CreateCommissionCache(uid int, t1, t2, lan string) string {
	tt := strings.Split(t1, " ")[0]

	var downExcel DownExcel
	db.Debug().Where("uid = ? and time = ?", uid, tt).First(&downExcel)

	if downExcel.Id == 0 {
		res, _ := SqlOperators(`SELECT c.id, c.ticket, c.login, c.symbol, c.volume, c.fee, c.close_time, c.commission_type, u1.true_name trade_name, u2.true_name ib_name FROM commission c
					  LEFT JOIN user u1
					  ON c.uid = u1.id
					  LEFT JOIN user u2
					  ON c.ib_id = u2.id
					  WHERE ib_id = ? and close_time >= ? and close_time <= ? ORDER BY close_time ASC`, uid, t1, t2)

		if len(res) != 0 {
			f := excelize.NewFile()

			arr := []string{"ticket", "login", "symbol", "volume", "fee", "close_time", "commission_type", "trade_name", "ib_name"}

			num := 65
			for _, v := range arr {
				f.SetCellValue("sheet1", fmt.Sprintf("%c1", num), v)
				num++
			}

			for k, v := range res {
				num = 65
				k += 2
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "ticket"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "login"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "symbol"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "volume"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "fee"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "close_time"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), GetCommissionTypeValue(ToInt(PtoString(v, "commission_type")), lan))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "trade_name"))
				num++
				f.SetCellValue("sheet1", fmt.Sprintf("%c%d", num, k), PtoString(v, "ib_name"))
			}

			fileName := fmt.Sprintf("upload/%v.xlsx", strings.Split(t1, " ")[0])

			if err := f.SaveAs(fileName); err != nil {
				fmt.Println("CreateCommissionCache err =", err)
			}

			path := fmt.Sprintf("%v.xlsx", strings.Split(t1, " ")[0])
			data := file.UploadFile(path, uid)

			fmt.Println("====data======", data)
			RemoveFile(fileName)

			if time.Now().Format("2006-01-02") != tt {
				downExcel.Uid = uid
				downExcel.Url = data
				downExcel.Time = tt

				db.Debug().Create(&downExcel)
			}

			return data
		}

		return ""
	}

	return downExcel.Url
}

func GetCommissionTypeValue(cType int, lan string) string {
	switch cType {
	case 0:
		return golbal.Wrong[lan][10530]
	case 1:
		return golbal.Wrong[lan][10531]
	case 2:
		return golbal.Wrong[lan][10532]
	}

	return ""
}

func FindCache(where string, arg ...any) (datas []Cache) {
	db.Debug().Where(where, arg...).Find(&datas)
	return
}

func GetCache(where string, arg ...any) (datas Cache) {
	db.Debug().Where(where, arg...).First(&datas)
	return
}
