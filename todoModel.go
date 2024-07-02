package abc

import "database/sql"

type (
	Token struct {
		ID         int             `json:"id"`
		Uid        int             `json:"uid"`
		Token      string          `json:"token"`
		Role       int             `json:"role"`
		Expire     int64           `json:"expire"`
		UpdateTime *sql.NullString `json:"update_time"`
	}
	Account struct {
		Login   int    `gorm:"primary_key" json:"login"`
		UserId  int    `json:"user_id"`
		RegTime string `json:"reg_time"`
		//CreateTime    	string  	   `json:"create_time"`
		Balance        float64 `json:"balance"`
		Credit         float64 `json:"credit"`
		Equity         float64 `json:"equity"`
		Margin         float64 `json:"margin"`
		MarginLevel    float64 `json:"margin_level"`
		Volume         float64 `json:"volume"`
		GroupName      string  `json:"group_name"`
		Enable         int     `json:"enable"`
		ReadOnly       int     `json:"read_only"`
		Name           string  `json:"name"`
		Country        string  `json:"country"`
		City           string  `json:"city"`
		Address        string  `json:"address"`
		Phone          string  `json:"phone"`
		Email          string  `json:"email"`
		Comment        string  `json:"comment"`
		Leverage       string  `json:"leverage"`
		RebateId       int     `json:"rebate_id"`
		LeverageStatus int     `json:"leverage_status"`
		LeverageFixed  int     `json:"leverage_fixed"`
		ApplyLeverage  int     `json:"apply_leverage"`
		Experience     int     `json:"experience"`
		ExperienceTime string  `json:"experience_time"`
		UserPath       string  `json:"user_path"`
		AB             string  `json:"ab"`
		IsMam          int     `json:"is_mam"`
		//ZipCode        string  `json:"zip_code"`
	}
	Activities struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Thumb       string `json:"thumb"`
		Tag         string `json:"tag"`
		Status      int    `json:"status"`
		Weight      int    `json:"weight"`
		Lang        string `json:"lang"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		ActivityUrl string `json:"activity_url"`
		UserType    string `json:"user_type"`
		CreateTime  string `json:"create_time"`
		Area        string `json:"area"`
	}
	ActivityBlackHouse struct {
		Id     int    `json:"id"`
		Type   string `json:"type"`
		UserId int    `json:"user_id"`
		Login  int    `json:"login"`
		Value  string `json:"value"`
	}
	ActivityCashVoucherConfig struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Type       int     `json:"type"`
		Days       int     `json:"days"`
		StartTime  string  `json:"start_time"`
		EndTime    string  `json:"end_time"`
		Amount     float64 `json:"amount"`
		InitAmount float64 `json:"init_amount"`
		Volume     float64 `json:"volume"`
		No         string  `json:"no"`
		IsDeposit  string  `json:"is_deposit"`
	}
	ActivityDisable struct {
		Id         int    `json:"id"`
		UserPath   string `json:"user_path"`
		CreateTime string `json:"create_time"`
		Type       string `json:"type"`
		Login      string `json:"login"`
		MatchType  int    `json:"match_type"`
	}
	Admin struct {
		Id         int
		AdminName  string
		RoleId     int
		CreateTime string
		Email      string
		IsDel      int
		Comment    string `json:"comment"`
	}
	Activity2023 struct {
		Id            int     `json:"id"`
		UserId        int     `json:"user_id"`
		CreateTime    string  `json:"create_time"`
		EndTime       string  `json:"end_time"`
		DepositAmount float64 `json:"deposit_amount"`
		HandselAmount float64 `json:"handsel_amount"`
		NeedVolume    float64 `json:"need_volume"`
		Volume        float64 `json:"volume"`
	}
	AdminLog struct {
		Id         int    `json:"id"`
		AdminName  string `json:"admin_name"`
		CreateTime string `json:"create_time"`
		Keys       string `json:"keys"`
		Comment    string `json:"comment"`
	}
	Announcement struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Tip        int    `json:"tip"`
		CreateTime string `json:"create_time"`
		UserId     int    `json:"user_id"`
		Clicks     int    `json:"clicks"`
	}
	AutoWithdrawDetail struct {
		Id         int     `json:"id"`
		PayId      int     `json:"pay_id"`
		Amount     float64 `json:"amount"`
		Step       int     `json:"step"`
		Result     int     `json:"result"`
		Comment    string  `json:"comment"`
		CreateTime string  `json:"create_time"`
	}
	AutoWithdrawRules struct {
		Id     int     `json:"id"`
		Sql    string  `json:"sql"`
		Title  string  `json:"title"`
		Weight int     `json:"weight"`
		Op     string  `json:"op"`
		Result float64 `json:"result"`
		Type   int     `json:"type"`
	}
	Bank struct {
		Id           int    `json:"id"`
		UserId       int    `json:"user_id"`
		BankName     string `json:"bank_name"`
		BankNo       string `json:"bank_no"`
		TrueName     string `json:"true_name"`
		CreateTime   string `json:"create_time"`
		Status       int    `json:"status"`
		Default      int    `json:"default"`
		BankAddress  string `json:"bank_address"`
		Swift        string `json:"swift"`
		Iban         string `json:"iban"`
		Files        string `json:"files"`
		Comment      string `json:"comment"`
		Area         int    `json:"area"`
		BankCardType int    `json:"bank_card_type"`
		BankCode     string `json:"bank_code"`
	}
	Captcha struct {
		Id         int    `json:"id"`
		Address    string `json:"address"`
		Code       string `json:"code"`
		CreateTime string `json:"create_time"`
		CreateAt   int64  `json:"create_at"`
		Used       int    `json:"used"`
		Type       int    `json:"type"`
		TryNum     int    `json:"try_num"`
	}
	CashVoucher struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CashNo     string  `json:"cash_no"`
		Amount     float64 `json:"amount"`
		Volume     float64 `json:"volume"`
		Status     int     `json:"status"`
		CreateTime string  `json:"create_time"`
		EndTime    string  `json:"end_time"`
		Comment    string  `json:"comment"`
	}
	Commission struct {
		Id              int     `json:"id"`
		Ticket          int     `json:"ticket"`
		Uid             int     `json:"uid"`
		Login           int     `json:"login"`
		Volume          float64 `json:"volume"`
		Symbol          string  `json:"symbol"`
		CloseTime       string  `json:"close_time"`
		IbId            int     `json:"ib_id"`
		Fee             float64 `json:"fee"`
		Comment         string  `json:"comment"`
		OrderCommission float64 `json:"order_commission"`
		CreateTime      string  `json:"create_time"`
		Difference      float64 `json:"difference"`
		Storage         float64 `json:"storage"`
		Profit          float64 `json:"profit"`
		CommissionType  int     `json:"commission_type"`
		UserPath        string  `json:"user_path"`
		SymbolType      int     `json:"symbol_type"`
	}
	CommissionSetCustom struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_Id"`
		Symbol     string  `json:"symbol"`
		Amount     float64 `json:"amount"`
		Type       int     `json:"type"`
		Status     int     `json:"status"`
		CreateTime string  `json:"create_Time"`
		UserPath   string  `json:"user_Path"`
	}
	CommissionConfig struct {
		Id     int     `json:"id"`
		Symbol string  `json:"symbol"`
		Super  float64 `json:"super"`
		Organ  float64 `json:"organ"`
		Person float64 `json:"person"`
		Point  float64 `json:"point"`
		Stock  int     `json:"stock"`
		DMA    int     `json:"dma"`
	}
	Config struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	Contest struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		TrueName   string  `json:"true_name"`
		CreateTime string  `json:"create_time"`
		Login      int     `json:"login"`
		InitAmount float64 `json:"init_amount"`
		StartTime  string  `json:"start_time"`
		EndTime    string  `json:"end_time"`
		Rank       int     `json:"rank"`
		Type       int     `json:"type"`
		Status     int     `json:"status"`
		Balance    float64 `json:"balance"`
		ProfitRate float64 `json:"profit_rate"`
		UpdateTime string  `json:"update_time"`
	}
	CouponCode struct {
		Id         string `json:"id"`
		CardNo     string `json:"card_no"`
		CreateTime string `json:"create_time"`
		UseTime    string `json:"use_time"`
		CreateId   string `json:"create_id"`
		UseId      string `json:"use_id"`
		Type       string `json:"type"`
		Value      string `json:"value"`
		Volume     string `json:"volume"`
	}
	CouponDetail struct {
		Id            int     `json:"id"`
		UserId        int     `json:"user_id"`
		CouponNo      string  `json:"coupon_no"`
		DepositAmount float64 `json:"deposit_amount"`
		PayAmount     float64 `json:"pay_amount"`
		CreateTime    string  `json:"create_time"`
		Volume        float64 `json:"volume"`
		EndTime       string  `json:"end_time"`
		Status        int     `json:"status"`
		SystemTime    string  `json:"system_time"`
		Login         int     `json:"login"`
	}
	Customer struct {
		Id         int     `json:"id"`
		OrderNo    string  `json:"order_no"`
		UserId     int     `json:"user_id"`
		Name       string  `json:"name"`
		Mobile     string  `json:"mobile"`
		Email      string  `json:"email"`
		FollowTxt  string  `json:"follow_txt"`
		CreateTime string  `json:"create_time"`
		FollowTime *string `json:"follow_time"`
		Type       string  `json:"type"`
		Status     int     `json:"status"`
		NextTime   string  `json:"next_time"`
		ExpireTime string  `json:"expire_time"`
		Aid        int     `json:"aid"`
		Only       int     `json:"only"`
	}
	DemoAccount struct {
		Id         int    `json:"id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		PhoneCode  string `json:"phone_code"`
		PhoneNo    string `json:"phone_no"`
		Email      string `json:"email"`
		Country    string `json:"country"`
		CreateTime string `json:"create_time"`
		Login      string `json:"login"`
		Password   string `json:"password"`
		Ip         string `json:"ip"`
		Status     int    `json:"status"`
		IbId       int    `json:"ib_id"`
		UserId     int    `json:"user_id"`
	}
	Edm struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`
		CreateTime string `json:"create_time"`
		UpdateTime string `json:"update_time"`
		Content    string `json:"content"`
		Json       string `json:"json"`
		CateId     int    `json:"cate_id"`
	}
	EdmDetail struct {
		Id           int    `json:"id"`
		CateId       int    `json:"cate_id"`
		CreateTime   string `json:"create_time"`
		UserId       int    `json:"user_id"`
		UserName     string `json:"user_name"`
		EmailAddress string `json:"email_address"`
		TplId        int    `json:"tpl_id"`
		Content      string `json:"content"`
		SendTime     string `json:"send_time"`
		Status       int    `json:"status"`
		Title        string `json:"title"`
		RetMessage   string `json:"ret_message"`
	}
	Follow struct {
		Id         int     `json:"id"`
		IbId       int     `json:"ib_id"`
		SignalNo   int     `json:"signal_no"`
		Model      int     `json:"model"`
		Proportion float64 `json:"proportion"`
		FollowNo   int     `json:"follow_no"`
		Status     int     `json:"status"`
		CreateTime string  `json:"create_time"`
		SignalExt  string  `json:"signal_ext"`
		Ext        string  `json:"ext"`
	}
	Fund struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Type       int     `json:"type"`
		CreateTime string  `json:"create_time"`
		Amount     float64 `json:"amount"`
	}
	InsuranceCoupon struct {
		Id            int     `json:"id"`
		UserId        int     `json:"user_id"`
		CouponNo      string  `json:"coupon_no"`
		Value         float64 `json:"value"`
		PayAmount     float64 `json:"pay_amount"`
		InsuranceCate string  `json:"insurance_cate"`
		Login         int     `json:"login"`
		StartTime     string  `json:"start_time"`
		EndTime       string  `json:"end_time"`
		Comment       string  `json:"comment" sql:"-"`
		CreateTime    string  `json:"create_time"`
		Compensation  string  `json:"compensation"`
		Status        int     `json:"status"`
		Equity        float64 `json:"equity"`
	}
	Interest struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Login      int     `json:"login"`
		CreateTime string  `json:"create_time"`
		Fee        float64 `json:"fee"`
		FreeMargin float64 `json:"free_margin"`
		Balance    float64 `json:"balance"`
		Equity     float64 `json:"equity"`
		Comment    string  `json:"comment"`
		Type       int     `json:"type"`
	}
	InviteCode struct {
		Id      int    `json:"id"`
		UserId  int    `json:"user_id"`
		Code    string `json:"code"`
		Name    string `json:"name"`
		Rights  string `json:"rights"`
		Type    string `json:"type"`
		Comment string `json:"comment"`
	}
	Log struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		UserName   string `json:"user_name"`
		Handle     string `json:"handle"`
		Type       string `json:"type"`
		Content    string `json:"content"`
		CreateTime string `json:"create_time"`
	}
	LotteryConfig struct {
		Id    int     `json:"id"`
		Name  string  `json:"name"`
		Type  string  `json:"type"`
		Value float64 `json:"value"`
		O     float64 `json:"0"`
		S1    float64 `json:"s1"`
		S2    float64 `json:"s2"`
		S3    float64 `json:"s3"`
		S4    float64 `json:"s4"`
		S5    float64 `json:"s5"`
	}
	LotteryDetail struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Level      string  `json:"level"`
		Result     float64 `json:"result"`
		ResultType string  `json:"result_type"`
		CreateTime string  `json:"create_time"`
		ResultTime string  `json:"result_time"`
		Status     int     `json:"status"`
		Comment    string  `json:"comment"`
		CouponId   int     `json:"coupon_id"`
		AssignId   int     `json:"assign_id"`
		//ByVolume					float64				`json:"by_volume"`
	}
	MailLog struct {
		Id         int    `json:"id"`
		Address    string `json:"address"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		CreateTime string `json:"create_time"`
	}
	MailTpl struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Tip        int    `json:"tip"`
		CreateTime string `json:"create_time"`
	}
	MamDetail struct {
		Id          int     `json:"id"`
		UserId      int     `json:"user_id"`
		MamId       int     `json:"mam_id"`
		UserLogin   int     `json:"user_login"`
		JoinTime    string  `json:"join_time"`
		UserAmount  float64 `json:"user_amount"`
		MamType     string  `json:"mam_type"`
		IsManager   int     `json:"is_manager"`
		Risk        float64 `json:"risk"`
		UserStatus  int     `json:"user_status"`
		FundStatus  int     `json:"fund_status"`
		Rate        float64 `json:"rate"`
		EndStatus   int     `json:"end_status"`
		DelayStatus int     `json:"delay_status"`
		Path        string  `json:"path"`
	}
	MamFund struct {
		Id         int     `json:"id"`
		MamId      int     `json:"mam_id"`
		UserId     int     `json:"user_id"`
		Login      int     `json:"login"`
		FundTime   string  `json:"fund_time"`
		Comment    string  `json:"comment"`
		Amount     float64 `json:"amount"`
		Deposit    int     `json:"deposit"`
		BackAmount float64 `json:"back_amount"`
	}
	MamProject struct {
		Id            int     `json:"id"`
		IpId          int     `json:"ip_id"`
		MamNo         string  `json:"mam_no"`
		Title         string  `json:"title"`
		MamStatus     int     `json:"mam_status"`
		MamAmount     float64 `json:"mam_amount"`
		CloseAmount   float64 `json:"close_amount"`
		MamCount      int     `json:"mam_count"`
		Login         int     `json:"login"`
		ManagerId     int     `json:"manager_id"`
		CreateTime    string  `json:"create_time"`
		SystemTime    string  `json:"system_time"`
		StartTime     string  `json:"start_time"`
		EndTime       string  `json:"end_time"`
		MamPeriod     int     `json:"mam_period"`
		NextFundTime  string  `json:"next_fund_time"`
		PrevTime      string  `json:"prev_time"`
		Description   string  `json:"description"`
		MamType       int     `json:"mam_type"`
		Protect       int     `json:"protect"`        //保本保息
		ProtectProfit float64 `json:"protect_profit"` //每一期金额
		EndStatus     int     `json:"end_status"`
		DelayStatus   int     `json:"delay_status"`
		Alarm         int     `json:"alarm"`
		FundDueTime   string  `json:"fund_due_time"`
		ManageRate    float64 `json:"manage_rate"`
	}
	MoveRewards struct {
		Id            int     `json:"id"`
		UserId        int     `json:"user_id"`
		CreateTime    string  `json:"create_time"`
		AdminTime     string  `json:"admin_time"`
		Status        int     `json:"status"`
		DepositAmount float64 `json:"deposit_amount"`
		RewardAmount  float64 `json:"reward_amount"`
		PayTime       string  `json:"pay_time"`
		Files         string  `json:"files"`
		Comment       string  `json:"comment"`
	}
	Orders struct {
		OrderId        int     `gorm:"primary_key" json:"order_id"`
		Login          int     `json:"login"`
		Symbol         string  `json:"symbol"`
		SymbolType     int     `json:"symbol_type"`
		Cmd            int     `json:"cmd"`
		Volume         float64 `json:"volume"`
		OpenPrice      float64 `json:"open_price"`
		OpenTime       string  `json:"open_time"`
		ClosePrice     float64 `json:"close_price"`
		CloseTime      string  `json:"close_time"`
		Sl             float64 `json:"sl"`
		Tp             float64 `json:"tp"`
		Profit         float64 `json:"profit"`
		Storage        float64 `json:"storage"`
		Commission     float64 `json:"commission"`
		Taxes          float64 `json:"taxes"`
		Comment        string  `json:"comment"`
		IsCommission   int     `json:"is_commission"`
		CommissionTime string  `json:"commission_time"`
		UserId         string  `json:"user_id"`
		Assign         string  `json:"assign"`
		AB             string  `json:"ab"`
		Spent          float64 `json:"spent"`
	}
	Payment struct {
		Id               int     `json:"id"`
		OrderNo          string  `json:"order_no"`
		UserId           int     `json:"user_id"`
		Login            int     `json:"login"`
		CreateTime       string  `json:"create_time"`
		Amount           float64 `json:"amount"`
		AmountRmb        float64 `json:"amount_rmb"`
		PayName          string  `json:"pay_name"`
		PayTime          string  `json:"pay_time"`
		Comment          string  `json:"comment"`
		Status           int     `json:"status"`
		Type             string  `json:"type"`
		Intro            string  `json:"intro"`
		TransferLogin    int     `json:"transfer_login"`
		PayFee           float64 `json:"pay_fee"`
		AStatus          float64 `json:"a_status"`
		BStatus          float64 `json:"b_status"`
		CStatus          float64 `json:"c_status"`
		WireDoc          string  `json:"wire_doc"`
		UserPath         string  `json:"user_path"`
		PayUrl           string  `json:"pay_url"`
		PayCompanyAmount float64 `json:"pay_company_amount"`
		ExchangeRate     float64 `json:"exchange_rate"`
		WaterNumber      string  `json:"water_number"`
	}
	PaymentLog struct {
		Id         int    `json:"id"`
		PaymentId  int    `json:"payment_id"`  //default 0  not null,
		CreateTime string `json:"create_time"` //default '' not null comment '创建时间',
		Comment    string `json:"comment"`     //default '' not null comment '备注'
	}
	PaymentConfig struct {
		Id           int     `json:"id"`
		Name         string  `json:"name"`
		Status       int     `json:"status"`
		MaxAmount    float64 `json:"max_amount"`
		MinAmount    float64 `json:"min_amount"`
		ExchangeRate float64 `json:"exchange_rate"`
		OpenTime     string  `json:"open_time"`
		CloseTime    string  `json:"close_time"`
		Weight       int     `json:"weight"`
		KeySecret    string  `json:"key_secret"`
		Type         int     `json:"type"`
		QuickPay     int     `json:"quick_pay"`
	}
	Propaganda struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`
		Cate       int    `json:"cate"`
		Content    string `json:"content"`
		UserType   string `json:"user_type"`
		CreateTime string `json:"create_time"`
		Thumb      string `json:"thumb"`
		Source     string `json:"source"`
		Path       string `json:"path"`
		DocType    string `json:"doc_type"`
		DocTime    string `json:"doc_time"`
		Url        string `json:"url"`
	}
	PwdBlackHouse struct {
		Id   int
		Keys string
	}
	Poster struct {
		Id         int    `json:"id"`
		Title      string `json:"title"`       //default '' not null comment '标题',
		Sample     string `json:"sample"`      //default '' not null comment '样例',
		Resource   string `json:"resource"`    //default '' not null comment '资源',
		Content    string `json:"content"`     //default '' not null comment '描述',
		CreateTime string `json:"create_time"` //default '' not null
	}
	Question struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		CreateTime string `json:"create_time"`
		UpdateTime string `json:"update_time"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Status     int    `json:"status"`
		Weight     int    `json:"weight"`
		Tag        string `json:"tag"`
	}
	QuestionDetail struct {
		Id         int    `json:"id"`
		QId        int    `json:"q_id"`
		AdminId    int    `json:"admin_id"`
		Content    string `json:"content"`
		CreateTime string `json:"create_time"`
		ReadTime   string `json:"read_time"`
		IsRead     int    `json:"is_read"`
		FilePath   string `json:"file_path"`
		FileName   string `json:"file_name"`
		TelegramId int    `json:"telegram_id"`
	}
	RiskgroupMapping struct {
		Id               int    `json:"id"`
		CurrentRiskGroup string `json:"current_risk_group"`
		ClientApply      string `json:"client_apply"`
		NewRiskGroup     string `json:"new_risk_group"`
	}
	SalesCommission struct {
		Id             int     `json:"id"`
		UserId         int     `json:"user_id"`
		CreateTime     string  `json:"create_time"`
		Amount         float64 `json:"amount"`
		Comment        string  `json:"comment"`
		Login          int     `json:"login"`
		Volume         float64 `json:"volume"`
		CommissionType int     `json:"commission_type"`
		SettleDate     string  `json:"settle_date"`
		Status         int     `json:"status"`
	}
	SalesCommissionSet struct {
		Id             int     `json:"id"`
		UserId         int     `json:"user_id"`
		CommissionType int     `json:"commission_type"`
		Amount         float64 `json:"amount"`
		Config         string  `json:"Config"`
	}
	CommissionSetConfig struct {
		Start float64
		End   float64
		Value float64
	}
	SalesTransfer struct {
		Id         int     `json:"id"`
		OldPath    string  `json:"old_path"`
		NewPath    string  `json:"new_path"`
		Month      string  `json:"month"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Deposit    float64 `json:"deposit"`
		Withdraw   float64 `json:"withdraw"`
		Forex      float64 `json:"forex"`
		Metal      float64 `json:"metal"`
		Silver     float64 `json:"silver"`
		Stock      float64 `json:"stock"`
		Dma        float64 `json:"dma"`
		MtDeposit  float64 `json:"mt_deposit"`
		MtWithdraw float64 `json:"mt_withdraw"`
	}
	ScoreDetail struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Title      string  `json:"title"`
		CreateTime string  `json:"create_time"`
		PayAmount  float64 `json:"pay_amount"`
		Amount     float64 `json:"amount"`
		Balance    float64 `json:"balance"`
		OrderNo    string  `json:"order_no"`
		GoodsId    int     `json:"goods_id"`
		Address    string  `json:"address"`
	}
	ScoreConfig struct {
		Id        int     `json:"id"`
		Name      string  `json:"name"`
		PayAmount float64 `json:"pay_amount"`
		Amount    float64 `json:"amount"`
		Unit      string  `json:"unit"`
		StartTime string  `json:"start_time"`
		EndTime   string  `json:"end_time"`
		UnitValue float64 `json:"unit_value"`
		IsGoods   int     `json:"is_goods"`
		BuyNum    int     `json:"buy_num"`
	}
	ScoreLog struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Ticket     int     `json:"ticket"`
		Login      int     `json:"login"`
		Symbol     string  `json:"symbol"`
		Cmd        int     `json:"cmd"`
		Volume     float64 `json:"volume"`
		CloseTime  string  `json:"close_time"`
		CreateTime string  `json:"create_time"`
		Amount     float64 `json:"amount"`
		Type       int     `json:"type"`
		UseAmount  float64 `json:"use_amount"`
	}
	Signal struct {
		Id         int     `json:"id"`
		Account    int     `json:"account"`
		Password   string  `json:"password"`
		Server     string  `json:"server"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Ibs        string  `json:"ibs"`
		Balance    float64 `json:"balance"`
		Equity     float64 `json:"equity"`
		Leverage   int64   `json:"leverage"`
		Status     int     `json:"status"`
		Ext        string  `json:"ext"`
	}
	User struct {
		Id            int             `json:"id"`
		Username      string          `json:"username"`
		Password      string          `json:"password"`
		CreateTime    string          `json:"create_time"`
		Status        int             `json:"status"`
		AuthStatus    int             `json:"auth_status"`
		LoginTime     *sql.NullString `json:"login_time"`
		LoginTimes    int             `json:"login_times"`
		UserType      string          `json:"user_type"`
		ParentId      int             `json:"parent_id"`
		Email         string          `json:"email"`
		TrueName      string          `json:"true_name"`
		Mobile        string          `json:"mobile"`
		CustomerId    int             `json:"customer_id"`
		Phonectcode   string          `json:"phonectcode"`
		InviteCode    string          `json:"invite_code"`
		SalesId       int             `json:"sales_id"`
		RebateId      int             `json:"rebate_id"`
		Path          string          `json:"path"`
		SalesType     string          `json:"sales_type"`
		IbNo          string          `json:"ib_no"`
		Link          string          `json:"link"`
		CoinStatus    int             `json:"coin_status"`
		RebateCate    int             `json:"rebate_cate"`
		RebateMulti   float64         `json:"rebate_multi"`
		SomeId        int             `json:"some_id"`
		WalletBalance float64         `json:"wallet_balance"`
		RiskComment   string          `json:"risk_comment"`
		PathFull      string          `json:"path_full"`
	}
	UserAddress struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		CreateTime string `json:"create_time"`
		TrueName   string `json:"true_name"`
		Area       string `json:"area"`
		Phone      string `json:"phone"`
		Zip        string `json:"zip"`
		Address    string `json:"address"`
	}
	UserFile struct {
		Id         int    `json:"id"`
		FileName   string `json:"file_name"`
		FileType   string `json:"file_type"`
		FileSize   int64  `json:"file_size"`
		CreateTime string `json:"create_time"`
		UserId     int    `json:"user_id"`
		Comment    string `json:"comment"`
		Status     int    `json:"status"`
		Front      string `json:"front"`
	}
	UserInfo struct {
		UserId          int             `gorm:"primary_key" json:"user_id"`
		Title           string          `json:"title"`
		Surname         string          `json:"surname"`
		Lastname        string          `json:"lastname"`
		Nickname        string          `json:"nickname"`
		Birthday        string          `json:"birthday"`
		IdentityType    string          `json:"identity_type"`
		Identity        string          `json:"identity"`
		Nationality     string          `json:"nationality"`
		Country         string          `json:"country"`
		City            string          `json:"city"`
		Forexp          string          `json:"forexp"`
		Otherexp        string          `json:"otherexp"`
		Investfreq      string          `json:"investfreq"`
		CurrencyType    string          `json:"currency_type"`
		AccountType     string          `json:"account_type"`
		Platform        string          `json:"platform"`
		Homephone       string          `json:"homephone"`
		Address         string          `json:"address"`
		AddressDate     string          `json:"address_date"`
		Birthcountry    string          `json:"birthcountry"`
		Face            string          `json:"face"`
		InfoStatus      int             `json:"info_status"`
		BankName        string          `json:"bank_name"`
		BankNo          string          `json:"bank_no"`
		BankAddress     string          `json:"bank_address"`
		Swift           string          `json:"swift"`
		BankCountry     string          `json:"bank_country"`
		UsdtAddress     string          `json:"usdt_address"`
		BankStatus      int             `json:"bank_status"`
		Agreement       string          `json:"agreement"`
		AgreementFee    string          `json:"agreement_fee"`
		AgreementTime   *sql.NullString `json:"agreement_time"`
		Iban            string          `json:"iban"`
		SortCode        string          `json:"sort_code"`
		ChineseName     string          `json:"chinese_name"`
		Company         string          `json:"company"`
		IdentityAddress string          `json:"identity_address"`
		ChineseIdentity string          `json:"chinese_identity"`
	}
	UserLog struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		Type       string `json:"type"`
		Email      string `json:"email"`
		CreateTime string `json:"create_time"`
		Ip         string `json:"ip"`
		Content    string `json:"content"`
	}
	UserMessage struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		CreateTime string `json:"create_time"`
		ReadTime   string `json:"read_time"`
		TemplateId int    `json:"template_id"`
		TemplateZh string `json:"template_zh"`
		TemplateHk string `json:"template_hk"`
		TemplateEn string `json:"template_en"`
		Status     int    `json:"status"`
	}
	UserMore struct {
		UserId            int    `gorm:"primary_key" json:"user_id"`
		Ispolitic         string `json:"ispolitic"`
		Istax             string `json:"istax"`
		Isusa             string `json:"isusa"`
		Isforusa          string `json:"isforusa"`
		Isearnusa         string `json:"isearnusa"`
		Employment        string `json:"employment"`
		Business          string `json:"business"`
		Position          string `json:"position"`
		Employername      string `json:"employername"`
		Incomesource      string `json:"incomesource"`
		Income            string `json:"income"`
		Netasset          string `json:"netasset"`
		Investaim         string `json:"investaim"`
		ExtraDoc          int    `json:"extra_doc"`
		NotRemindAgain    int    `json:"not_remind_again"`
		IsMam             int    `json:"is_mam"`
		AccountStatus     int    `json:"account_status"`
		TransactionStatus int    `json:"transaction_status"`
		FinancialStatus   int    `json:"financial_status"`
		DocumentsStatus   int    `json:"documents_status"`
		Phonectcode       string `json:"phonectcode"`
		Mobile            string `json:"mobile"`
	}
	UserAbility struct {
		Id              int    `json:"id"`
		UserId          int    `json:"user_id"`
		UserPath        string `json:"user_path"`
		Path            []int
		UpdateTime      string  `json:"update_time"`
		Deposit         float64 `json:"deposit"`
		Withdraw        float64 `json:"withdraw"`
		MtDeposit       float64 `json:"mt_deposit"`
		MtWithdraw      float64 `json:"mt_withdraw"`
		Forex           float64 `json:"forex"`
		Metal           float64 `json:"metal"`
		Silver          float64 `json:"silver"`
		Stock           float64 `json:"stock"`
		Dma             float64 `json:"dma"`
		Profit          float64 `json:"profit"`
		Commission      float64 `json:"commission"`
		Storage         float64 `json:"storage"`
		Balance         float64 `json:"balance"`
		Credit          float64 `json:"credit"`
		Equity          float64 `json:"equity"`
		PositionVolume  float64 `json:"position_volume"`
		Wallet          float64 `json:"wallet"`
		StockCommission float64 `json:"stock_commission"`
	}
	UserExamine struct {
		UserId   int `json:"user_id"`
		Standard int `json:"standard"`
		Status   int `json:"status"`
		Result   int `json:"result"` //达标的个数
		//SomeStandard		int		`json:"some_standard"`
		UpdateTime      string `json:"update_time"`
		NextExamineTime string `json:"next_examine_time"`
		Comment         string `json:"comment"`
	}
	UserWallet struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Address    string  `json:"address"`
		Symbol     string  `json:"symbol"`
		PrivateKey string  `json:"private_key"`
		HexKey     string  `json:"hex_key"`
		Status     int     `json:"status"`
		Balance    float32 `json:"balance"`
	}
	Version struct {
		Id         int    `json:"id"`
		Version    string `json:"version"`
		Content    string `json:"content"`
		Level      int    `json:"level"`
		Url        string `json:"url"`
		WgtUrl     string `json:"wgt_url"`
		WgtVersion string `json:"wgt_version"`
	}
	UserVip struct {
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Score      float64 `json:"score"`
		Grade      int     `json:"grade"`
		UpdateTime string  `json:"update_time"`
		UpTime     string  `json:"up_time"`
	}
	UserVipCash struct {
		Id              int    `json:"id"`
		UserId          int    `json:"user_id"`
		OrderNo         string `json:"order_no"`
		CreateTime      string `json:"create_time"`
		PayAmount       int    `json:"pay_amount"`
		DeductionAmount int    `json:"deduction_amount"`
		Status          int    `json:"status"`
		UseTime         string `json:"use_time"`
		PayId           int    `json:"pay_id"`
		Comment         string `json:"comment"`
	}
	lang struct {
		Zh string `json:"zh"`
		Hk string `json:"hk"`
		En string `json:"en"`
	}
	UserVipConfig struct {
		Id              int     `json:"id"`
		GradeId         int     `json:"grade_id"`
		GradeName       string  `json:"grade_name"`
		Flag            string  `json:"flag"`
		Invite          int     `json:"invite"`
		Coupon          int     `json:"coupon"`
		CashCoupon      string  `json:"cash_coupon"`
		BirthCash       int     `json:"birth_cash"`
		MonthTradeScore float64 `json:"month_trade_score"`
		ScoreMultiple   float64 `json:"score_multiple"`
		FreeWithdraw    int     `json:"free_withdraw"`
	}
	UserVipRights struct {
		Id        int    `json:"id"`
		Title     string `json:"title"`
		Rights    int    `json:"rights"`
		ContentZh string `json:"content_zh"`
		ContentHk string `json:"content_hk"`
		ContentEn string `json:"content_en"`
		Href      string `json:"href"`
	}
	UserVipFlow struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Score      float64 `json:"score"`
		Grade      int     `json:"grade"`
		PrevGrade  int     `json:"prev_grade"`
	}
	Wallet struct {
		Id          int    `json:"id"`
		Address     string `json:"address"`
		AddressType string `json:"address_type"`
		Name        string `json:"name"`
		Tag         string `json:"tag"`
		CreateTime  string `json:"create_time"`
		UserId      int    `json:"user_id"`
		IsDel       int    `json:"is_del"`
		Status      int    `json:"status"`
	}
	Workorder struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		OrderNo    string `json:"order_no"`
		UserName   string `json:"user_name"`
		Title      string `json:"title"`
		CreateTime string `json:"create_time"`
		Problem    string `json:"problem"`
		Status     int    `json:"status"`
		ReplayTime string `json:"replay_time"`
		UpdateTime string `json:"update_time"`
	}
	Rights struct {
		Id       int    `json:"id"`
		TypeName string `json:"type_name"`
		Content  string `json:"content"`
	}
	Works struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		UserName   string `json:"user_name"`
		CreateTime string `json:"create_time"`
		Problem    string `json:"problem"`
		Status     int    `json:"status"`
	}
	RebateConfig struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Forex       string `json:"forex"`
		ForexConfig string `json:"forex_config"`
		Metal       string `json:"metal"`
		MetalConfig string `json:"metal_config"`
		GroupName   string `json:"group_name"`
		GroupType   string `json:"group_type"`
		IsInvite    string `json:"is_invite"`
		Leverage    string `json:"leverage"`
		DisplayName string `json:"display_name"`
	}
	ExtraRate struct {
		Id         int     `json:"id"`
		CreateTime string  `json:"create_time"`
		Value      float64 `json:"value"`
		Xauusd     float64 `json:"xauusd"`
		Goal       float64 `json:"goal"`
		Deposits   float64 `json:"deposits"`
	}
	MamLog struct {
		Id         int    `json:"id"`
		MamId      int    `json:"mam_id"`
		UserId     int    `json:"user_id"`
		AdminId    int    `json:"admin_id"`
		CreateTime string `json:"create_time"`
		Content    string `json:"content"`
	}
	CustomerDetail struct {
		Id          int    `json:"id"`
		CId         int    `json:"c_id"`
		UserId      int    `json:"user_id"`
		CreateTime  string `json:"create_time"`
		Status      int    `json:"status"`
		UserName    string `json:"user_name"`
		CName       string `json:"c_name"`
		CEmail      string `json:"c_email"`
		CPhone      string `json:"c_phone"`
		Content     string `json:"content"`
		Files       string `json:"files"`
		NextTime    string `json:"next_time"`
		NewUserId   int    `json:"new_user_id"`
		NewUserName string `json:"new_user_name"`
	}
	Coupon struct {
		Id            int     `json:"id"`
		Type          int     `json:"type"`
		UserId        int     `json:"user_id"`
		CouponNo      string  `json:"coupon_no"`
		Amount        float64 `json:"amount"`
		Status        int     `json:"status"`
		Comment       string  `json:"comment"`
		CreateTime    string  `json:"create_time"`
		UsedStartTime string  `json:"used_start_time"`
		UsedEndTime   string  `json:"used_end_time"`
		Login         string  `json:"login"`
		Credit        float64 `json:"credit"`
	}
	MamWallet struct {
		Id         int     `json:"id"`
		PammId     int     `json:"pamm_id"`
		WalletId   int     `json:"wallet_id"`
		ToWalletId int     `json:"to_wallet_id"`
		Value      float64 `json:"value"`
		ToName     string  `json:"to_name"`
	}
	UserActivity struct {
		Id       int     `json:"id"`
		UserId   int     `json:"user_id"`
		ByVolume float64 `json:"by_volume"`
		ByUser   int     `json:"by_user"`
	}
	CreditDetail struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		Login      int     `json:"login"`
		CreateTime string  `json:"create_time"`
		OverTime   string  `json:"over_time"`
		DeductTime string  `json:"deduct_time"`
		Balance    float64 `json:"balance"`
		Source     int     `json:"source"`
		Status     int     `json:"status"`
		Equity     float64 `json:"equity"`
		Comment    string  `json:"comment"`
		Volume     float64 `json:"volume"`
		CouponNo   string  `json:"coupon_no"`
		Deposit    float64 `json:"deposit"`
	}
	UserMessageConfig struct {
		Id        int    `json:"id"`
		Title     string `json:"title"`
		ContentZh string `json:"content_zh"`
		ContentHk string `json:"content_hk"`
		ContentEn string `json:"content_en"`
	}
	Wage struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		Status     int     `json:"status"`
		TrueName   string  `json:"true_name"`
		Address    string  `json:"-"`
		Type       string  `json:"type"`
		Amount     float64 `json:"amount"`
		Comment    string  `json:"comment"`
	}
	FootballTeam struct {
		TeamId int     `json:"team_id"`
		Name   string  `json:"name"`
		Path   string  `json:"path"`
		Status int     `json:"status"`
		Odds   float64 `json:"odds"`
	}
	FootballMatch struct {
		Id          int     `json:"id"`
		HomeTeam    string  `json:"home_team"`
		AwayTeam    string  `json:"away_team"`
		MatchTime   string  `json:"match_time"`
		EndTime     string  `json:"end_time"`
		Result      string  `json:"result"`
		Win         float64 `json:"win"`
		Draw        float64 `json:"draw"`
		Lose        float64 `json:"lose"`
		Status      int     `json:"status"`
		HomeTeamPic string  `json:"home_team_pic"`
		AwayTeamPic string  `json:"away_team_pic"`
		GameRound   string  `json:"game_round"`
		HomeScore   int     `json:"home_score"`
		AwayScore   int     `json:"away_score"`
	}
	FootballOrders struct {
		Id         int     `json:"id"`
		UserId     int     `json:"user_id"`
		CreateTime string  `json:"create_time"`
		MatchId    int     `json:"match_id"`
		PayScore   float64 `json:"pay_score"`
		Status     int     `json:"status"`
		HomeName   string  `json:"home_name"`
		HomeScore  int     `json:"home_score"`
		AwayName   string  `json:"away_name"`
		AwayScore  int     `json:"away_score"`
		Score      float64 `json:"score"`
		Rate       float64 `json:"rate"`
		Type       string  `json:"type"`
		EndTime    string  `json:"end_time"`
		HomePic    string  `json:"home_pic"`
		AwayPic    string  `json:"away_pic"`
		MatchType  int     `json:"match_type"`
	}
	MonitorAccount struct {
		Login      int     `json:"login"`
		CreateTime string  `json:"create_time"`
		Comment    string  `json:"comment"`
		Weight     int     `json:"weight"`
		UserPath   string  `json:"user_path"`
		OpenTime   string  `json:"open_time"`
		Equity     float64 `json:"equity"`
		Balance    float64 `json:"balance"`
		Credit     float64 `json:"credit"`
		TrueName   string  `json:"true_name"`
		Profits    float64 `json:"profits"`
		YesProfits float64 `json:"yes_profits"`
		Commission float64 `json:"commission"`
		Rebate     float64 `json:"rebate"`
		YesRebate  float64 `json:"yes_rebate"`
		YesDate    string  `json:"yes_date"`
	}
	InviteCodeSet struct {
		Id       int    `json:"id"`
		UserType string `json:"user_type"`
		CodeType string `json:"code_type"`
		Status   int    `json:"status"`
		IsAdmin  int    `json:"is_admin"`
	}
	CommissionSet struct {
		Id      int     `json:"id"`
		Name    string  `json:"name"`
		Cate    int     `json:"cate"`
		Partner string  `json:"partner"`
		Symbol  string  `json:"symbol"`
		Amount  float64 `json:"amount"`
		Dma     int     `json:"dma"`
		Stock   int     `json:"stock"`
		Type    int     `json:"type"`
	}
	PartnerPlan struct {
		Id         int    `json:"id"`
		Sql        string `json:"sql"`
		Sql2       string `json:"sql2"`
		Percentage string `json:"percentage"`
		PlanAmount string `json:"plan_amount"`
		Values     struct {
			Level string `json:"level"`
			Value int    `json:"value"`
		} `json:"values"`
		Type int `json:"type"`
	}
	Ads struct {
		Id         int    `json:"id"`
		PositionId int    `json:"position_id"`
		Content    string `json:"content"`
		Link       string `json:"link"`
		Lang       string `json:"lang"`
	}
	UserAuditLog struct {
		UserId     int    `json:"user_id"`
		CreateTime string `json:"create_time"`
		Comment    int    `json:"comment"`
		AdminName  string `json:"admin_name"`
		Old        int    `json:"old"`
	}
	UserG2code struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		Secret     string `json:"secret"`
		CreateTime string `json:"create_time"`
		Status     int    `json:"status"`
	}
	Documents struct {
		Id         int    `json:"id"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		CreateTime int    `json:"create_time"`
		Size       int    `json:"size"`
		Thumb      string `json:"thumb"`
		Language   string `json:"language"`
		Source     int    `json:"source"`
	}
	RegionBlackHouse struct {
		Id      int    `json:"id"`
		Code    string `json:"code"`
		comment string `json:"comment"`
	}
	Config2 struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}
	LiveChat struct {
		Id      int    `json:"id"`
		Name    string `json:"name"`
		Content string `json:"content"`
		Status  int    `json:"status"`
	}
	InfoBlackHouse struct {
		Id         int    `json:"id"`
		Address    string `json:"address"`
		CreateTime string `json:"create_time"`
		Status     int    `json:"status"`
	}
	Cache struct {
		Id                   int     `json:"id"`
		Uid                  int     `json:"uid"`
		Commission           float64 `json:"commission"`
		CommissionDifference float64 `json:"commission_difference"`
		Fee                  float64 `json:"fee"`
		Volume               float64 `json:"volume"`
		Forex                float64 `json:"forex"`
		Metal                float64 `json:"metal"`
		StockCommission      float64 `json:"stock_commission"`
		Silver               float64 `json:"silver"`
		Dma                  float64 `json:"dma"`
		Time                 string  `json:"time"`
		Type                 int     `json:"type"`
		Comment              string  `json:"comment"`
		Quantity             int     `json:"quantity"`
	}
	SalesCache struct {
		Id                   int     `json:"id"`
		Uid                  int     `json:"uid"`
		Forex                float64 `json:"forex"`
		Metal                float64 `json:"metal"`
		StockCommission      float64 `json:"stock_commission"`
		Silver               float64 `json:"silver"`
		Dma                  float64 `json:"dma"`
		DirectlyInviteAgents string  `json:"directly_invite_agents"`
		Time                 string  `json:"time"`
		Type                 int     `json:"type"`
	}
	UserWhiteHouse struct {
		Id         int    `json:"id"`
		UserId     int    `json:"user_id"`
		Weight     int    `json:"weight"`
		CreateTime string `json:"create_time"`
		Comment    string `json:"comment"`
	}
	DownExcel struct {
		Id   int    `json:"id"`
		Uid  int    `json:"uid"`
		Url  string `json:"url"`
		Time string `json:"time"`
	}
)

