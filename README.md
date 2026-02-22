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

### How to Test Phase 2

1. Get a unique URL from [webhook.site](https://webhook.site).
2. Start the server: `go run cmd/webhook-dispatcher/main.go`
3. Send this event in a new terminal (replace the destination URL):

```bash
curl -X POST http://localhost:8080/ingest \
-H "Content-Type: application/json" \
-d '{
  "id": "evt_123",
  "source": "billing_service",
  "type": "payment_failed",
  "destination_url": "  https://webhook.site/7150a180-5df0-4858-9b5f-cc1600239c7a",
  "payload": {"customer_id": "cus_999", "amount": 5000}
}'
```

4. Verify the console logs "accepted" and the payload appears on your
   webhook.site dashboard!

### How to Test Phase 3 (Load Testing)

With worker pools, our server can now ingest thousands of requests per second
without blocking. Let's prove it using `hey`!

1. Install `hey`: `sudo apt install hey`
2. Start the server in one terminal: `go run cmd/webhook-dispatcher/main.go`
3. Run a load test in a new terminal (10,000 requests, 100 concurrent workers)
   replacing YOUR-UNIQUE-URL:

```bash
hey -n 10000 -c 100 -m POST -T "application/json" -d '{
  "id": "evt_load_test",
  "source": "load_generator",
  "type": "test_event",
  "destination_url": "https://webhook.site/YOUR-UNIQUE-URL",
  "payload": {"status": "testing"}
}' http://localhost:8080/ingest
```

4. Watch your Go server terminal—it will ingest all 10,000 requests in
   milliseconds and smoothly process them in the background!

### How to Test Phase 4 (Exponential Backoff)

Let's simulate a broken API to watch our dispatcher automatically retry with
longer delays.

1. Start the server: `go run cmd/webhook-dispatcher/main.go`
2. Send an event intentionally pointing to a broken domain:

```bash
curl -X POST http://localhost:8080/ingest \
-H "Content-Type: application/json" \
-d '{
  "id": "evt_backoff",
  "source": "billing_service",
  "type": "payment_failed",
  "destination_url": "https://this-domain-is-fake-and-broken.com/webhook",
  "payload": {"status": "testing"}
}'
```

3. Watch the Go Server terminal. You will see it fail (attempt 1), wait 1s. Fail
   (attempt 2), wait 2s. Fail (attempt 3), wait 4s. Fail (attempt 4), wait 8s.

### How to Test Phase 5 (Persistence & Dead Letter Queue)

We've successfully migrated the Webhook Dispatcher from an ephemeral memory-only
processing system to a hardened, stateful processing system using
**PostgreSQL**.

1. Start the PostgreSQL and Adminer containers:
   ```bash
   docker-compose up -d
   ```
2. Start the Dispatcher:
   ```bash
   go run cmd/webhook-dispatcher/main.go
   ```
3. Send a Failing Event (Dead Letter Queue Test) to an invalid URL, which will
   trigger our exponential backoff retries.

```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "id": "evt_fail_456",
    "source": "my_app",
    "type": "user.deleted",
    "destination_url": "https://this-domain-does-not-exist.local",
    "payload": {"user_id": 99}
  }'
```

4. The dispatcher will log failures and wait progressively longer (`1s`, `2s`,
   `4s`, `8s`, `16s`). Once the 5th attempt exhausts, it will give up and
   permanently mark the event as `FAILED` in the database!

### View Database via Adminer UI

We've also bundled the lightweight Adminer GUI into our `docker-compose.yml`.
You can view and edit the database graphically straight from your browser!

1. Open [http://localhost:8081](http://localhost:8081) in your browser.
2. Under "System", leave it as **PostgreSQL**.
3. **Server**: `postgres` **(CRITICAL: use `postgres` as the server name, NOT
   `localhost`)**
4. **Username**: The username you set in `.env`
5. **Password**: The password you set in `.env`
6. **Database**: `webhook_db`

Click **Login**, navigate to the `events` table, and you can see your Dead
Letter Queue tracking visually!
