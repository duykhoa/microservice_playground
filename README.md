# Microservice Playground

## Overview

This project is a simple microservice application that demonstrates a basic e-commerce system. It consists of:

- A web frontend built with Vue.js and served by Nginx.
- A `products-service` written in Go.
- An `order-service` written in Go, which uses RabbitMQ for messaging.
- A `warehouse` service written in Go.

The entire application is containerized using Docker and deployed to Kubernetes using Skaffold and Kustomize.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Devbox](https://www.jetpack.io/devbox/)
- [Docker](https://www.docker.com/)
- [Make](https://www.gnu.org/software/make/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)

## Getting Started

1.  **Start the development environment:**

    Open your terminal and navigate to the project root. Then, start the Devbox shell:

    ```bash
    devbox shell
    ```

2.  **Deploy the application:**

    Once inside the Devbox shell, you can build the Docker images and deploy all the services to your Kubernetes cluster using the following command:

    ```bash
    make dev
    ```

    This command uses Skaffold to manage the build and deployment process.

## Usage

-   **Access the Web UI:**

    To open the application's user interface in your browser, run the following command. This will create a tunnel to the `web` service running in Minikube.

    ```bash
    minikube service web
    ```

-   **Open the Kubernetes Dashboard:**

    To view the Kubernetes dashboard and inspect the running services, you can use the following Make command:

    ```bash
    make dashboard
    ```
