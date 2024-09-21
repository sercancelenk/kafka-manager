package topic

import (
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/topic"
)

type GetTopicInfoRequest struct {
	ClusterId string `json:"clusterId"`
	Topic     string `param:"topic"`
}
type GetTopicInfoResponse struct {
	Info *topic.TopicInfo `json:"info"`
}

type GetTopicListRequest struct {
	ClusterId string `json:"clusterId"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

type GetTopicListResponse struct {
	Topics service.PaginatedResponse[string] `json:"topics"`
}

type GetTopicMessagesRequest struct {
	ClusterId string `json:"clusterId"`
	Topic     string `param:"topic"`
	Size      int    `json:"size"`
}

type GetTopicMessagesResponse struct {
	Messages []*topic.MessageInfo `json:"messages"`
}
