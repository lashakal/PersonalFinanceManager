package dal

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	mongoURI = "mongodb://localhost:27017"
	dbName   = "financeManagerDB"
)

func ConnectToMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping MongoDB server for connection verification
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

// Structure to fit user info
type User struct {
	Username       string    `json:"Username"`
	Email          string    `json:"Email"`
	Password       string    `json:"Password"`
	FirstName      string    `json:"FirstName"`
	LastName       string    `json:"LastName"`
	DateRegistered time.Time `json:"DateRegistered"`
	LastLogin      time.Time `json:"LastLogin"`
}

// Structure to fit transcations info
type Transaction struct {
	UUID        string    `json:"UUID"`
	Username    string    `json:"Username"`
	Type        string    `json:"Type"`
	Amount      float32   `json:"Amount"`
	Category    string    `json:"Category"`
	Date        time.Time `json:"Date"`
	Description string    `json:"Description"`
	IsRecurring string    `json:"IsRecurring"`
	Frequency   string    `json:"Frequency"`
}

// Structure to fit budget plan info
type BudgetPlan struct {
	UUID        string    `json:"UUID"`
	Username    string    `json:"Username"`
	Category    string    `json:"Category"`
	BudgetLimit float32   `json:"BudgetLimit"`
	StartDate   time.Time `json:"StartDate"`
	EndDate     time.Time `json:"EndDate"`
}

type FinanceManagerDB struct {
	Transaction []Transaction
	BudgetPlan  []BudgetPlan
}

// Function to fetch all database collections
func FetchCollections(client *mongo.Client, dbName string) (*FinanceManagerDB, error) {
	financeManagerDB := &FinanceManagerDB{}

	// Define a helper function to fetch and decode documents
	fetchAndDecode := func(collectionName string, results interface{}) error {
		cursor, err := client.Database(dbName).Collection(collectionName).Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.Background())
		if err = cursor.All(context.Background(), results); err != nil {
			return err
		}
		return nil
	}

	// Fetch each collection
	if err := fetchAndDecode("Transaction", &financeManagerDB.Transaction); err != nil {
		return nil, err
	}
	if err := fetchAndDecode("BudgetPlan", &financeManagerDB.BudgetPlan); err != nil {
		return nil, err
	}

	return financeManagerDB, nil
}

// Function to fetch all users
func FetchAllUsers(client *mongo.Client) ([]User, error) {
	collection := client.Database(dbName).Collection("Users")

	// Finding multiple documents returns a cursor
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var users []User
	for cursor.Next(context.TODO()) {
		var user User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	fmt.Printf("Retrieved %d users from the database\n", len(users))
	return users, nil
}

// Function to fetch a user using a usernmaae
func FetchUser(client *mongo.Client, userName string) (User, error) {
	collection := client.Database(dbName).Collection("Users")

	// Create a filter to specify the criteria of the query
	fmt.Println("All Users", collection)
	filter := bson.M{"username": userName}

	// Finding multiple documents returns a cursor
	var user User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// User was not found
			return User{}, fmt.Errorf("no user found with username: %s", userName)
		}
		// Some other error occurred
		return User{}, err
	}

	fmt.Println("the user: ", user)
	return user, nil
}

// Function to create a new user in the database
func CreateUser(client *mongo.Client, user *User) error {
	collection := client.Database(dbName).Collection("Users")

	// Check if the user already exists in DB
	var existingUser User
	filter := bson.M{"username": user.Username}
	err := collection.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err == nil {
		return fmt.Errorf("username %s already exists", user.Username)
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}

	return nil
}

// Function to delete an existing user from the database
func DeleteUser(client *mongo.Client, username string) error {
	collection := client.Database(dbName).Collection("Users")

	filter := bson.M{"username": username}

	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Function to fetch transactions that belong to a particular user
func FetchUserTransactions(client *mongo.Client, username string) ([]Transaction, error) {
	collection := client.Database(dbName).Collection("Transaction")

	// Use a timeout context for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Creating a filter to fetch transactions only for the specified username
	filter := bson.M{"Username": username}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []Transaction
	for cursor.Next(ctx) {
		var transaction Transaction
		err := cursor.Decode(&transaction)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	fmt.Println("light from db dal: ", transactions)
	return transactions, nil
}

// Function to fetch budget plans that belong to a particular user
func FetchUserBudgetPlans(client *mongo.Client, username string) ([]BudgetPlan, error) {
	collection := client.Database(dbName).Collection("BudgetPlan")

	// Use a timeout context for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Creating a filter to fetch budget plans only for the specified username
	filter := bson.M{"Username": username}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgetPlans []BudgetPlan
	for cursor.Next(ctx) {
		var budgetPlan BudgetPlan
		err := cursor.Decode(&budgetPlan)
		if err != nil {
			return nil, err
		}
		budgetPlans = append(budgetPlans, budgetPlan)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	fmt.Println("light from db dal: ", budgetPlans)
	return budgetPlans, nil
}
