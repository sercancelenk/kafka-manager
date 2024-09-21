package topic

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/manager"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/config"
	"go.uber.org/zap"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

type TopicService interface {
	GetTopicList(ctx *fiber.Ctx, clusterId string, page, size int) (service.PaginatedResponse[string], error)
	GetTopicInfo(ctx *fiber.Ctx, clusterId, topic string) (*TopicInfo, error)
	FetchTopMessagesFromAllPartitions(ctx *fiber.Ctx, clusterId, topic string, size int) ([]*MessageInfo, error)
}

type topicService struct {
	ConsumerPoolManager *manager.ConsumerPoolManager
	AdminClientService  manager.AdminClientService
}

func NewTopicService(logger *zap.Logger, cfg *config.ApplicationConfig, consumerPoolManager *manager.ConsumerPoolManager, adminClientService manager.AdminClientService) TopicService {
	return &topicService{
		ConsumerPoolManager: consumerPoolManager,
		AdminClientService:  adminClientService,
	}
}

func (t *topicService) GetTopicList(ctx *fiber.Ctx, clusterId string, page, size int) (service.PaginatedResponse[string], error) {
	client, admin, err := t.AdminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return service.PaginatedResponse[string]{}, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)

	topicList, err := getTopicList(client)
	if err != nil {
		return service.PaginatedResponse[string]{}, err
	}

	totalItems := len(topicList)
	totalPages := int(math.Ceil(float64(totalItems) / float64(size)))

	// Calculate start and end indices for the pagination
	start := int(math.Min(float64(page*size), float64(totalItems)))
	end := int(math.Min(float64((page*size)+size), float64(totalItems)))

	// Get the paginated slice
	paginatedConsumerGroupIds := topicList[start:end]

	// Return paginated response

	return service.PaginatedResponse[string]{
		Items:      paginatedConsumerGroupIds,
		Page:       page,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

func (t *topicService) GetTopicInfo(ctx *fiber.Ctx, clusterId, topic string) (*TopicInfo, error) {
	client, admin, err := t.AdminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return &TopicInfo{}, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)

	var topicInfo TopicInfo
	topicInfo.ClusterID = clusterId
	topicInfo.Name = topic

	// Get the topic description to fetch partition information
	topicDescription, err := describeTopic(clusterId, topic, admin)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}
	topicInfo.PartitionCount = len(topicDescription.Partitions)

	// Fetch the retention period from topic configuration
	topicConfig, err := describeTopicConfig(admin, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic config: %w", err)
	}
	retentionMs, _ := strconv.ParseInt(topicConfig["retention.ms"], 10, 64)
	topicInfo.RetentionDay = retentionMs / (1000 * 60 * 60 * 24) // Convert retention from milliseconds to days

	// Calculate total message count and replica information
	var totalMessageCount int64
	replicas := sync.Map{}

	for _, partitionInfo := range topicDescription.Partitions {
		partition := partitionInfo.ID

		// Get earliest and latest offsets for the partition
		earliestOffset, err := listOffsets(client, topic, partition, OffsetSpec(sarama.OffsetOldest))
		if err != nil {
			return nil, fmt.Errorf("failed to get earliest offset for partition %d: %w", partition, err)
		}
		latestOffset, err := listOffsets(client, topic, partition, OffsetSpec(sarama.OffsetNewest))
		if err != nil {
			return nil, fmt.Errorf("failed to get latest offset for partition %d: %w", partition, err)
		}

		// Calculate the message count for the partition
		partitionMessageCount := latestOffset - earliestOffset
		atomic.AddInt64(&totalMessageCount, partitionMessageCount)

		// Store replica information
		for _, replica := range partitionInfo.Replicas {
			broker := getBrokerByID(client, replica)
			if broker != nil {
				replicaAddr := fmt.Sprintf("%s:%d", broker.Addr(), 0)
				replicas.Store(replicaAddr, true)
			}
		}
	}

	// Set total message count and replicas in the topic info
	topicInfo.MessageCount = totalMessageCount

	// Count unique replicas
	replicaCount := 0
	replicas.Range(func(_, _ interface{}) bool {
		replicaCount++
		return true
	})
	topicInfo.Replicas = replicaCount

	// Set topic configuration
	topicInfo.Config = topicConfig

	return &topicInfo, nil
}

func (t *topicService) FetchTopMessagesFromAllPartitions(ctx *fiber.Ctx,
	clusterId, topic string,

	size int) ([]*MessageInfo, error) {
	client, admin, err := t.AdminClientService.GetAdminClient(clusterId)
	if admin == nil || client == nil || err != nil {
		return []*MessageInfo{}, err
	}
	defer func(client sarama.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close client: %v", err)
		}
	}(client)

	pool, err := t.ConsumerPoolManager.GetConsumerPool(clusterId)
	if err != nil {
		return []*MessageInfo{}, err
	}
	consumer, err := pool.GetConsumer()
	if err != nil {
		return []*MessageInfo{}, err
	}
	defer pool.ReturnConsumer(consumer)

	// Get partitions for the topic
	partitions, err := (*consumer).Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("error fetching partitions: %w", err)
	}

	// Use a slice to collect messages and protect it with a mutex for concurrent access
	var allMessages []MessageWithTimestamp
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, partition := range partitions {
		wg.Add(1)

		go func(partition int32) {
			defer wg.Done()

			// Get the latest offset for each partition to know when to stop consuming
			latestOffset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				fmt.Printf("Error fetching latest offset for partition %d: %v\n", partition, err)
				return
			}

			startOffset, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
			if err != nil {
				fmt.Printf("Error fetching latest offset for partition %d: %v\n", partition, err)
				return
			}

			if latestOffset >= int64(size) {
				startOffset = latestOffset - int64(size)
			}

			// Start consuming from the determined offset
			partitionConsumer, err := (*consumer).ConsumePartition(topic, partition, startOffset)
			if err != nil {
				fmt.Printf("Error creating partition consumer for partition %d: %v\n", partition, err)
				return
			}
			defer func(partitionConsumer sarama.PartitionConsumer) {
				err := partitionConsumer.Close()
				if err != nil {
					fmt.Printf("Failed to close partition consumer for partition %d: %v\n", partition, err)
				}
			}(partitionConsumer)

			for msg := range partitionConsumer.Messages() {
				mu.Lock() // Protect concurrent access to the allMessages slice
				allMessages = append(allMessages, MessageWithTimestamp{
					Message:   msg,
					Timestamp: msg.Timestamp.UnixNano(), // Use the message timestamp (nanoseconds)
				})
				mu.Unlock()

				// Stop consuming if we reach the latest offset
				if msg.Offset+1 >= latestOffset {
					break
				}
			}
		}(partition)
	}

	// Wait for all partitions to finish consuming
	wg.Wait()

	// Sort messages by timestamp in descending order
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].Timestamp > allMessages[j].Timestamp
	})

	// Return only the top N messages
	topMessages := make([]*MessageInfo, 0, size)
	for i := 0; i < size && i < len(allMessages); i++ {
		m := allMessages[i].Message
		headers := make([]*MessageHeader, 0)
		for index, header := range m.Headers {
			key := string(header.Key)
			v := string(header.Value)
			headers[index] = &MessageHeader{
				Key:   key,
				Value: v,
			}
		}

		topMessages = append(topMessages, &MessageInfo{
			Headers:        headers,
			Timestamp:      m.Timestamp,
			BlockTimestamp: m.BlockTimestamp,
			Key:            string(m.Key),
			Value:          string(m.Value),
			Topic:          m.Topic,
			Partition:      m.Partition,
			Offset:         m.Offset,
		})
	}

	return topMessages, nil
}
