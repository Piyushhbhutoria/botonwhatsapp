package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
)

var (
	requestChannel chan *events.Message
	now            = time.Now()
	cli            *whatsmeow.Client
	wLog           waLog.Logger
	zlog           *zap.SugaredLogger
	storeContainer *sqlstore.Container
	logLevel       = "INFO"
)

func init() {
	debugLogs := flag.Bool("debug", false, "Enable debug logs?")
	dbDialect := flag.String("db-dialect", "sqlite3", "Database dialect (sqlite3 or postgres)")
	dbAddress := flag.String("db-address", "file:db/examplestore.db?_foreign_keys=on", "Database address")
	flag.Parse()

	if *debugLogs {
		logLevel = "DEBUG"
	}
	wLog = waLog.Stdout("Main", logLevel, true)

	dbLog := waLog.Stdout("Database", logLevel, true)
	var err error
	storeContainer, err = sqlstore.New(*dbDialect, *dbAddress, dbLog)
	if err != nil {
		wLog.Errorf("Failed to connect to database: %v", err)
		return
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"bot.log"},
		ErrorOutputPaths: []string{"boterr.log"},
	}
	logger, _ := cfg.Build()
	if err != nil {
		panic(err)
	}
	logger.WithOptions(zap.AddCallerSkip(1))
	defer logger.Sync() // flushes buffer, if any
	zlog = logger.Sugar()

	err = godotenv.Load(".env")
	if err != nil {
		zlog.Fatalf("Some error occured. Err: %s", err)
	}
}

func main() {
	fmt.Println("loaded env ", os.Getenv("SAP")) // remove this line
	waBinary.IndentXML = true

	device, err := storeContainer.GetFirstDevice()
	if err != nil {
		wLog.Errorf("Failed to get device: %v", err)
		return
	}

	cli = whatsmeow.NewClient(device, waLog.Stdout("Client", logLevel, true))

	ch, err := cli.GetQRChannel(context.Background())
	if err != nil {
		// This error means that we're already logged in, so ignore it.
		if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			wLog.Errorf("Failed to get QR channel: %v", err)
		}
	} else {
		go func() {
			for evt := range ch {
				if evt.Event == "code" {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				} else {
					wLog.Infof("QR channel result: %s", evt.Event)
				}
			}
		}()
	}

	cli.AddEventHandler(handler)
	err = cli.Connect()
	if err != nil {
		wLog.Errorf("Failed to connect: %v", err)
		return
	}

	go func() {
		for {
			evt, ok := <-requestChannel
			if ok {
				temp, err := askBot(evt)
				if err != nil {
					zlog.Infof("Error in post request: %v\n", err)
				}

				if len(temp.Results.Messages) > 0 && temp.Results.Messages[0].Content != "I trigger the fallback skill because I don't understand or I don't know what I'm supposed to do..." {
					mess := temp.Results.Messages[0].Content
					fmt.Println(mess)
					to := evt.Info.Sender.User
					zlog.Infof("%v --> %s\nBot --> %v", to, evt.Message.GetConversation(), mess)
					zlog.Infof("-------------------------------")
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
	router.Use(cors.Default())

	router.GET("/", helloworld)
	router.GET("/ping", ping)
	router.GET("/sendText", sendText)
	router.POST("/sendBulk", sendBulk)

	if err := router.Run(":8080"); err != nil {
		wLog.Infof("Shutdown with error: %v\n", err)
	}
}
