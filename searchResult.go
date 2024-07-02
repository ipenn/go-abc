package abc

type (
	ScoreCountDetail struct {
		Total float64 `json:"total"`
		Used  float64 `json:"used"`
		Using float64 `json:"using"`
	}

	InviteInfo struct {
		TrueName      string `json:"true_name"`
		UserType      string `json:"user_type"`
		IbNo          string `json:"ib_no"`
		Agreement     string `json:"agreement"`
		AgreementFee  string `json:"agreement_fee"`
		AgreementTime string `json:"agreement_time"`
	}
	InviteCount struct {
		STD    int `json:"std"`
		DMA    int `json:"dma"`
		Level1 int `json:"Level Ⅰ"`
		Level2 int `json:"Level Ⅱ"`
		Level3 int `json:"Level Ⅲ"`
	}
	DirectInvitation struct {
		Id         int    `json:"id"`
		UserType   string `json:"user_Type"`
		CreateTime string `json:"create_time"`
		TrueName   string `json:"true_name"`
		Email      string `json:"email"`
		Mobile     string `json:"mobile"`
		AuthStatus int    `json:"auth_status"`
		Status     int    `json:"status"`
		UserStatus int    `json:"user_status"`
		GroupType  string `json:"group_type"`
		Count      int    `json:"count"`
	}
	UserIdAndUserType struct {
		Id       int    `json:"id"`
		UserType string `json:"user_type"`
	}
	InterestData struct {
		Total     float64 `json:"total"`
		Yesterday float64 `json:"yesterday"`
		ExtraRate float64 `json:"extra_rate"`
	}
	DateData struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	}
	LotteryConfigOne struct {
		Id     int     `json:"id"`
		Name   string  `json:"name"`
		Type   string  `json:"type"`
		Value  float64 `json:"value"`
		Chance float64 `json:"chance"`
	}
	CouponReturn struct {
		Id         int     `json:"id"`
		Type       int     `json:"type"` //2moveRewards 3userVipCash 4birthday 5cashVoucher
		CouponId   string  `json:"coupon_id"`
		Amount     float64 `json:"amount"`
		Status     int     `json:"status"` //0 not use 1using 2 used -1 exceed
		CreateTime string  `json:"create_time"`
		EndTime    string  `json:"end_time"`
	}
	OrderSpent struct {
		Id        int     `json:"id"`
		CloseTime string  `json:"close_time"`
		Volume    float64 `json:"volume"`
		Spent     float64 `json:"spent"`
	}
	CreditDetailReturn struct {
		Id           int     `json:"id"`
		UserId       int     `json:"user_id"`
		Login        int     `json:"login"`
		CreateTime   string  `json:"create_time"`
		OverTime     string  `json:"over_time"`
		DeductTime   string  `json:"deduct_time"`
		Amount       float64 `json:"amount"`
		Source       int     `json:"source"`
		Status       int     `json:"status"`
		Comment      string  `json:"comment"`
		Volume       float64 `json:"volume"`
		UsableVolume float64 `json:"usable_volume"`
		CouponNo     string  `json:"coupon_no"`
		Deposit      float64 `json:"deposit"`
	}
	LotterySimple struct {
		Id    int     `json:"id"`
		Name  string  `json:"name"`
		Type  string  `json:"type"`
		Value float64 `json:"value"`
	}
	kv struct {
		K int `json:"k"`
		V int `json:"v"`
	}
	Address struct {
		TrueName string `json:"true_name"`
		Area     string `json:"area"`
		Phone    string `json:"phone"`
		Zip      string `json:"zip"`
		Address  string `json:"address"`
	}
	ScoreConfigReturn struct {
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
		Thumb     string  `json:"thumb"`
		Surplus   int     `json:"surplus"`
		Weight    int     `json:"weight"`
	}
	ActivityCashVoucherList struct {
		CashNo     string  `json:"cash_no"`
		CreateTime string  `json:"create_time"`
		EndTime    string  `json:"end_time"`
		TrueName   string  `json:"true_name"`
		Email      string  `json:"email"`
		Amount     float64 `json:"amount"`
		Volume     float64 `json:"volume"`
		Status     int     `json:"status"`
	}
	AmountInfo struct {
		Volume float64 `json:"volume"`
		Amount float64 `json:"amount"`
		Fee    float64 `json:"fee"`
	}
	PaymentCount struct {
		Total   int `json:"total"`
		Success int `json:"success"`
		Pending int `json:"pending"`
	}
	WageReturn struct {
		Id             int     `json:"id"`
		UserId         int     `json:"user_id"`
		CreateTime     string  `json:"create_time"`
		Status         int     `json:"status"`
		Type           string  `json:"type"`
		Amount         float64 `json:"amount"`
		CommissionType int     `json:"commission_type"`
	}
	PositionData struct {
		Balance any          `json:"balance"`
		Code    any          `json:"code"`
		Count   any          `json:"count"`
		Equity  any          `json:"equity"`
		Lots    any          `json:"lots"`
		Margin  any          `json:"margin"`
		Profit  any          `json:"profit"`
		List    []OrdersData `json:"list"`
		Pedding []OrdersData `json:"pedding"`
	}
	OrdersData struct {
		Account    any `json:"account"`
		Cmd        any `json:"cmd"`
		Comment    any `json:"comment"`
		Commission any `json:"commission"`
		Login      any `json:"login"`
		OpenPrice  any `json:"open_price"`
		ClosePrice any `json:"close_price"`
		OpenTime   any `json:"open_time"`
		CloseTime  any `json:"close_time"`
		OrderId    any `json:"order_id"`
		Profit     any `json:"profit"`
		Sl         any `json:"sl"`
		Storage    any `json:"storage"`
		Symbol     any `json:"symbol"`
		Tp         any `json:"tp"`
		Volume     any `json:"volume"`
	}
	AccountInfoData struct {
		Balance     any `json:"balance"`
		Code        int `json:"code"`
		Credit      any `json:"credit"`
		Equity      any `json:"equity"`
		LevelType   any `json:"level_type"`
		Leverage    any `json:"leverage"`
		Margin      any `json:"margin"`
		MarginFree  any `json:"margin_free"`
		MarginLevel any `json:"margin_level"`
		Volume      any `json:"volume"`
	}
	OrdersSimple struct {
		OrderId    int     `json:"order_id"`
		Login      int     `json:"login"`
		UserId     string     `json:"user_id"`
		TrueName   string  `json:"true_name"`
		Symbol     string  `json:"symbol"`
		SymbolType int     `json:"symbol_type"`
		Cmd        int     `json:"cmd"`
		Volume     float64 `json:"volume"`
		OpenPrice  float64 `json:"open_price"`
		OpenTime   string  `json:"open_time"`
		ClosePrice float64 `json:"close_price"`
		CloseTime  string  `json:"close_time"`
		Sl         float64 `json:"sl"`
		Tp         float64 `json:"tp"`
		Profit     float64 `json:"profit"`
		Storage    float64 `json:"storage"`
		Commission float64 `json:"commission"`
		Taxes      float64 `json:"taxes"`
		Comment    string  `json:"comment"`
	}
	PaymentSimple struct {
		Id          int     `json:"id"`
		OrderNo     string  `json:"order_no"`
		UserId      int     `json:"user_id"`
		TrueName    string  `json:"true_name"`
		Email       string  `json:"email"`
		Phonectcode string  `json:"phonectcode"`
		Mobile      string  `json:"mobile"`
		UserType    string  `json:"user_type"`
		CreateTime  string  `json:"create_time"`
		Amount      float64 `json:"amount"`
		PayName     string  `json:"pay_name"`
		PayTime     string  `json:"pay_time"`
		Comment     string  `json:"comment"`
		Status      int     `json:"status"`
		Type        string  `json:"type"`
		PayFee      float64 `json:"pay_fee"`
		PayCompanyAmount float64 `json:"pay_company_amount"`
	}
	UserGrade struct {
		UserId int `json:"user_id"`
		Grade  int `json:"grade"`
	}
	RedisListData struct {
		List  []byte `json:"list"`
		Count int64    `json:"count"`
		Total Total  `json:"total"`
	}
	Total struct {
		Amount     float64 `json:"amount"`
		Volume     float64 `json:"volume"`
		Profit     float64 `json:"profit"`
		Storage    float64 `json:"storage"`
		Commission float64 `json:"commission"`
	}
	AccountInformation struct {
		Nationality    string `json:"nationality" validate:"required"`
		IdentityType   string `json:"identity_type" validate:"required"`
		Identity       string `json:"identity" validate:"required"`
		Title          string `json:"title" validate:"required"`
		Birthday       string `json:"birthday" validate:"required"`
		Birthcountry   string `json:"birthcountry" validate:"required"`
		Address        string `json:"address" validate:"required"`
		Country        string `json:"country"`
		AddressDate    string `json:"address_date"`
		Type           int    `json:"type" validate:"required,gt=0"`
		OldPhonectcode string `json:"old_phonectcode"`
		OldMobile      string `json:"old_mobile"`
	}
	Transaction struct {
		CurrencyType string `json:"currency_type" validate:"required"`
		Platform     string `json:"platform" validate:"required"`
		Forexp       string `json:"forexp" validate:"required"`
		Investfreq   string `json:"investfreq" validate:"required"`
		Otherexp     string `json:"otherexp" validate:"required"`
		AccountType  string `json:"account_type" validate:"required"`
		Investaim    string `json:"investaim" validate:"required"`
		Type         int    `json:"type" validate:"required,gt=0"`
	}
	FinancialInformation struct {
		Incomesource string `json:"incomesource" validate:"required"`
		Employment   string `json:"employment" validate:"required"`
		Business     string `json:"business" validate:"required"`
		Position     string `json:"position"`
		Isusa        string `json:"isusa" validate:"required"`
		Isforusa     string `json:"isforusa" validate:"required"`
		Isearnusa    string `json:"isearnusa" validate:"required"`
		Istax        string `json:"istax" validate:"required"`
		Ispolitic    string `json:"ispolitic" validate:"required"`
		Income       string `json:"income" validate:"required"`
		Netasset     string `json:"netasset" validate:"required"`
		Company      string `json:"company"`
		Type         int    `json:"type" validate:"required,gt=0"`
	}
	IdDocument struct {
		Surname  string `json:"surname" validate:"required"`
		Lastname string `json:"lastname" validate:"required"`
		IdType   int    `json:"id_type" validate:"required,gt=0"`
		IdFront  string `json:"id_front"`
		IdBack   string `json:"id_back"`
		Other    string `json:"other"`
		Type     int    `json:"type" validate:"required,gt=0"`
	}
	BankInformation struct {
		BankType        int    `json:"bank_type" validate:"required,gt=0"`
		Name            string `json:"name" validate:"required"`
		BankName        string `json:"bank_name" validate:"required"`
		BankNo          string `json:"bank_no" validate:"required"`
		BankAddress     string `json:"bank_address" validate:"required"`
		BankFile        string `json:"bank_file"`
		Swift           string `json:"swift"`
		Iban            string `json:"iban"`
		Area            int    `json:"area"`
		BankCardType    int    `json:"bank_card_type"`
		ChineseIdentity string `json:"chinese_identity" validate:"required"`
		Type            int    `json:"type" validate:"required,gt=0"`
	}

	PaymentSimple2 struct {
		Id            int     `json:"id"`
		OrderNo       string  `json:"order_no"`
		UserId        int     `json:"user_id"`
		CreateTime    string  `json:"create_time"`
		Amount        float64 `json:"amount"`
		PayName       string  `json:"pay_name"`
		PayTime       string  `json:"pay_time"`
		Status        int     `json:"status"`
		Type          string  `json:"type"`
		Intro         string  `json:"intro"`
		TransferLogin int     `json:"transfer_login"`
		PayFee        float64 `json:"pay_fee"`
		AStatus       float64 `json:"a_status"`
		BStatus       float64 `json:"b_status"`
		CStatus       float64 `json:"c_status"`
		WireDoc       string  `json:"wire_doc"`
		PayUrl        string  `json:"pay_url"`
		WaterNumber   string  `json:"water_number"`
	}
	UserListData struct {
		Inviter         string `json:"Inviter"`
		AuthStatus      string `json:"auth_status"`
		Balance         string `json:"balance"`
		CreateTime      string `json:"create_time"`
		Deposit         string `json:"deposit"`
		Dma             string `json:"dma"`
		Email           string `json:"email"`
		Equity          string `json:"equity"`
		Forex           string `json:"forex"`
		Grade           string `json:"grade"`
		GroupName       string `json:"group_name"`
		Id              string `json:"id"`
		Metal           string `json:"metal"`
		Mobile          string `json:"mobile"`
		OldMobile       string `json:"old_mobile"`
		OldPhonectcode  string `json:"old_phonectcode"`
		Phonectcode     string `json:"phonectcode"`
		Silver          string `json:"silver"`
		Status          string `json:"status"`
		StockCommission string `json:"stockCommission"`
		TrueName        string `json:"true_name"`
		UserStatus      int    `json:"user_status"`
		UserType        string `json:"user_type"`
		WalletIn        string `json:"walletIn"`
		WalletOut       string `json:"walletOut"`
		WalletBalance   string `json:"wallet_balance"`
		Withdraw        string `json:"withdraw"`
	}
	UserSimple struct {
		Id        int    `json:"id"`
		TrueName  string `json:"true_name"`
		UserType  string `json:"user_type"`
		SalesType string `json:"sales_type"`
	}
	Export struct {
		Name         string      `json:"name"`
		Role         string      `json:"role"`
		SuperiorName string      `json:"superior_name"`
		My           interface{} `json:"my"`
		Customer     interface{} `json:"customer"`
		ParentId     int         `json:"parent_id"`
		Path         string      `json:"path"`
	}
	CustomExport struct {
		Name         string      `json:"name"`
		Role         string      `json:"role"`
		SuperiorName string      `json:"superior_name"`
		My           interface{} `json:"my"`
	}
)
