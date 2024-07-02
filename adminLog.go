package abc

func (adminLog AdminLog) CreateAdminLog() {
	db.Debug().Create(&adminLog)
}
