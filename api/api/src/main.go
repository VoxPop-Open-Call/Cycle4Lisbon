package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/achievements"
	"bitbucket.org/pensarmais/cycleforlisbon/src/aws"
	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database/query"
	"bitbucket.org/pensarmais/cycleforlisbon/src/firebase"
	"bitbucket.org/pensarmais/cycleforlisbon/src/jobs"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server"
	"bitbucket.org/pensarmais/cycleforlisbon/src/server/dex"
	"bitbucket.org/pensarmais/cycleforlisbon/src/util/latlon"
	"bitbucket.org/pensarmais/cycleforlisbon/src/worker"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

var (
	seed          = flag.Bool("seed", false, "Insert seed data in the database")
	envFile       = flag.String("env", ".env", "Path to the env file to load")
	dexConfigFile = flag.String("dex", ".dexrc.json", "Path to the Dex config file")
	fbaseCredFile = flag.String("firebase", ".firebase.json", "Path to the Firebase credentials file")
)

func main() {
	flag.Parse()

	conf, err := config.Load(*envFile)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              conf.SENTRY_DSN,
		AttachStacktrace: true,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Environment:      conf.ENV,
	}); err != nil {
		log.Fatalf("failed to initialize Sentry: %v\n", err)
	}
	defer handlePanic(quit)

	// gin reads the env in an init function, before godotenv loads it.
	// We have to set the mode manually.
	gin.SetMode(conf.GIN_MODE)

	db, err := database.Init(conf.DbDsn())
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}
	if *seed {
		database.Seed(db)
	}
	database.CreateAdminUser(conf.API_ADMIN_EMAIL, conf.API_ADMIN_PASSWD, db)

	fbase, err := firebase.New(fbaseCredFile)

	if err != nil {
		log.Fatalf("failed to initialize Firebase: %v", err)
	}

	awsClient, err := aws.New(aws.Config{
		Region:          conf.AWS_REGION,
		AccessKeyId:     conf.AWS_ACCESS_KEY_ID,
		SecretAccessKey: conf.AWS_SECRET_ACCESS_KEY,
		BucketName:      conf.AWS_S3_BUCKET,
		SESDomain:       conf.AWS_SES_DOMAIN,
		SESFromName:     conf.AWS_SES_FROM_NAME,
		APIScheme:       conf.SCHEME,
		APIEndpoint:     conf.API_HOST + "/api",
	})
	if err != nil {
		log.Fatalf("failed to initialize the AWS client: %v", err)
	}

	dexStore, err := dex.OpenStorage(&dex.StorageConfig{
		Database: conf.DEX_DB_NAME,
		User:     conf.DEX_DB_USER,
		Password: conf.DEX_DB_PASSWORD,
		Host:     conf.DEX_DB_HOST,
		Port:     conf.DEX_DB_PORT,
		SSL:      conf.DEX_DB_SSL,
	})
	if err != nil {
		log.Fatalf("failed to open Dex storage: %v", err)
	}
	defer dexStore.Close()

	achs, err := achievements.New(db, query.Achievements)
	if err != nil {
		log.Fatalf("failed to initialize achievements: %v", err)
	}

	wrkr := worker.New(worker.NewDbQueue(db))

	if err = wrkr.Register(
		jobs.All(wrkr, fbase, db, achs, conf.ServerBaseURL())...,
	); err != nil {
		log.Fatalf("error registering job: %v", err)
	}

	scheduleRecurringTasks(wrkr)

	srv, err := server.New(&server.Config{
		ApiHost:       conf.API_HOST,
		ServerBaseURL: conf.ServerBaseURL(),
		DB:            db,
		DbDsn:         conf.DbDsn(),
		DexConfigFile: *dexConfigFile,
		DexStorage:    dexStore,
		Worker:        wrkr,
		AWS:           awsClient,
		Geocoder:      latlon.NewGeocoder(conf.GOOGLE_API_KEY),
	})
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}

	go func() {
		log.Println("starting worker")
		defer handlePanic(quit)
		wrkr.Run()
	}()

	go func() {
		log.Println("starting http server")
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("%v", err)
		}
	}()

	<-quit

	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println(err)
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := wrkr.Stop(ctx); err != nil {
			log.Println(err)
		}
	}
}

func scheduleRecurringTasks(wrkr *worker.Worker) {
	if err := wrkr.Schedule(&worker.TaskConfig{
		JobName:     jobs.FcmTokenCleanup,
		ScheduledTo: time.Now().Add(10 * time.Second),
	}); err != nil {
		log.Printf("failed to schedule token cleanup: %v", err)
	}

	if err := wrkr.Schedule(&worker.TaskConfig{
		JobName:     jobs.PasswordResetCodeCleanup,
		ScheduledTo: time.Now().Add(12 * time.Second),
	}); err != nil {
		log.Printf("failed to schedule password reset code cleanup: %v", err)
	}

	if err := wrkr.Schedule(&worker.TaskConfig{
		JobName:     jobs.FetchEvents,
		ScheduledTo: time.Now().Add(15 * time.Second),
	}); err != nil {
		log.Printf("failed to schedule event fetching: %v", err)
	}

	if err := wrkr.Schedule(&worker.TaskConfig{
		JobName:     jobs.FetchNews,
		ScheduledTo: time.Now().Add(20 * time.Second),
	}); err != nil {
		log.Printf("failed to schedule news fetching: %v", err)
	}
}

// handlePanic recovers form panics, reports them to Sentry and sends an
// interrupt signal on the quit channel.
func handlePanic(quit chan<- os.Signal) {
	if err := recover(); err != nil {
		log.Println("panic:", err)
		sentry.CurrentHub().Recover(err)
		sentry.Flush(time.Second * 5)
		quit <- os.Interrupt
	}
}
