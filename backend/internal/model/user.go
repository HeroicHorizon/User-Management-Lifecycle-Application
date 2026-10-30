// Package model defines the user document and request/response types.
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

func (s Status) Valid() bool { return s == StatusActive || s == StatusInactive }

type User struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName   string        `bson:"firstName" json:"firstName"`
	LastName    string        `bson:"lastName" json:"lastName"`
	Email       string        `bson:"email" json:"email"`
	Phone       string        `bson:"phone" json:"phone"`
	Department  string        `bson:"department" json:"department"`
	Designation string        `bson:"designation" json:"designation"`
	Status      Status        `bson:"status" json:"status"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

// UserInput is the body for create and update requests.
type UserInput struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
}

type StatusInput struct {
	Status string `json:"status"`
}

type ListQuery struct {
	Page     int
	PageSize int
	Search   string
	Status   Status // empty means no filter
}

type ListResult struct {
	Items    []User `json:"items"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    int64  `json:"total"`
}

type Stats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
}
