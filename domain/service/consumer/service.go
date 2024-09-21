package consumer

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/manager"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/config"
	"go.uber.org/zap"
	"log"
	"math"
)

//go:generate mockery --name TopicService --filename topic_service.go --output ../../mocks
type ConsumerService interface {
	GetConsumerGroupList(ctx *fiber.Ctx, clusterId string, page, size int) (service.PaginatedResponse[string], error)
	GetConsumerGroupInfoByTopic(ctx *fiber.Ctx, clusterId, topic string) ([]ConsumerGroupInfo, error)
	GetConsumerGroupInfoByGroup(ctx *fiber.Ctx, clusterId, group string) (ConsumerGroupInfo, error)
}

type consumerService struct {
	logger             *zap.Logger
	cfg                *config.ApplicationConfig
	adminClientService manager.AdminClientService
}

func NewConsumerService(
	logger *zap.Logger,
	cfg *config.ApplicationConfig,
	adminClientService manager.AdminClientService,
) (ConsumerService, error) {
	t := &consumerService{logger, cfg, adminClientService}
	return t, nil
}

// Get consumer group info by topic
func (c *consumerService) GetConsumerGroupInfoByTopic(ctx *fiber.Ctx, clusterId string, topic string) ([]ConsumerGroupInfo, error) {
	client, admin, err := c.adminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return nil, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)
	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, err
	}
	groupIdList := make([]string, len(groups))
	for groupID := range groups {
		groupIdList = append(groupIdList, groupID)
	}
	groupDescs, err := admin.DescribeConsumerGroups(groupIdList)
	if err != nil {
		log.Printf("Failed to describe consumer groups%v", err)
	}

	var consumerGroupInfos []ConsumerGroupInfo

	// Filter and extract consumer group info by topic
	for _, groupDescription := range groupDescs {
		info := getConsumerGroupInfo(clusterId, groupDescription, client, admin)
		_, exists := info.MembersByTopic[topic]
		if exists {
			consumerGroupInfos = append(consumerGroupInfos, info)
		}
	}

	return consumerGroupInfos, nil
}

func (c *consumerService) GetConsumerGroupInfoByGroup(ctx *fiber.Ctx, clusterId string, group string) (ConsumerGroupInfo, error) {
	client, admin, err := c.adminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return ConsumerGroupInfo{}, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)

	groupDesc, err := getConsumerGroupDescription(group, admin)
	if err != nil {
		log.Printf("Failed to describe consumer groups%v", err)
	}

	return getConsumerGroupInfo(clusterId, groupDesc, client, admin), nil
}

func (c *consumerService) GetConsumerGroupList(ctx *fiber.Ctx, clusterId string, page, size int) (service.PaginatedResponse[string], error) {
	client, admin, err := c.adminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return service.PaginatedResponse[string]{}, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)

	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return service.PaginatedResponse[string]{}, err
	}
	groupIdList := make([]string, 0)
	for groupId, _ := range groups {
		groupIdList = append(groupIdList, groupId)
	}

	totalItems := len(groupIdList)
	totalPages := int(math.Ceil(float64(totalItems) / float64(size)))

	// Calculate start and end indices for the pagination
	start := int(math.Min(float64(page*size), float64(totalItems)))
	end := int(math.Min(float64((page*size)+size), float64(totalItems)))

	// Get the paginated slice
	paginatedConsumerGroupIds := groupIdList[start:end]

	// Return paginated response

	return service.PaginatedResponse[string]{
		Items:      paginatedConsumerGroupIds,
		Page:       page,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}
