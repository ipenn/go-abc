package abc

func GetRebate(where string, params ...interface{}) RebateConfig {
	var r RebateConfig
	db.Debug().Where(where, params).First(&r)

	return r
}
