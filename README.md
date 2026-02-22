# Go Webhook Dispatcher

A resilient, scalable, and concurrent distributed webhook dispatcher built in
Go.

## Overview

This project serves as a reliable middleman for event delivery. It ingests
events from source applications (like a core API or e-commerce backend) and
dispatches them to registered destination URLs.

Key architectural features include:

- **High Throughput**: Concurrent event processing using Goroutines and worker
  pools.
- **Robustness**: Advanced failure handling with exponential backoff and
  retries.
- **Reliability**: Zero-data-loss architecture utilizing Dead Letter Queues
  (DLQ).

For in-depth architectural details, presentation scenarios, and feature
breakdowns, please see [PROJECT_DETAILS.md](PROJECT_DETAILS.md).

## How to Run

1. Clone the repository and navigate to the project root.
2. Start the server:
   ```bash
   go run cmd/webhook-dispatcher/main.go
   ```
3. In a separate terminal window, verify the server is running by hitting the
   health check endpoint:
   ```bash
   curl http://localhost:8080/healthz
   ```
   _You should see `OK` in the terminal, and structured JSON logs in the server
   terminal._
4. To gracefully stop the server, press `Ctrl+C` in the terminal where it is
   running.
