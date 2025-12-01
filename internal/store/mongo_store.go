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
			// Calculate totalTasks and completedTasks for this course
			tasksCol := m.db.Collection("tasks")
			tasksCtx, tasksCancel := context.WithTimeout(context.Background(), 5*time.Second)

			// Count total tasks for this course
			totalCount, _ := tasksCol.CountDocuments(tasksCtx, bson.M{
				"userId":   userID,
				"courseId": c.ID,
			})
			c.TotalTasks = int(totalCount)

			// Count completed tasks for this course
			completedCount, _ := tasksCol.CountDocuments(tasksCtx, bson.M{
				"userId":    userID,
				"courseId":  c.ID,
				"completed": true,
			})
			c.CompletedTasks = int(completedCount)

			tasksCancel()
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

func (m *MongoStore) GetUserByVerificationToken(token string) (models.User, error) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var u models.User
	res := col.FindOne(ctx, bson.M{"verificationToken": token})
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

func (m *MongoStore) UpdateUser(id string, u models.User) (models.User, error) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure we don't overwrite the ID or Email if not intended, but here we expect u to have updated fields
	// For safety, let's just update Name and Email if provided.
	// However, the interface takes a User model. Let's assume the caller prepares the User object correctly.
	// But wait, we need to be careful about not clearing other fields if we had them (like Password).
	// The current User model only has ID, Name, Email, Password.
	// We should probably fetch the existing user first to preserve the password if it's not being updated,
	// or assume the caller handles it.
	// Given the plan "UpdateUser(id string, u models.User)", let's do a replacement or partial update.
	// A partial update is safer.

	update := bson.M{}
	if u.Name != "" {
		update["name"] = u.Name
	}
	if u.Email != "" {
		update["email"] = u.Email
	}
	if u.IsVerified {
		update["isVerified"] = u.IsVerified
	}
	if u.VerificationToken != "" {
		update["verificationToken"] = u.VerificationToken
	}
	// We explicitly don't update password here for now as it wasn't in the requirements,
	// but if we needed to, we would.

	if len(update) == 0 {
		return models.User{}, nil // Nothing to update
	}

	res, err := col.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		return models.User{}, err
	}
	if res.MatchedCount == 0 {
		return models.User{}, ErrNotFound
	}

	// Return the updated user
	return m.GetUser(id)
}

func (m *MongoStore) UpdateUserPassword(id string, hashedPassword string) (models.User, error) {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.M{"password": hashedPassword}
	res, err := col.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		return models.User{}, err
	}
	if res.MatchedCount == 0 {
		return models.User{}, ErrNotFound
	}
	return m.GetUser(id)
}

func (m *MongoStore) MarkUserVerified(id string) error {
	col := m.db.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"isVerified":        true,
		"verificationToken": "",
	}

	res, err := col.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Notifications
func (m *MongoStore) GetNotifications(userID string) []models.Notification {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sort by createdAt desc
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cur, err := col.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return []models.Notification{}
	}
	defer cur.Close(ctx)
	var res []models.Notification
	for cur.Next(ctx) {
		var n models.Notification
		if err := cur.Decode(&n); err == nil {
			res = append(res, n)
		}
	}
	return res
}

func (m *MongoStore) GetNotificationByReferenceID(refID string, nType string) (models.Notification, error) {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var n models.Notification
	err := col.FindOne(ctx, bson.M{"referenceId": refID, "type": nType}).Decode(&n)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.Notification{}, ErrNotFound
		}
		return models.Notification{}, err
	}
	return n, nil
}

func (m *MongoStore) CreateNotification(n models.Notification) models.Notification {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	if n.CreatedAt == "" {
		n.CreatedAt = time.Now().Format(time.RFC3339)
	}
	_, _ = col.InsertOne(ctx, n)
	return n
}

func (m *MongoStore) MarkNotificationAsRead(id string) error {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := col.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": bson.M{"read": true}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *MongoStore) GetUnreadNotificationsOlderThan(duration string) ([]models.Notification, error) {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-d).Format(time.RFC3339)

	// Find unread notifications created before cutoff and not yet emailed
	filter := bson.M{
		"read":      false,
		"emailed":   false,
		"createdAt": bson.M{"$lt": cutoff},
	}

	cur, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []models.Notification
	for cur.Next(ctx) {
		var n models.Notification
		if err := cur.Decode(&n); err == nil {
			res = append(res, n)
		}
	}
	return res, nil
}

func (m *MongoStore) MarkNotificationAsEmailed(id string) error {
	col := m.db.Collection("notifications")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := col.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": bson.M{"emailed": true}})
	return err
}

// Worker Helpers
func (m *MongoStore) GetTasksDueIn(duration string) ([]models.Task, error) {
	col := m.db.Collection("tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}

	// We want tasks due between now and now+duration
	now := time.Now()
	target := now.Add(d)

	// Assuming DueDate is "YYYY-MM-DD" and DueTime is "HH:MM"
	// This is a bit tricky with string dates.
	// Let's assume we can construct a comparable string or we need to fetch and filter.
	// Fetching all incomplete tasks and filtering in Go is safer for string dates if dataset isn't huge.
	// Or we can rely on strict format.

	// Let's fetch all incomplete tasks and filter in memory for simplicity and correctness with string formats
	cur, err := col.Find(ctx, bson.M{"completed": false})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var res []models.Task
	for cur.Next(ctx) {
		var t models.Task
		if err := cur.Decode(&t); err == nil {
			// Parse due date/time
			// Format: 2023-10-27 14:30
			dueStr := t.DueDate + " " + t.DueTime
			due, err := time.Parse("2006-01-02 15:04", dueStr)
			if err == nil {
				// Check if due is within the range [now, target]
				// Also check if we already notified?
				// The requirement says "before 24 hours".
				// We probably need a flag on Task or check if a notification exists.
				// Checking if notification exists is expensive.
				// Let's assume we run this periodically and we want to catch tasks due in ~24h.
				// To avoid duplicates, we can check if we are close to the 24h mark (e.g. 23h-24h window)
				// OR we can add a "Notified24h" flag to Task.
				// Adding a flag is better. But I can't easily change the schema right now without more files.
				// Let's check if notification exists for this task with type TASK_DUE.

				if due.After(now) && due.Before(target) {
					res = append(res, t)
				}
			}
		}
	}
	return res, nil
}

func (m *MongoStore) GetEventsStartingIn(duration string) ([]models.Event, error) {
	col := m.db.Collection("events")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	target := now.Add(d)

	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var res []models.Event
	for cur.Next(ctx) {
		var e models.Event
		if err := cur.Decode(&e); err == nil {
			startStr := e.Date + " " + e.StartTime
			start, err := time.Parse("2006-01-02 15:04", startStr)
			if err == nil {
				if start.After(now) && start.Before(target) {
					res = append(res, e)
				}
			}
		}
	}
	return res, nil
}