func (DownExcel) TableName() string {
	return "down_excel"
}

func (UserWhiteHouse) TableName() string {
	return "user_white_house"
}

func (SalesCache) TableName() string {
	return "sales_cache"
}

func (Cache) TableName() string {
	return "cache"
}

func (InfoBlackHouse) TableName() string {
	return "info_black_house"
}

func (LiveChat) TableName() string {
	return "live_chat"
}

func (Config2) TableName() string {
	return "config2"
}

func (RegionBlackHouse) TableName() string {
	return "region_black_house"
}

func (Documents) TableName() string {
	return "documents"
}

func (UserG2code) TableName() string {
	return "user_g2code"
}

func (UserAuditLog) TableName() string {
	return "user_audit_log"
}

func (Token) TableName() string {
	return "token"
}

func (User) TableName() string {
	return "user"
}

func (LotteryDetail) TableName() string {
	return "lottery_detail"
}

func (MoveRewards) TableName() string {
	return "move_rewards"
}

func (ScoreLog) TableName() string {
	return "score_log"
}

func (ScoreDetail) TableName() string {
	return "score_detail"
}

func (ScoreConfig) TableName() string {
	return "score_config2"
}
func (Account) TableName() string {
	return "account"
}
func (ActivityBlackHouse) TableName() string {
	return "activity_black_house"
}
func (ActivityCashVoucherConfig) TableName() string {
	return "activity_cash_voucher_config"
}
func (ActivityDisable) TableName() string {
	return "activity_disable"
}
func (Admin) TableName() string {
	return "admin"
}
func (Activity2023) TableName() string {
	return "activity2023"
}
func (AdminLog) TableName() string {
	return "admin_log"
}
func (Announcement) TableName() string {
	return "announcement"
}
func (AutoWithdrawDetail) TableName() string {
	return "auto_withdraw_detail"
}
func (AutoWithdrawRules) TableName() string {
	return "auto_withdraw_rules"
}
func (Bank) TableName() string {
	return "bank"
}
func (Captcha) TableName() string {
	return "captcha"
}
func (CashVoucher) TableName() string {
	return "cash_voucher"
}
func (Commission) TableName() string {
	return "commission"
}
func (CommissionConfig) TableName() string {
	return "commission_config"
}
func (Config) TableName() string {
	return "config"
}
func (Contest) TableName() string {
	return "contest"
}
func (CouponCode) TableName() string {
	return "coupon_code"
}
func (CouponDetail) TableName() string {
	return "coupon_detail"
}
func (Customer) TableName() string {
	return "customer"
}
func (DemoAccount) TableName() string {
	return "demo_account"
}
func (Edm) TableName() string {
	return "edm"
}
func (EdmDetail) TableName() string {
	return "edm_detail"
}
func (Follow) TableName() string {
	return "follow"
}
func (Fund) TableName() string {
	return "fund"
}
func (InsuranceCoupon) TableName() string {
	return "insurance_coupon"
}
func (Interest) TableName() string {
	return "interest"
}
func (InviteCode) TableName() string {
	return "invite_code"
}
func (Log) TableName() string {
	return "log"
}
func (LotteryConfig) TableName() string {
	return "lottery_config"
}

