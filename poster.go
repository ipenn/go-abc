package abc

func FindPoster(where string,arg ...any)  (data []Poster){
	db.Debug().Where(where,arg...).Order("create_time desc").Find(&data)
	return
}
