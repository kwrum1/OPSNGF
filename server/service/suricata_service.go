package service

import (
    "context"
    "time"

    "github.com/HUAHUAI23/simple-waf/server/model"
    "github.com/HUAHUAI23/simple-waf/server/repository"
    "go.mongodb.org/mongo-driver/bson"
)

type SuricataService interface {
    GetEvents(ctx context.Context, severity, srcIP, dstIP string, start, end time.Time, limit int64) ([]model.SuricataEvent, error)
}

type suricataServiceImpl struct {
    repo repository.SuricataRepository
}

func NewSuricataService(repo repository.SuricataRepository) SuricataService {
    return &suricataServiceImpl{repo: repo}
}

func (s *suricataServiceImpl) GetEvents(ctx context.Context, severity, srcIP, dstIP string, start, end time.Time, limit int64) ([]model.SuricataEvent, error) {
    filter := bson.M{
        "timestamp": bson.M{"$gte": start, "$lte": end},
    }

    if severity != "" {
        filter["severity"] = severity
    }
    if srcIP != "" {
        filter["src_ip"] = srcIP
    }
    if dstIP != "" {
        filter["dst_ip"] = dstIP
    }

    return s.repo.FindEvents(ctx, filter, limit)
}
