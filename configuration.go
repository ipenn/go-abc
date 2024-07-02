package abc

func GetConf(id any) Config {
	var conf Config
	db.Debug().Where("id = ?", id).First(&conf)
	return conf
}
