# Go-Cartographer

A concurrent microservice and rendering engine for generating 2D grid-based maps. Designed to support tabletop role-playing games and community-driven bots, the engine processes tile arrays, safely renders images using a dedicated worker pool, and syncs the output to Google Cloud.

## Features

* **Concurrent Rendering Engine:** An internal worker pool processes heavy image rendering jobs asynchronously, preventing the main API threads from blocking or running out of memory under load.
* **RESTful API:** Clean, standard routing for submitting jobs, checking status, and managing maps.
* **Cloud Storage:** Integrates securely with Google Cloud Storage for persistent image hosting and Firestore for fast, document-based job tracking.
* **Dedicated Go Client:** Includes a fully featured, idiomatic Go client package (`pkg/client`) utilizing the functional options pattern for seamless integration into other microservices.
* **Docker Ready:** Separated into API and Worker containers for scalable deployment.

## Architecture

* **`cmd/api`**: The REST API that authenticates requests, stores the initial job state in Firestore, and publishes the job to the queue.
* **`cmd/worker`**: The background processor that picks up jobs, builds the map using the `internal/renderer`, and uploads the final PNG to Google Cloud Storage.
* **`pkg/client`**: The decoupled, versioned SDK for consuming the API.

## Prerequisites

* Go 1.22+
* Docker & Docker Compose
* A Google Cloud Service Account with permissions for Cloud Storage and Firestore (Datastore)

## Getting Started

1. **Clone the repository:**
  ```bash
  git clone [https://github.com/wbhemingway/go-cartographer.git](https://github.com/wbhemingway/go-cartographer.git)
  cd go-cartographer
  ```

2. **Configure the environment**
  Copy the example environment, and set up your google cloud environment to get the necessary variables. The api key is something you'll have to make and put into firestore yourself, as well as a user attatched to it.
  ```bash
  cp .env.example .env
  ```

3. **Run the services**
  Use the included Makefile to build and start the API and Worker containers by inputting:
  ```bash
  make all
  ```
**Testing**
To verify it is working as intended run the client-test
 ```bash
 go run ./cmd/client-test
  ```


## API Reference

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/maps` | Submits a `WorldRequest` payload. Returns `202 Accepted` and a Job ID. |
| `GET` | `/maps/{id}` | Retrieves job status. Returns the GCS Image URL when `status == "completed"`. |
| `GET` | `/maps` | Lists all maps associated with the authenticated user. |
| `DELETE` | `/maps/{id}` | Deletes the job from Firestore and removes the image from Cloud Storage. |

## Client Package Usage

The repository includes a decoupled client package abstracting transport logic and enforcing JSON contracts.

```go
import (
    "context"
    "time"
    cartographer "[github.com/wbhemingway/go-cartographer/pkg/client](https://github.com/wbhemingway/go-cartographer/pkg/client)"
)

func main() {
    c := cartographer.NewClient(
        "http://localhost:8080",
        cartographer.WithTimeout(10*time.Second),
        cartographer.WithAPIKey("your-secret-key"),
    )

    // Submit a map render job
    resp, err := c.Render(context.Background(), myWorldRequest)
}
```
