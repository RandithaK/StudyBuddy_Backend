package graph

import "github.com/RandithaK/StudyBuddy_Backend/pkg/store"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Store store.Store
}
