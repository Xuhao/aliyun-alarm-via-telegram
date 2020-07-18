package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

type AlterPayload struct {
	AlterName    string `form:"alertName" json:"alertName" binding:"required"`
	CurValue     string `form:"curValue" binding:"required"`
	Expression   string `form:"expression" binding:"required"`
	InstanceName string `form:"instanceName" binding:"required"`
	Timestamp    int64  `form:"timestamp" binding:"required"`
	MetricName   string `form:"metricName" binding:"required"`
}

func handleErr(err error, c *gin.Context) {
	if err != nil {
		fmt.Println(err)
		sentry.CaptureException(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "failed", "message": err.Error()})
	}
}

var divider string = "--------------------------------------------------------------------------------------------------"

// Assembly post body
func assembly(c *gin.Context) (bodyStr string, err error) {
	var payload AlterPayload
	if err = c.ShouldBind(&payload); err != nil {
		return
	}
	chatID := c.Query("chat_id")
	title := "ECS服务器报警!"
	if c.Query("product") == "lb" {
		title = "负载均衡器健康检查报警!"
	}

	textArr := []string{
		fmt.Sprintf("<b>%s</b> - %s", title, time.Unix(payload.Timestamp/1000, 0).Format("2006-01-02 15:04:05")),
		fmt.Sprintf("实例：%s", payload.InstanceName),
		fmt.Sprintf("指标名称：%s", payload.MetricName),
		fmt.Sprintf("当前值：%s", payload.CurValue),
		fmt.Sprintf("触发条件：%s", payload.Expression),
	}

	bodyStr = fmt.Sprintf("parse_mode=html&disable_web_page_preview=true&chat_id=%s&text=%s", chatID, strings.Join(textArr[:], "%0A"))
	return
}

func sendMessage(c *gin.Context) {
	var bodyStr string
	var err error
	if messageType := c.Query("text"); messageType == "assembly" {
		bodyStr, err = assembly(c)
		if err != nil {
			handleErr(err, c)
			return
		}
	} else {
		rawBody, err := c.GetRawData()
		if err != nil {
			handleErr(err, c)
			return
		}
		bodyStr = string(rawBody)
	}

	bot := c.Param("bot")
	query := c.Request.URL.RawQuery
	contentType := c.Request.Header.Get("Content-Type")

	if bodyStr != "" {
		contentType = "application/x-www-form-urlencoded"
		query = ""
	}
	url := fmt.Sprintf("https://api.telegram.org/%s/sendMessage?%s", bot, query)

	fmt.Printf("\n%s\n", divider)
	fmt.Printf("%s 开始转发 💪💪💪\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("[request url]:", url)
	fmt.Println("[request body]:", bodyStr)
	// fmt.Println("[original body]:", originalBody)
	fmt.Println("[Content-Type]:", contentType)

	response, err := http.Post(url, contentType, strings.NewReader(bodyStr))
	if err != nil {
		handleErr(err, c)
		fmt.Println("err:", err.Error())
		fmt.Printf("%s\n\n", divider)
		return
	}

	defer response.Body.Close()
	fmt.Println("[response status]:", response.StatusCode)

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		handleErr(err, c)
		fmt.Println("err:", err.Error())
		fmt.Printf("%s\n\n", divider)
		return
	}

	fmt.Println("[response body]:", string(data))
	fmt.Printf("%s\n\n", divider)
	status := "failed"
	if response.StatusCode == http.StatusOK {
		status = "success"
	}

	c.JSON(http.StatusOK, gin.H{"status": status, "message": response.StatusCode})
}

func main() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:     "http://2b09a75c705541e1a0a156fffcfd5e39@localhost:9000/1",
		Debug:   true,
		Release: "v1.0.0",
	}); err != nil {
		log.Fatalf("Sentry.Init: %s", err)
	}
	defer sentry.Flush(time.Second * 2)

	app := gin.Default()

	app.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	app.Use(func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetTag("version", "v1.0.0")
		}
		ctx.Next()
	})

	app.POST("/:bot/sendMessage", sendMessage)
	app.GET("/:bot/sendMessage", sendMessage)
	app.Run()
}
