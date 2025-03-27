package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var rnd *renderer.Render
var db *mongo.Database

const (
	hostName       string = "localhost:27017"
	dbName         string = "demo_todo"
	collectionName string = "todo"
	port           string = "9000"
)

type (
	todoModel struct {
		ID        string    `bson:"_id,omitempty"`
		Title     string    `bson:"title"`
		Completed bool      `bson:"completed"`
		CreatedAt time.Time `bson:"createdAt"`
	}

	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {
	rnd = renderer.New()
	log.Println("Attempting to connect to MongoDB at localhost:27017")

	//Set MongoDB URI and options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	//// New MongoDB URI for Kubernetes (Internal Service)
	//clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Create new MongoDB client
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Error creating MongoDB client: %v\n", err)
	}

	// Connect to MongoDB
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v\n", err)
	}

	// Test connection
	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v\n", err)
	}

	db = client.Database(dbName)
	log.Println("Successfully connected to MongoDB!")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	// simple validation
	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is required",
		})
		return
	}

	// if input is okay, create a todo
	tm := todoModel{
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	// Insert the new todo into the MongoDB collection
	collection := db.Collection(collectionName)
	result, err := collection.InsertOne(context.Background(), tm)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to save todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": result.InsertedID,
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	// Check if the id is valid
	if id == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	// simple validation
	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title field is required",
		})
		return
	}

	// Update the todo in the database
	collection := db.Collection(collectionName)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to update todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	var todos []todoModel

	// Fetch all todos from the MongoDB collection
	collection := db.Collection(collectionName)
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fetch todos",
			"error":   err,
		})
		return
	}
	defer cursor.Close(context.Background())

	// Parse the results into the todos slice
	for cursor.Next(context.Background()) {
		var todo todoModel
		if err := cursor.Decode(&todo); err != nil {
			rnd.JSON(w, http.StatusProcessing, renderer.M{
				"message": "Error decoding todo",
				"error":   err,
			})
			return
		}
		todos = append(todos, todo)
	}

	// Convert todoModel to the response format
	todoList := []todo{}
	for _, t := range todos {
		todoList = append(todoList, todo{
			ID:        t.ID,
			Title:     t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	// Check if the id is valid
	if id == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	// Delete the todo from the database
	collection := db.Collection(collectionName)
	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to delete todo",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	// Start the server
	srv := &http.Server{
		Addr:         "localhost:" + port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("Server  stopped!")
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
