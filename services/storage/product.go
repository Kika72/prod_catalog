package storage

import (
	"context"

	"prod_catalog/services/data"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	namespaceExistsErrCode int32 = 48
)

type productStorage struct {
	collection *mongo.Collection
	cln        *mongo.Client
}

func New(ctx context.Context, cln *mongo.Client, collection *mongo.Collection) (Storage, error) {
	if err := collection.Database().CreateCollection(context.Background(), collection.Name()); err != nil {
		mErr, ok := err.(mongo.CommandError)
		if !(ok && mErr.Code == namespaceExistsErrCode) {
			return nil, err
		}
	}

	idxOptions := &options.IndexOptions{}
	idxOptions.
		SetName(collection.Name() + "_uq").
		SetUnique(true)

	idxModel := mongo.IndexModel{
		Keys: bson.D{
			{"name", 1},
		},
		Options: idxOptions,
	}
	collection.Indexes().CreateOne(ctx, idxModel)

	return productStorage{
		collection: collection,
		cln:        cln,
	}, nil
}

func (p productStorage) List(ctx context.Context, order []data.OrderParam, offset, limit int64) ([]data.Product, error) {
	ord := make(bson.D, len(order))
	for i, o := range order {
		ord[i] = primitive.E{
			Key:   o.FieldName,
			Value: o.Direction,
		}
	}

	opt := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(ord)

	cursor, err := p.collection.Find(ctx, primitive.M{}, opt)
	if err != nil {
		return nil, err
	}

	result := make([]data.Product, 0)
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p productStorage) Store(ctx context.Context, products ...data.Product) error {
	p.collection.Database()
	operations := make([]mongo.WriteModel, 0, len(products))

	sess, err := p.cln.StartSession()
	if err != nil {
		return err
	}
	_, err = sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		for _, product := range products {
			operation := mongo.NewUpdateOneModel().
				SetUpsert(true).
				SetFilter(bson.M{
					"name": bson.M{"$eq": product.Name},
				}).
				SetUpdate(bson.M{
					//"$currentDate": bson.M{"updated_at": true},
					"$set": bson.M{
						"lock": primitive.NewObjectID(),
					},
					"$setOnInsert": bson.M{
						"name":          product.Name,
						"price":         -1,
						"updates_count": 0,
					},
				})
			operations = append(operations, operation)
		}

		_, err = p.collection.BulkWrite(ctx, operations)
		if err != nil {
			return nil, err
		}

		operations = operations[:0]
		for _, product := range products {
			operation := mongo.NewUpdateOneModel().
				SetFilter(bson.M{
					"name":  bson.M{"$eq": product.Name},
					"price": bson.M{"$ne": product.Price},
				}).
				SetUpdate(bson.M{
					"$currentDate": bson.M{"updated_at": true},
					"$set": bson.M{
						"price": product.Price,
					},
					"$inc": bson.M{
						"updates_count": 1,
					},
				})
			operations = append(operations, operation)
		}
		_, err = p.collection.BulkWrite(ctx, operations)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return err
}
