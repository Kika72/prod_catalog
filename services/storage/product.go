package storage

import (
	"context"

	"prod_catalog/services/data"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productStorage struct {
	collection *mongo.Collection
}

func New(collection *mongo.Collection) (Storage, error) {

	return productStorage{
		collection: collection,
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
	operations := make([]mongo.WriteModel, 0, len(products))

	for _, product := range products {
		operation := mongo.NewUpdateOneModel()
		operation.SetUpsert(true)
		operation.SetFilter(bson.M{"name": bson.M{"$eq": product.Name}})
		operation.SetUpdate(bson.M{
			"$currentDate": bson.M{"updated_at": true},
			"$set": bson.M{
				"price": product.Price,
			},
			"$setOnInsert": bson.M{
				"name": product.Name,
			},
			"$inc": bson.M{
				"updates_count": 1,
			},
		})
		operations = append(operations, operation)
	}

	_, err := p.collection.BulkWrite(ctx, operations)
	return err
}
