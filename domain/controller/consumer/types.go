package consumer

import (
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/consumer"
)

type GetConsumersRequest struct {
	ClusterId string `json:"clusterId"`
	Topic     string `param:"topic"`
}

type GetConsumersResponse struct {
	Infos []consumer.ConsumerGroupInfo `json:"infos"`
}

type GetConsumerInfoResponse struct {
	Info consumer.ConsumerGroupInfo `json:"infos"`
}

//-----------

type GetConsumerInfoRequest struct {
	ClusterId string `json:"clusterId"`
	GroupId   string `param:"groupId"`
}

type GetConsumerGroupListRequest struct {
	ClusterId string `json:"clusterId"`
	Page      int    `param:"page"`
	Size      int    `param:"size"`
}

type GetConsumerGroupListResponse[T any] struct {
	Groups service.PaginatedResponse[T] `json:"groups"`
}

//-----------
