package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PiyushhBhutoria/botonwhatsapp/config"
	"github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
)

var (
	requestChannel chan whatsapp.TextMessage
	wac, _         = whatsapp.NewConn(20 * time.Second)
	now            = time.Now().Unix()
	sapEndpoint    = "https://api.cai.tools.sap/build/v1/dialog"
)

func init() {
	config.Init("config")
	config := config.GetConfig()

	fmt.Println("for session issue  remove whatsapp.gob from --> " + os.TempDir())
	fmt.Println("running on " + strconv.Itoa(runtime.NumCPU()) + " cores.")

	raven.SetDSN(config.GetString("sentry"))

	requestChannel = make(chan whatsapp.TextMessage, runtime.NumCPU())

	wac.SetClientVersion(2, 2021, 4)
	wac.AddHandler(&waHandler{wac})
	if err := login(wac); err != nil {
		panic("Error logging in: \n" + err.Error())
	}

	<-time.After(3 * time.Second)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var temp resp

	go func() {
		for {
			request, ok := <-requestChannel
			if ok {
				payload, err := createPayload(Payload{
					Message: payloadData{
						Content:        request.Text,
						Type:           "text",
						ConversationID: request.Info.RemoteJid[2:12],
					},
				})
				if err != nil {
					log.Printf("Error creating payload: %v\n", err)
				}
				data, err := postRequest(sapEndpoint, payload)
				if err != nil {
					log.Printf("Error in post request: %v\n", err)
				}

				err = json.Unmarshal([]byte(data), &temp)
				if err != nil {
					log.Printf("Error decoding body: %v\n", err)
				}

				if len(temp.Results.Messages) > 0 && temp.Results.Messages[0].Content != "I trigger the fallback skill because I don't understand or I don't know what I'm supposed to do..." {
					fmt.Println(temp.Results.Messages[0].Content)
					to := request.Info.RemoteJid[2:12]
					mess := temp.Results.Messages[0].Content
					log.Printf("%v --> %s\nBot --> %v", to, request.Text, mess)
					log.Printf("-------------------------------")
					fmt.Println(texting(to, mess))
				}
			}
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] \"%s %s %s %d %s %s\"\n",
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))
	router.Use(sentry.Recovery(raven.DefaultClient, true))
	router.Use(cors.Default())

	router.GET("/", helloworld)
	router.GET("/ping", ping)
	router.GET("/sendText", sendText)
	router.POST("/sendBulk", sendBulk)

	if err := router.Run(":8080"); err != nil {
		log.Printf("Shutdown with error: %v\n", err)
	}

}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func helloworld(c *gin.Context) {
	c.String(http.StatusOK, "The bot is running")
}

func sendText(c *gin.Context) {
	to := strings.Replace(c.DefaultQuery("to", "1234567890"), " ", "", -1)
	mess := c.DefaultQuery("msg", "testing")
	c.String(http.StatusOK, texting(to, mess))
}

func sendBulk(c *gin.Context) {
	var data sendBulkText
	m := make(map[string]string)

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, each := range data.List {
		each.Receiver = strings.Replace(each.Receiver, " ", "", -1)
		if each.Receiver != "" {
			m[each.Receiver] = texting(each.Receiver, each.Message)
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": data})
}