func (MailLog) TableName() string {
	return "mail_log"
}
func (MailTpl) TableName() string {
	return "mail_tpl"
}
func (MamDetail) TableName() string {
	return "mam_detail"
}
func (MamFund) TableName() string {
	return "mam_fund"
}
func (MamProject) TableName() string {
	return "mam_project"
}
func (Orders) TableName() string {
	return "orders"
}
func (Payment) TableName() string {
	return "payment"
}
func (PaymentLog) TableName() string {
	return "Payment_log"
}
func (PaymentConfig) TableName() string {
	return "payment_config"
}
func (Propaganda) TableName() string {
	return "propaganda"
}
func (PwdBlackHouse) TableName() string {
	return "pwd_black_house"
}
func (Poster) TableName() string {
	return "poster"
}
func (Question) TableName() string {
	return "question"
}
func (QuestionDetail) TableName() string {
	return "question_detail"
}
func (RiskgroupMapping) TableName() string {
	return "riskgroup_mapping"
}
func (SalesCommission) TableName() string {
	return "sales_commission"
}
func (SalesCommissionSet) TableName() string {
	return "sales_commission_set"
}
func (CommissionSetConfig) TableName() string {
	return "commission_set_config"
}
func (SalesTransfer) TableName() string {
	return "sales_transfer"
}

