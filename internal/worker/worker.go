package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/RandithaK/StudyBuddy_Backend/internal/email"
	"github.com/RandithaK/StudyBuddy_Backend/internal/models"
	"github.com/RandithaK/StudyBuddy_Backend/internal/store"
)

type Worker struct {
	Store store.Store
}

func NewWorker(s store.Store) *Worker {
	return &Worker{Store: s}
}

func (w *Worker) Start() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	go func() {
		for range ticker.C {
			w.CheckUpcomingTasks()
			w.CheckUpcomingEvents()
			w.CheckUnreadNotifications()
		}
	}()
}

func (w *Worker) CheckUpcomingTasks() {
	// Get tasks due in the next 24 hours
	tasks, err := w.Store.GetTasksDueIn("24h")
	if err != nil {
		log.Printf("Error getting upcoming tasks: %v", err)
		return
	}

	for _, t := range tasks {
		// Check if we already created a notification for this task
		_, err := w.Store.GetNotificationByReferenceID(t.ID, "TASK_DUE")
		if err == nil {
			// Notification already exists
			continue
		}

		// Create notification
		n := models.Notification{
			UserID:      t.UserID,
			Message:     fmt.Sprintf("Task '%s' is due in less than 24 hours!", t.Title),
			Type:        "TASK_DUE",
			ReferenceID: t.ID,
			Read:        false,
			Emailed:     false,
		}
		w.Store.CreateNotification(n)
		log.Printf("Created notification for task %s", t.ID)
	}
}

func (w *Worker) CheckUpcomingEvents() {
	events, err := w.Store.GetEventsStartingIn("24h")
	if err != nil {
		log.Printf("Error getting upcoming events: %v", err)
		return
	}

	for _, e := range events {
		_, err := w.Store.GetNotificationByReferenceID(e.ID, "EVENT_START")
		if err == nil {
			continue
		}

		n := models.Notification{
			UserID:      e.UserID,
			Message:     fmt.Sprintf("Event '%s' is starting in less than 24 hours!", e.Title),
			Type:        "EVENT_START",
			ReferenceID: e.ID,
			Read:        false,
			Emailed:     false,
		}
		w.Store.CreateNotification(n)
		log.Printf("Created notification for event %s", e.ID)
	}
}

func (w *Worker) CheckUnreadNotifications() {
	// Get unread notifications older than 1 hour
	notifications, err := w.Store.GetUnreadNotificationsOlderThan("1h")
	if err != nil {
		log.Printf("Error getting unread notifications: %v", err)
		return
	}

	for _, n := range notifications {
		user, err := w.Store.GetUser(n.UserID)
		if err != nil {
			log.Printf("Error getting user %s: %v", n.UserID, err)
			continue
		}

		// Only send email if user is verified
		if !user.IsVerified {
			log.Printf("Skipping email for unverified user %s", user.Email)
			// We still mark as emailed so we don't keep checking?
			// Or we leave it as not emailed?
			// If we leave it, we'll keep checking every minute.
			// Better to mark it as emailed (or "processed") to avoid loop.
			w.Store.MarkNotificationAsEmailed(n.ID)
			continue
		}

		err = email.SendNotificationEmail(user.Email, "You have an unread notification", n.Message)
		if err != nil {
			log.Printf("Error sending email to %s: %v", user.Email, err)
			continue
		}

		// Mark as emailed so we don't send again
		w.Store.MarkNotificationAsEmailed(n.ID)
		log.Printf("Sent email for notification %s", n.ID)
	}
}
