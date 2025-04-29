package internal

import (
	"context"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// LogStore 定义日志存储接口
type LogStore interface {
	Store(log model.WAFLog) error
	Start(ctx context.Context)
	Close()
}

// MongoLogStore MongoDB实现的日志存储
type MongoLogStore struct {
	mongo           *mongo.Client
	mongoDB         string
	mongoCollection string
	logChan         chan model.WAFLog
	logger          zerolog.Logger
}

const (
	defaultChannelSize = 1000 // 默认通道缓冲大小
)

// NewMongoLogStore 创建新的MongoDB日志存储器
func NewMongoLogStore(client *mongo.Client, database, collection string, logger zerolog.Logger) *MongoLogStore {
	return &MongoLogStore{
		mongo:           client,
		mongoDB:         database,
		mongoCollection: collection,
		logChan:         make(chan model.WAFLog, defaultChannelSize),
		logger:          logger,
	}
}

// Store 非阻塞地发送日志到存储通道
func (s *MongoLogStore) Store(log model.WAFLog) error {
	select {
	case s.logChan <- log:
		return nil
	default:
		// 通道已满，丢弃日志
		s.logger.Warn().Msg("log channel is full, dropping log entry")
		return nil
	}
}

// Start 启动日志存储处理循环
func (s *MongoLogStore) Start(ctx context.Context) {
	go s.processLogs(ctx)
}

// Close 关闭日志存储器
func (s *MongoLogStore) Close() {
	close(s.logChan)
}

// processLogs 处理日志存储循环
func (s *MongoLogStore) processLogs(ctx context.Context) {
	collection := s.mongo.Database(s.mongoDB).Collection(s.mongoCollection)

	for {
		select {
		case log, ok := <-s.logChan:
			if !ok {
				return // 通道已关闭
			}

			// 使用带超时的上下文进行存储操作
			storeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

			_, err := collection.InsertOne(storeCtx, log)
			cancel()

			if err != nil {
				s.logger.Error().Err(err).Msg("failed to save firewall log to MongoDB")
			}

		case <-ctx.Done():
			return
		}
	}
}
