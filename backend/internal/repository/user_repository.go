// Package repository persists users in MongoDB.
package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/model"
)

// UserRepository is the persistence seam used by the service layer.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
	UpdateStatus(ctx context.Context, id bson.ObjectID, s model.Status, updatedAt time.Time) (*model.User, error)
	Delete(ctx context.Context, id bson.ObjectID) error
	List(ctx context.Context, skip, limit int64, search string, status model.Status) ([]model.User, int64, error)
	Stats(ctx context.Context) (model.Stats, error)
}

type MongoUserRepository struct {
	coll *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{coll: db.Collection("users")}
}

// EnsureIndexes creates the unique email index.
func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})
	return mapErr(err)
}

func (r *MongoUserRepository) Create(ctx context.Context, u *model.User) error {
	res, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		return mapErr(err)
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		u.ID = id
	}
	return nil
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, u *model.User) error {
	set := bson.M{
		"firstName":   u.FirstName,
		"lastName":    u.LastName,
		"email":       u.Email,
		"phone":       u.Phone,
		"department":  u.Department,
		"designation": u.Designation,
		"updatedAt":   u.UpdatedAt,
	}
	var out model.User
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": u.ID}, bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		return mapErr(err)
	}
	*u = out
	return nil
}

func (r *MongoUserRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, s model.Status, updatedAt time.Time) (*model.User, error) {
	var out model.User
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": s, "updatedAt": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		return nil, mapErr(err)
	}
	return &out, nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return mapErr(err)
	}
	if res.DeletedCount == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *MongoUserRepository) List(ctx context.Context, skip, limit int64, search string, status model.Status) ([]model.User, int64, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	if search != "" {
		// search is already regex-escaped by the service
		rx := bson.Regex{Pattern: search, Options: "i"}
		filter["$or"] = bson.A{
			bson.M{"firstName": rx},
			bson.M{"lastName": rx},
			bson.M{"email": rx},
		}
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, mapErr(err)
	}
	cur, err := r.coll.Find(ctx, filter, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(skip).SetLimit(limit))
	if err != nil {
		return nil, 0, mapErr(err)
	}
	items := []model.User{}
	if err := cur.All(ctx, &items); err != nil {
		return nil, 0, mapErr(err)
	}
	return items, total, nil
}

func (r *MongoUserRepository) Stats(ctx context.Context) (model.Stats, error) {
	var s model.Stats
	var err error
	if s.Total, err = r.coll.CountDocuments(ctx, bson.M{}); err != nil {
		return s, mapErr(err)
	}
	if s.Active, err = r.coll.CountDocuments(ctx, bson.M{"status": model.StatusActive}); err != nil {
		return s, mapErr(err)
	}
	if s.Inactive, err = r.coll.CountDocuments(ctx, bson.M{"status": model.StatusInactive}); err != nil {
		return s, mapErr(err)
	}
	return s, nil
}

// mapErr translates driver errors into sentinel errors. Context errors pass
// through so callers can detect cancellation.
func mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, mongo.ErrNoDocuments):
		return apperror.ErrNotFound
	case mongo.IsDuplicateKeyError(err):
		return apperror.ErrDuplicateEmail
	case errors.Is(err, context.Canceled):
		return context.Canceled
	case errors.Is(err, context.DeadlineExceeded), mongo.IsTimeout(err), mongo.IsNetworkError(err),
		strings.Contains(err.Error(), "server selection"):
		return apperror.ErrUnavailable
	}
	return err
}
