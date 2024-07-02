package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chenqgp/abc"
	golbal "github.com/chenqgp/abc/global"
	nonConcurrent "github.com/chenqgp/abc/task/task-nonConcurrent"
	"github.com/chenqgp/abc/third/telegram"
	"github.com/gin-gonic/gin"
	"net/http"
)

// QuestionTelegram 工单发送telegram消息
func QuestionTelegram(userType, teleData, filePath string) {
	var r []string
	json.Unmarshal([]byte(filePath), &r)
	tid := 13
	if userType == "user" {
		tid = 14
	}
	telegram.SendMsg(telegram.TEXT, tid, teleData)
	for _, item := range r {
		telegram.SendMsg(telegram.PHOTO, tid, item)
	}
}

func AddQuestion(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	title := c.PostForm("title")
	content := c.PostForm("content")
	tag := c.PostForm("tag")
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	if title == "" || content == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	tag = abc.SwitchTag(tag)
	if tag == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	question := abc.Question{}
	r.Status, question = abc.CreateQuestion(uid, title, content, tag)
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	image := HandleFilesAllFiles(c, uid, "upload", "file")
	files := ""
	if len(image) != 0 {
		files = string(abc.ToJSON(image))
	}

	abc.CreateQuestionDetail(question.Id, content, files)

	user := abc.GetUserById(uid)
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	abc.WriteUserLog(uid, "Post Question", c.ClientIP(), question.Title)

	QuestionTelegram(user.UserType, fmt.Sprintf("%d > %s > %s > %s > 新站内信:%s > %s",
		question.Id, user.TrueName, user.Email, question.Tag, question.Title, question.Content), files)

	c.JSON(http.StatusOK, r.Response())
}

func ReplyQuestion(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	id := abc.ToInt(c.PostForm("id"))
	content := c.PostForm("content")
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	if content == "" {
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	question := abc.GetQuestion(id)
	question.Tag = abc.ReplaceTag(question.Tag)
	if question.Id == 0 || question.UserId != uid {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	} else if question.Status == 2 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10109]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	image := HandleFilesAllFiles(c, uid, "upload", "file")
	files := ""
	if len(image) != 0 {
		files = string(abc.ToJSON(image))
	}

	r.Status = abc.CreateQuestionDetail(question.Id, content, files)
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusInternalServerError, r.Response())
		return
	}

	user := abc.GetUserById(uid)
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	QuestionTelegram(user.UserType, fmt.Sprintf("%d > %s > %s > %s > 回复:%s ",
		question.Id, user.TrueName, user.Email, question.Title, content), files)

	c.JSON(http.StatusOK, r.Response())
}

func ClosedQuestion(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	id := abc.ToInt(c.PostForm("id"))
	uid := abc.ToInt(c.MustGet("uid"))

	r := R{}

	ok, done := abc.LimiterWait(nonConcurrent.Queue, uid)
	if !ok {
		r.Msg = golbal.Wrong[language][10119]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	defer done()

	user := abc.GetUser(fmt.Sprintf("id=%d", uid))
	if user.AuthStatus != 1 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10027]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	question := abc.GetQuestion(id)
	if question.Id == 0 || question.UserId != uid {
		r.Status, r.Msg = 0, golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	} else if question.Status == 2 {
		r.Status, r.Msg = 0, golbal.Wrong[language][10109]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	question.Status = 2
	question.UpdateTime = abc.FormatNow()
	r.Status = abc.SaveQuestion(question)
	if r.Status == 0 {
		r.Msg = golbal.Wrong[language][10100]
		c.JSON(http.StatusOK, r.Response())
		return
	}

	c.JSON(http.StatusOK, r.Response())

	abc.WriteUserLog(uid, "Closed Question", c.ClientIP(), question.Title)
}

func QuestionList(c *gin.Context) {
	language := abc.ToString(c.MustGet("language"))
	uid := abc.ToInt(c.MustGet("uid"))
	status := abc.ToInt(c.PostForm("status"))
	page := abc.ToInt(c.PostForm("page"))
	size := abc.ToInt(c.PostForm("size"))
	if page <= 0 || size <= 0 {
		r := R{}
		r.Msg = golbal.Wrong[language][10000]
		c.JSON(http.StatusOK, r.Response())
		return
	}
	r := ResponseLimit{}
	r.Status, r.Msg, r.Count, r.Data = abc.QuestionList(uid, status, page, size)
	c.JSON(http.StatusOK, r.Response(page, size, r.Count))
}

func QuestionDetail(c *gin.Context) {
	//language := abc.ToString(c.MustGet("language"))
	//uid := abc.ToInt(c.MustGet("uid"))
	id := abc.ToInt(c.PostForm("id"))

	r := R{}
	question := abc.GetQuestion(id)
	question.Tag = abc.ReplaceTag(question.Tag)
	questionDetails := struct {
		Question abc.Question         `json:"question"`
		Detail   []abc.QuestionDetail `json:"detail"`
	}{
		Question: question,
		Detail:   abc.GetQuestionDetail(id),
	}
	r.Data, r.Status, r.Msg = questionDetails, 1, ""
	c.JSON(http.StatusOK, r.Response())
	abc.ReadQuestion(id)
}