func (Signal) TableName() string {
	return "signal"
}

func (UserAddress) TableName() string {
	return "user_address"
}
func (UserFile) TableName() string {
	return "user_file"
}
func (UserInfo) TableName() string {
	return "user_info"
}
func (UserLog) TableName() string {
	return "user_log"
}
func (UserMessage) TableName() string {
	return "user_message"
}
func (UserMore) TableName() string {
	return "user_more"
}
func (UserAbility) TableName() string {
	return "user_ability"
}
func (UserExamine) TableName() string {
	return "user_examine"
}
func (UserWallet) TableName() string {
	return "user_wallet"
}
func (Version) TableName() string {
	return "version"
}
func (UserVip) TableName() string {
	return "user_vip"
}
func (UserVipCash) TableName() string {
	return "user_vip_cash"
}
func (UserVipConfig) TableName() string {
	return "user_vip_config"
}
func (UserVipFlow) TableName() string {
	return "user_vip_flow"
}
func (Wallet) TableName() string {
	return "wallet"
}
func (Workorder) TableName() string {
	return "workorder"
}
func (Rights) TableName() string {
	return "rights"
}
func (RebateConfig) TableName() string {
	return "rebate_config"
}
func (ExtraRate) TableName() string {
	return "extra_rate"
}
func (MamLog) TableName() string {
	return "mamLog"
}
func (CustomerDetail) TableName() string {
	return "customer_detail"
}
func (Coupon) TableName() string {
	return "coupon"
}
func (MamWallet) TableName() string {
	return "mamWallet"
}
func (UserActivity) TableName() string {
	return "user_activity"
}
func (CreditDetail) TableName() string {
	return "credit_detail"
}
func (UserMessageConfig) TableName() string {
	return "user_message_config"
}
func (Wage) TableName() string {
	return "wage"
}
func (FootballTeam) TableName() string {
	return "football_team"
}
func (FootballMatch) TableName() string {
	return "football_match"
}
func (FootballOrders) TableName() string {
	return "football_orders"
}
func (InviteCodeSet) TableName() string {
	return "invite_code_set"
}
func (CommissionSet) TableName() string {
	return "commission_set"
}
func (PartnerPlan) TableName() string {
	return "partner_plan"
}
func (Ads) TableName() string {
	return "ads"
}
func (CommissionSetCustom) TableName() string {
	return "commission_set_custom"
}

func (Activities) TableName() string {
	return "activities"
}
