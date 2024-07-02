package abc

func GetRiskgroupMapping(where string) (data RiskgroupMapping) {
	db.Debug().Where(where).Order("id desc").First(data)
	return data
}
