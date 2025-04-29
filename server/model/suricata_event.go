package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// SuricataEvent 表示一条 Suricata 日志事件
type SuricataEvent struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
    Severity  string             `bson:"severity" json:"severity"`
    SrcIP     string             `bson:"src_ip,omitempty" json:"src_ip"`
    DstIP     string             `bson:"dst_ip,omitempty" json:"dst_ip"`
    Msg       string             `bson:"msg" json:"msg"`
    // 如果需要更多字段，可自行扩展
}
