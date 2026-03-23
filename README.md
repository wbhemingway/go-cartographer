# Go-Cartographer

A concurrent microservice and rendering engine for generating 2D grid-based maps. Designed to support tabletop role-playing games and community-driven bots, the engine processes tile arrays, safely renders images using a dedicated worker pool, and syncs the output to Google Cloud.

## Features

* **Concurrent Rendering Engine:** An internal worker pool processes heavy image rendering jobs asynchronously, preventing the main API threads from blocking or running out of memory under load.
* **RESTful API:** Clean, standard routing for submitting jobs, checking status, and managing maps.
* **Cloud Storage:** Integrates securely with Google Cloud Storage for persistent image hosting and Firestore for fast, document-based job tracking.
* **Dedicated Go Client:** Includes a fully featured, idiomatic Go client package (`pkg/client`) utilizing the functional options pattern for seamless integration into other microservices.
* **Docker Ready:** Seperated into API and Worker containers for scalable deployment.

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
  Copy the example enviorment, and set up your google cloud enviorment to get the necessary variables. The api key is something you'll have to make and put into firestore yourself, as well as a user attatched to it.

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
