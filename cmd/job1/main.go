/**
 * SPDX-FileComment: Job 1
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file main.go
 * @brief Example scheduled job demonstrating IPC status reporting
 * @version 1.0.0
 * @date 2026-06-02
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package main

import (
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

	// 1. Started
	client.SendEvent(ipc.StatusEvent{
		RunID:    runID,
		Status:   "started",
		Message:  "Job 1 is starting its work",
		Progress: 0,
	})

	// 2. Simulate Work
	time.Sleep(1 * time.Second)

	// Audit Log Example
	client.SendEvent(ipc.StatusEvent{
		RunID:   runID,
		Type:    "audit",
		Message: "Job 1 is performing a security-sensitive operation",
	})

	time.Sleep(1 * time.Second)

	// 3. Processing
	client.SendEvent(ipc.StatusEvent{
		RunID:    runID,
		Status:   "processing",
		Message:  "Job 1 is halfway done",
		Progress: 50,
	})

	time.Sleep(2 * time.Second)

	// 4. Finished
	client.SendEvent(ipc.StatusEvent{
		RunID:    runID,
		Status:   "finished",
		Message:  "Job 1 completed successfully",
		Progress: 100,
	})

	log.Println("Job 1 finished")
}
