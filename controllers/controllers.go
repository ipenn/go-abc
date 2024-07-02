package controllers

import (
	"context"
	"fmt"
	"github.com/chenqgp/abc/conf"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	file2 "github.com/chenqgp/abc/third/uFile"

	"github.com/chenqgp/abc"
	"github.com/gin-gonic/gin"

	"github.com/go-redis/redis"
)

type R struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

var codes = []int{
	0,   // 通用错误
	1,   // 正常返回
	401, // 通证错误
	402, // 权限错乱
	101, // 账户密码错误
}

func (r R) Response() R {
	for _, code := range codes {
		if code == r.Status {
			return r
		}
	}
	r.Msg = "invalid code ff00000198 from system"
	return r
}

type ResponseLimit struct {
	Status          int         `json:"status"`
	Msg             string      `json:"msg"`
	Data            interface{} `json:"data"`
	HasPreviousPage bool        `json:"hasPreviousPage"`
	HasNextPage     bool        `json:"hasNextPage"`
	PreviousPage    int         `json:"previousPage"`
	NextPage        int         `json:"nextPage"`
	Pages           float64     `json:"pages"`
	Count           int64         `json:"count"`
	Limit           int         `json:"limit"`
	PageNum         int         `json:"pageNum"`
}

func (r ResponseLimit) Response(page, size int, count int64) ResponseLimit {
	r.Pages = math.Ceil(float64(count) / float64(size))
	r.HasNextPage = true
	r.NextPage = page + 1
	if r.Pages-float64(page) <= 0 {
		r.HasNextPage = false
		r.NextPage = page
	}

	r.HasPreviousPage = false
	r.PreviousPage = 1
	if page > 1 {
		r.HasPreviousPage = true
		r.PreviousPage = page - 1
	}
	r.Count = count
	r.Limit = size
	r.PageNum = page
	for _, code := range codes {
		if code == r.Status {
			return r
		}
	}
	r.Msg = "invalid code ff00000198 from system"
	return r
}

func init() {
	abc.RDB = redis.NewClient(&redis.Options{
		Addr:         conf.RdbAddr,
		Password:     conf.RdbPwd,
		DB:           15,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  3 * time.Second,
	})
	abc.Rctx = context.Background()
	pong, err := abc.RDB.Ping(abc.Rctx).Result()
	if err != nil {
		fmt.Println(pong, err)
	}

	bootRDB14()

	//go func() {
	//	tickerSecond := time.NewTicker(5 * time.Second)
	//	for {
	//		select {
	//		case <-tickerSecond.C:
	//			rebate.Rebate()
	//			tickerSecond.Reset(5 * time.Second)
	//		}
	//	}
	//}()
}

func bootRDB14() {
	abc.RDB14 = redis.NewClient(&redis.Options{
		Addr:         conf.RdbAddr,
		Password:     conf.RdbPwd,
		DB:           14,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  3 * time.Second,
	})
	abc.Rctx14 = context.Background()
	pong, err := abc.RDB14.Ping(abc.Rctx14).Result()
	if err != nil {
		fmt.Println(pong, err)
	}
}

func HandleFiles(c *gin.Context, dir, formName string) []string {
	var relaFiles []string
	form, err := c.MultipartForm()
	if err != nil {
		fmt.Println(">>> MultipartForm: ", err.Error())
	}
	files := form.File[formName]
	fileName := ""
	fmt.Println("files len: ", len(files))
	for _, file := range files {
		if file != nil {
			fmt.Println("files size: ", file.Size/1024)
			times := strconv.FormatInt(time.Now().Unix(), 10)
			random := abc.ToString(abc.RandonNumber(6))
			// Generating unique name.
			fileExt := random + times + "." + strings.Split(file.Filename, ".")[1]
			fileName = dir + "/" + fileExt
			// Save file to spec-location from temporary dir.
			err := c.SaveUploadedFile(file, fileName)
			if err != nil {
				fmt.Println("SaveUploadedFile: ", fileName, err)
				continue
			}
			// Implemented functions from different package. So the directory shouldn't
			// pass into other package.
			pic, err := os.Stat(fileName)
			if err != nil {
				fmt.Println("os.Stat: ", fileName, err)
				continue
			}
			_, err = abc.PicCompress(pic, fileName)
			if err != nil {
				fmt.Println("controller HandleFiles ", fileName, err)
				// NOTICE!!
				// Perhaps is a harmful software or script!!!
				os.Remove(fileName)
				continue
			}
			relaFiles = append(relaFiles, fileExt)
		}
	}
	return relaFiles
}
func HandleFilesNoPic(c *gin.Context, dir, formName string) []string {
	var relaFiles []string
	form, _ := c.MultipartForm()
	files := form.File[formName]
	fileName := ""
	for _, file := range files {
		if file != nil {
			times := strconv.FormatInt(time.Now().Unix(), 10)
			random := abc.ToString(abc.RandonNumber(6))
			// Generating unique name.
			fileExt := random + times + "." + strings.Split(file.Filename, ".")[1]
			fileName = dir + "/" + fileExt
			// Save file to spec-location from temporary dir.
			c.SaveUploadedFile(file, fileName)
			// Implemented functions from different package. So the directory shouldn't
			// pass into other package.
			//pic, _ := os.Stat(fileName)
			//_, err := abc.PicCompress(pic, fileName)
			//if err != nil {
			//	fmt.Println("controller HandleFiles ", fileName, err)
			//	// NOTICE!!
			//	// Perhaps is a harmful software or script!!!
			//	os.Remove(fileName)
			//	continue
			//}

			relaFiles = append(relaFiles, fileExt)
		}
	}
	return relaFiles
}

func HandleFilesAllFiles(c *gin.Context, uid int, dir, formName string) []string {
	var relaFiles []string
	form, _ := c.MultipartForm()
	files := form.File[formName]
	fileName := ""
	for _, file := range files {
		if file != nil {
			times := strconv.FormatInt(time.Now().Unix(), 10)
			random := abc.ToString(abc.RandonNumber(6))
			// Generating unique name.
			suffix := strings.Split(file.Filename, ".")[1]
			fileExt := random + times + "." + suffix
			fileName = dir + "/" + fileExt
			// Save file to spec-location from temporary dir.
			c.SaveUploadedFile(file, fileName)
			// Implemented functions from different package. So the directory shouldn't
			// pass into other package.
			if suffix == "jpg" || suffix == "jpeg" || suffix == "png" || suffix == "gif" {
				pic, _ := os.Stat(fileName)
				_, err := abc.PicCompress(pic, fileName)
				if err != nil {
					fmt.Println("controller HandleFiles ", fileName, err)
					// NOTICE!!
					// Perhaps is a harmful software or script!!!
					os.Remove(fileName)
					continue
				}
			}
			relaFiles = append(relaFiles, file2.UploadFile(fileExt, uid))
		}
	}
	return relaFiles
}
