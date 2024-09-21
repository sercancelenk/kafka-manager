package topic

import (
	"fmt"
	"github.com/IBM/sarama"
)

type OffsetSpec int64

const (
	EarliestOffset OffsetSpec = OffsetSpec(sarama.OffsetOldest)
	LatestOffset   OffsetSpec = OffsetSpec(sarama.OffsetNewest)
)

func listOffsets(client sarama.Client, topic string, partition int32, offsetSpec OffsetSpec) (int64, error) {
	offset, err := client.GetOffset(topic, partition, int64(offsetSpec))
	if err != nil {
		return 0, fmt.Errorf("failed to get offset for topic %s, partition %d: %w", topic, partition, err)
	}
	return offset, nil
}

func describeTopicConfig(admin sarama.ClusterAdmin, topic string) (map[string]string, error) {
	// Define the config resource (topic) to fetch its configuration
	configResource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topic,
	}

	// Describe the topic configuration
	configs, err := admin.DescribeConfig(configResource)
	if err != nil {
		return nil, fmt.Errorf("failed to describe configs for topic %s: %w", topic, err)
	}

	// Create a map to store the config entries
	topicConfig := make(map[string]string)

	// Iterate over the config entries and populate the map
	for _, entry := range configs {
		topicConfig[entry.Name] = entry.Value
	}

	return topicConfig, nil
}

func describeTopic(clusterId, topic string, admin sarama.ClusterAdmin) (*sarama.TopicMetadata, error) {
	// Placeholder for actual topic description logic
	metadata, err := admin.DescribeTopics([]string{topic})
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		return metadata[0], nil
	}
	return nil, fmt.Errorf("topic %s not found", topic)
}

func getBrokerByID(client sarama.Client, brokerID int32) *sarama.Broker {
	brokers := client.Brokers()
	for _, broker := range brokers {
		if broker.ID() == brokerID {
			return broker
		}
	}
	return nil
}

func getTopicList(client sarama.Client) ([]string, error) {
	topics, err := client.Topics()
	if err != nil {
		return nil, err
	}
	return topics, nil
}
