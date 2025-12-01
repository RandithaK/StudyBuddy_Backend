package store

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/RandithaK/StudyBuddy/backend/internal/models"
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
func (m *MongoStore) GetTasks(userID string) []models.Task {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return []models.Task{}
	}
	defer cur.Close(ctx)
	var res []models.Task
	for cur.Next(ctx) {
		var t models.Task
		if err := cur.Decode(&t); err == nil {
			res = append(res, t)
		}
	}
	return res
}

func (m *MongoStore) GetTask(id string) (models.Task, error) {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// try to search by id field
	var t models.Task
	res := col.FindOne(ctx, bson.M{"id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Task{}, ErrNotFound
		}
		return models.Task{}, err
	}
	if err := res.Decode(&t); err != nil {
		return models.Task{}, err
	}
	return t, nil
}

func (m *MongoStore) CreateTask(t models.Task) models.Task {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, t)
	return t
}

func (m *MongoStore) UpdateTask(id string, t models.Task) (models.Task, error) {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	t.ID = id
	res, err := col.ReplaceOne(ctx, bson.M{"id": id}, t)
	if err != nil {
		return models.Task{}, err
	}
	if res.MatchedCount == 0 {
		return models.Task{}, ErrNotFound
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
func (m *MongoStore) GetCourses(userID string) []models.Course {
	col := m.db.Collection("courses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return []models.Course{}
	}
	defer cur.Close(ctx)
	var res []models.Course
	for cur.Next(ctx) {
		var c models.Course
		if err := cur.Decode(&c); err == nil {
			res = append(res, c)
		}
	}
	return res
}

func (m *MongoStore) GetCourse(id string) (models.Course, error) {
	col := m.db.Collection("courses")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var c models.Course
	res := col.FindOne(ctx, bson.M{"id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Course{}, ErrNotFound
		}
		return models.Course{}, err
	}
	if err := res.Decode(&c); err != nil {
		return models.Course{}, err
	}
	return c, nil
}

func (m *MongoStore) CreateCourse(c models.Course) models.Course {
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
func (m *MongoStore) GetEvents(userID string) []models.Event {
	col := m.db.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := col.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return []models.Event{}
	}
	defer cur.Close(ctx)
	var res []models.Event
	for cur.Next(ctx) {
		var e models.Event
		if err := cur.Decode(&e); err == nil {
			res = append(res, e)
		}
	}
	return res
}

func (m *MongoStore) CreateEvent(e models.Event) models.Event {
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
func (m *MongoStore) GetUser(id string) (models.User, error) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u models.User
	res := col.FindOne(ctx, bson.M{"id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	if err := res.Decode(&u); err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (m *MongoStore) GetUserByEmail(email string) (models.User, bool) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u models.User
	res := col.FindOne(ctx, bson.M{"email": email})
	if err := res.Err(); err != nil {
		return models.User{}, false
	}
	if err := res.Decode(&u); err != nil {
		return models.User{}, false
	}
	return u, true
}

func (m *MongoStore) CreateUser(u models.User) models.User {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	_, _ = col.InsertOne(ctx, u)
	return u
}
