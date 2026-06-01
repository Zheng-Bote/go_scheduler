package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"go-scheduler/internal/ipc"
)

func main() {
	runIDStr := os.Getenv("RUN_ID")
	socketPath := os.Getenv("SCHEDULER_SOCKET_PATH")

	if runIDStr == "" || socketPath == "" {
		log.Fatal("Missing environment variables RUN_ID or SCHEDULER_SOCKET_PATH")
	}

	runID, _ := strconv.Atoi(runIDStr)
	client := &ipc.Client{SocketPath: socketPath}

	log.Printf("Job 2 started (RunID: %d)", runID)

	for i := 0; i <= 100; i += 20 {
		status := "processing"
		if i == 0 {
			status = "started"
		} else if i == 100 {
			status = "finished"
		}

		err := client.SendEvent(ipc.StatusEvent{
			RunID:    runID,
			Status:   status,
			Message:  fmt.Sprintf("Job 2 progress: %d%%", i),
			Progress: i,
		})
		if err != nil {
			log.Printf("Failed to send IPC event: %v", err)
		}

		if i < 100 {
			time.Sleep(1 * time.Second)
		}
	}

	log.Println("Job 2 finished")
}
