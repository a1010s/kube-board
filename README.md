# ⚝ kube-board

kube-board is a simple web-based utility built with Golang and Kubernetes client libraries to monitor and display information about pods and deployments in a Kubernetes cluster.

## ▷ Features

- Display status and health information of pods including readiness, status, reasons, and container statuses.
- View deployment details such as name, namespace, replicas, and ready replicas.

## ✓ Requirements

- Go programming language (Golang)
- Kubernetes cluster
- Kubernetes client libraries (client-go)

## ⏎ Installation

- Clone the kube-board repository to your local machine:

   ```bash
   git clone https://github.com/a1010s/kube-board.git
   cd kube-board
   go build
   ./kube-board
   ```
- Or build the Dockerfile:
   ```bash
   docker build -t kube-board:latest .
   docker run -d -p 8080:8080 kube-board:latest
   ```

## ⎈ Usage
- Navigate to http://localhost:8080/pod to view information about pods in the Kubernetes cluster.
- Access http://localhost:8080/deploy to see details about deployments.
