package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/bson"
)

// Task represents a to-do item
type Task struct {
    ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Title       string             `json:"title,omitempty" bson:"title,omitempty"`
    Description string             `json:"description,omitempty" bson:"description,omitempty"`
    Completed   bool              `json:"completed,omitempty" bson:"completed,omitempty"`
}

var client *mongo.Client

// CreateTask creates a new task
func CreateTask(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "application/json")
    var task Task
    _ = json.NewDecoder(request.Body).Decode(&task)
    collection := client.Database("todolist").Collection("tasks")
    result, err :=collection.InsertOne(context.TODO(), task)
	if err!=nil {
		log.Fatal(err);
	}
    json.NewEncoder(response).Encode(result)
}

// GetTasks retrieves all tasks
func GetTasks(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "application/json")
    var tasks []Task
    collection := client.Database("todolist").Collection("tasks")
    cursor, err := collection.Find(context.TODO(), bson.M{})
    if err != nil {
        log.Fatal(err)
    }
    defer cursor.Close(context.TODO())
    for cursor.Next(context.TODO()) {
        var task Task
        cursor.Decode(&task)
        tasks = append(tasks, task)
    }
    json.NewEncoder(response).Encode(tasks)
}

// GetTask retrieves a single task by ID
func GetTask(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "application/json")
    params := mux.Vars(request)
    id, _:= primitive.ObjectIDFromHex(params["id"])
    var task Task
    collection := client.Database("todolist").Collection("tasks")
    err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&task)
    if err != nil {
        json.NewEncoder(response).Encode("Task not found")
        return
    }
    json.NewEncoder(response).Encode(task)
}

// UpdateTask updates a task by ID
func UpdateTask(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "application/json")
    params := mux.Vars(request)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    var task Task
    _ = json.NewDecoder(request.Body).Decode(&task)
    collection := client.Database("todolist").Collection("tasks")
    update := bson.M{"$set": task}
    _, err := collection.UpdateOne(context.TODO(), bson.M{"_id": id}, update)
    if err != nil {
        log.Fatal(err)
    }
    json.NewEncoder(response).Encode("Task updated")
}

// DeleteTask deletes a task by ID
func DeleteTask(response http.ResponseWriter, request *http.Request) {
    response.Header().Set("Content-Type", "application/json")
    params := mux.Vars(request)
    id, _ := primitive.ObjectIDFromHex(params["id"])
    collection := client.Database("todolist").Collection("tasks")
    _, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id})
    if err != nil {
        log.Fatal(err)
    }
    json.NewEncoder(response).Encode("Task deleted")
}

func main() {
    fmt.Println("Starting the application...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, _ = mongo.Connect(ctx, clientOptions)
    router := mux.NewRouter()
    router.HandleFunc("/create", CreateTask).Methods("POST")
    router.HandleFunc("/tasks", GetTasks).Methods("GET")
    router.HandleFunc("/one/{id}", GetTask).Methods("GET")
    router.HandleFunc("/update/{id}", UpdateTask).Methods("PUT")
    router.HandleFunc("/delete/{id}", DeleteTask).Methods("DELETE")
    log.Fatal(http.ListenAndServe(":4000", router))
}
