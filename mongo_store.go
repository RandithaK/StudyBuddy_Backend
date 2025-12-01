package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoStore(ctx context.Context, uri, dbName string) (*MongoStore, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	// Ping with timeout
	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing, nil); err != nil {
		return nil, err
	}
	db := client.Database(dbName)
	log.Printf("connected to mongodb database %s", dbName)
	return &MongoStore{client: client, db: db}, nil
}

var ErrMongoNotFound = errors.New("not found")

// Helper IDs
func toObjectID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NilObjectID, nil
	}
	// If string is a valid hex for ObjectID, parse it, else use uuid string as _id.
	if oid, err := primitive.ObjectIDFromHex(id); err == nil {
		return oid, nil
	}
	// Not an ObjectID; store as a string id in field "id"
	return primitive.NilObjectID, nil
}

// MongoStore implements Store
func (m *MongoStore) GetTasks() []Task {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return []Task{}
	}
	defer cur.Close(ctx)
	var res []Task
	for cur.Next(ctx) {
		var t Task
		if err := cur.Decode(&t); err == nil {
			res = append(res, t)
		}
	}
	return res
}

func (m *MongoStore) GetTask(id string) (Task, error) {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// try to search by id field
	var t Task
	res := col.FindOne(ctx, bson.M{"id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return Task{}, ErrNotFound
		}
		return Task{}, err
	}
	if err := res.Decode(&t); err != nil {
		return Task{}, err
	}
	return t, nil
}

func (m *MongoStore) CreateTask(t Task) Task {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, t)
	return t
}

func (m *MongoStore) UpdateTask(id string, t Task) (Task, error) {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.ID = id
	res, err := col.ReplaceOne(ctx, bson.M{"id": id}, t)
	if err != nil {
		return Task{}, err
	}
	if res.MatchedCount == 0 {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (m *MongoStore) DeleteTask(id string) error {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := col.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Courses
func (m *MongoStore) GetCourses() []Course {
	col := m.db.Collection("courses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return []Course{}
	}
	defer cur.Close(ctx)
	var res []Course
	for cur.Next(ctx) {
		var c Course
		if err := cur.Decode(&c); err == nil {
			res = append(res, c)
		}
	}
	return res
}

func (m *MongoStore) CreateCourse(c Course) Course {
	col := m.db.Collection("courses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, c)
	return c
}

// Events
func (m *MongoStore) GetEvents() []Event {
	col := m.db.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return []Event{}
	}
	defer cur.Close(ctx)
	var res []Event
	for cur.Next(ctx) {
		var e Event
		if err := cur.Decode(&e); err == nil {
			res = append(res, e)
		}
	}
	return res
}

func (m *MongoStore) CreateEvent(e Event) Event {
	col := m.db.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, e)
	return e
}

// Users
func (m *MongoStore) GetUser(id string) (User, error) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u User
	res := col.FindOne(ctx, bson.M{"id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	if err := res.Decode(&u); err != nil {
		return User{}, err
	}
	return u, nil
}

func (m *MongoStore) GetUserByEmail(email string) (User, bool) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u User
	res := col.FindOne(ctx, bson.M{"email": email})
	if err := res.Err(); err != nil {
		return User{}, false
	}
	if err := res.Decode(&u); err != nil {
		return User{}, false
	}
	return u, true
}

func (m *MongoStore) CreateUser(u User) User {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, u)
	return u
}
