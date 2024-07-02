package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/chenqgp/abc"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/ws"
	"github.com/gin-gonic/gin"
)

func Test1(c *gin.Context) {
	r := R{}
	uid := abc.ToString(c.MustGet("uid"))
	name := c.PostForm("name")
	log.Println(uid, name)
	if c, ok := ws.SingleBoardcast(abc.ToInt(uid)); ok {
		c.Send([]byte(name+`:{"hello world `+uid+`"}`))
	}
	ws.Send([]byte(name+`:{"hello world"}`))
	//ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	//if !ok {
	//	r.Msg = "frequently operated"
	//	c.JSON(http.StatusOK, r.Response())
	//	return
	//}
	//defer done()

	//for i := 0; i < 3; i++ {
	//	time.Sleep(1 * time.Second)
	//	done := abc.LimiterPer(nonConcurrent.Queue, uid)
	//	log.Println("12412412443414219437129074901274190241279012")

	//	time.Sleep(4 * time.Second)

	//	done()

	//}

	//go tron.CreateUserWallet(abc.LimiterPer(wallet.Queue, uid))
	//done := abc.LimiterPer(wallet.Queue, uid)
	//done()

	r.Status = 1
	c.JSON(http.StatusOK, r.Response())
}

func Test2(c *gin.Context) {
	r := R{}
	uid := abc.ToInt(c.MustGet("uid"))
	name := c.PostForm("name")
	page := c.PostForm("page")
	size := c.PostForm("size")
	log.Println(uid, name, page, size)
	//ws.SingleBoardcast(uid).Send([]byte(`reader:{"hello world"}`))
	//ws.Send([]byte(`reader:{"hello world"}`))
	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = "frequently operated"
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()
	time.Sleep(5 * time.Second)

	r.Status = 1
	c.JSON(http.StatusAccepted, r.Response())
}
