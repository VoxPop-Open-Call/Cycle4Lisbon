package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"bitbucket.org/pensarmais/cycleforlisbon/src/config"
	"bitbucket.org/pensarmais/cycleforlisbon/src/database"
)

func Example() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	config, _ := config.Load("../../.env")
	db, err := database.Init(config.DbDsn())
	if err != nil {
		fmt.Printf("error starting db: %v", err)
		return
	}

	db.Exec("DELETE FROM worker_tasks")

	// Setup and start worker
	MaxRoutineNum = 1
	wrkr := New(NewDbQueue(db))
	wrkr.LogLevel = Info
	wrkr.PollingSleepTime = 50 * time.Millisecond
	go func() {
		log.Println("starting worker")
		wrkr.Run()
	}()

	// Define and register a job
	ch := make(chan int, 1)
	job := &Job{
		Name: "example",
		Handler: func(_ context.Context, _ []byte) error {
			fmt.Println("doing some work")
			ch <- 1
			return nil
		},
	}
	err = wrkr.Register(job)
	if err != nil {
		fmt.Printf("error registering job: %v\n", err)
		return
	}

	// Schedule a task
	date, _ := time.Parse("2006-01-02", "2023-01-01")
	err = wrkr.Schedule(&TaskConfig{
		JobName:     "example",
		ScheduledTo: date,
	})
	if err != nil {
		fmt.Printf("error scheduling task: %v\n", err)
		return
	}

	select {
	case <-ch:
		// the job was completed
		time.Sleep(10 * time.Millisecond)
	case <-time.After(5 * time.Second):
		fmt.Println("timeout running job")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := wrkr.Stop(ctx); err != nil {
		log.Println(err)
	}

	// Output:
	// starting worker
	// worker: running task: example at 2023-01-01 00:00:00
	// doing some work
	// worker: task completed: example at 2023-01-01 00:00:00
	// worker: stopping
	// worker: waiting for running tasks
	// worker: stopped
}
