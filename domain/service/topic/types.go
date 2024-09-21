package topic

import (
	"github.com/IBM/sarama"
	"time"
)

type TopicInfo struct {
	ClusterID      string            `json:"clusterId"`
	Name           string            `json:"name"`
	PartitionCount int               `json:"partitionCount"`
	RetentionDay   int64             `json:"retentionDay"`
	MessageCount   int64             `json:"messageCount"`
	Replicas       int               `json:"replicas"`
	Config         map[string]string `json:"config"`
}

type MessageWithTimestamp struct {
	Message   *sarama.ConsumerMessage
	Timestamp int64
}

type MessageHeader struct {
	Key   string
	Value string
}

type MessageInfo struct {
	Headers        []*MessageHeader // only set if kafka is version 0.11+
	Timestamp      time.Time        // only set if kafka is version 0.10+, inner message timestamp
	BlockTimestamp time.Time        // only set if kafka is version 0.10+, outer (compressed) block timestamp

	Key, Value string
	Topic      string
	Partition  int32
	Offset     int64
}
