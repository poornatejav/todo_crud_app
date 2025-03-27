# Todo App with MongoDB and Go

This is a simple **Todo** application built with **Go** that uses **MongoDB** for data storage. The application allows users to create, update, fetch, and delete todo items via RESTful API endpoints. The app also includes basic validation and error handling.

## Features
- Create a new todo
- Fetch all todos
- Update an existing todo
- Delete a todo
- Built with **Go** and **MongoDB**

## Prerequisites

Before running the application, ensure you have the following installed:

1. **Go** (version 1.17+)
   - [Install Go](https://golang.org/dl/)

2. **MongoDB** (version 4.0+)
   - [Install MongoDB](https://www.mongodb.com/try/download/community)

## Setup and Installation

Follow these steps to get the API running on your local machine:

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/poornatejav/todo_crud_app.git
    ```

2.  **Navigate to the project directory:**
    ```bash
   cd todo_crud_app
   ```

3.  **Download dependencies:**
    ```bash
    go mod tidy
    ```
    This command will download all the necessary Go modules defined in the `go.mod` file.

4.  **Ensure MongoDB is running:**
    Make sure your MongoDB server is running and accessible at `localhost:27017` (as configured in the `main.go` file). If your MongoDB instance is running on a different host or port, you'll need to update the `hostName` constant in the `main.go` file accordingly.

## Running the API

1.  **Start the Go application:**
    ```bash
    go run main.go
    ```
    You should see output similar to:
    ```
    2025/03/27 09:55:00 Attempting to connect to MongoDB at localhost:27017
    2025/03/27 09:55:00 Successfully connected to MongoDB!
    2025/03/27 09:55:00 Listening on 9000
    ```
    This indicates that the API server has started and is listening for requests on port `9000`.

## Accessing the Application from the Web

Once the application is running, you can access a basic HTML page that interacts with the API by opening your web browser and navigating to: http://localhost:9000/

## Using the API

You can use tools like `curl` or Postman to interact with the API endpoints. Here are a few examples using `curl`:

* **Get all todos:**
    ```bash
    curl http://localhost:9000/todo
    ```

* **Create a new todo:**
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"title": "Pay bills"}' http://localhost:9000/todo
    ```
    
## Stopping the Server

To stop the API server, you can typically press `Ctrl+C` in the terminal where it's running. The application is set up to gracefully shut down upon receiving an interrupt signal.
