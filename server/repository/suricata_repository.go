package repository

import (
    "context"
    "time"

    "github.com/kwrum1/waf/server/config"
    "github.com/kwrum1/waf/server/model"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type SuricataRepository interface {
    FindEvents(ctx context.Context, filter bson.M, limit int64) ([]model.SuricataEvent, error)
}

type suricataRepositoryImpl struct{}

func NewSuricataRepository() SuricataRepository {
    return &suricataRepositoryImpl{}
}

func (r *suricataRepositoryImpl) FindEvents(ctx context.Context, filter bson.M, limit int64) ([]model.SuricataEvent, error) {
    collection := config.GetMongoClient().
        Database(config.DBName).
        Collection("suricata_events")

    findOpts := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(limit)

    cursor, err := collection.Find(ctx, filter, findOpts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var results []model.SuricataEvent
    if err = cursor.All(ctx, &results); err != nil {
        return nil, err
    }
    return results, nil
}
