package main

import (
	"context"
	"errors"
	"github.com/almenglee/general"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const dbName = "server"

func ConnectDB() (client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)

	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")

	// MongoDB 연결
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		Log("Error: ConnectDB:", err.Error())
	}

	return client, ctx, cancel
}

func DBFindUsers(filter interface{}) (*general.List[User], error) {
	client, ctx, cancle := ConnectDB()
	defer client.Disconnect(ctx)
	defer cancle()

	var Users []User
	if filter == nil {
		filter = bson.D{}
	}
	usersCollection := GetCollection(client, "user")

	cursor, err := usersCollection.Find(ctx, filter)
	if err != nil {
		Log("Error: DBFindUsers", err.Error())
		return general.EmptyList[User](), err
	}
	if err = cursor.All(ctx, &Users); err != nil {
		Log("Error: DBFindUsers", err.Error())
		return general.EmptyList[User](), err
	}
	return general.AsList(Users), nil
}

var (
	UserNotExistError     = errors.New("user does not exist")
	UserAlreadyExistError = errors.New("user already exist")
)

func DBFindUserOne(filter interface{}) (*User, error) {
	client, ctx, cancle := ConnectDB()
	defer client.Disconnect(ctx)
	defer cancle()

	usersCollection := GetCollection(client, "user")

	num, err := usersCollection.CountDocuments(ctx, filter)
	if err != nil {
		Log("Error: DBFindUserOne", err.Error())
		return nil, err
	}

	if num == 0 {
		return nil, UserNotExistError
	}
	user := new(User)
	err = usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		Log("Error: DBFindUserOne", err.Error())
		return nil, err
	}

	return user, nil
}

func GetCollection(client *mongo.Client, colName string) *mongo.Collection {
	return client.Database(dbName).Collection(colName)
}

func DBCreateUser(user User) (*User, error) {
	client, ctx, cancel := ConnectDB()
	defer client.Disconnect(ctx)
	defer cancel()

	filter := bson.M{"email": user.Email}
	if user.Class == ClassDefault {
		user.Class = ClassUser
	}
	col := GetCollection(client, "user")

	num, err := col.CountDocuments(ctx, filter)
	if err != nil {
		Log("Error: DBCreateUser", err.Error())
		return nil, err
	}

	if num != 0 {
		return nil, UserAlreadyExistError
	}
	res, err := col.InsertOne(ctx, user)
	oid := res.InsertedID.(primitive.ObjectID)
	user.ID = oid.Hex()
	usr := new(User)
	*usr = user
	return usr, nil
}

func DBUpdateUser(filter interface{}, user User) error {
	client, ctx, cancel := ConnectDB()
	defer client.Disconnect(ctx)
	defer cancel()

	update := bson.M{"$set": user}

	_, err := GetCollection(client, "user").UpdateOne(ctx, filter, update)
	return err
}

func DBDeleteUser(username string) error {
	client, ctx, cancel := ConnectDB()
	defer client.Disconnect(ctx)
	defer cancel()

	filter := bson.M{"email": username}
	col := GetCollection(client, "user")

	num, err := col.CountDocuments(ctx, filter)
	if err != nil {
		Log("Error: DBDeleteUsers", err.Error())
		return err
	}

	if num == 0 {
		return UserNotExistError
	}
	_, err = GetCollection(client, "user").DeleteOne(ctx, filter)
	return err
}
