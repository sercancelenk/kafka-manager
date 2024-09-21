package consumer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/IBM/sarama"
	"sync"
	"sync/atomic"
)

func decodeMemberAssignment(assignmentData []byte) (map[string][]int32, error) {
	buf := bytes.NewReader(assignmentData)
	assignment := make(map[string][]int32)

	// Read version
	var version int16
	if err := binary.Read(buf, binary.BigEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %v", err)
	}

	// Read number of topics
	var topicCount int32
	if err := binary.Read(buf, binary.BigEndian, &topicCount); err != nil {
		return nil, fmt.Errorf("failed to read topic count: %v", err)
	}

	// Read each topic
	for i := 0; i < int(topicCount); i++ {
		// Read topic name length
		var topicLen int16
		if err := binary.Read(buf, binary.BigEndian, &topicLen); err != nil {
			return nil, fmt.Errorf("failed to read topic length: %v", err)
		}

		// Read topic name
		topicName := make([]byte, topicLen)
		if _, err := buf.Read(topicName); err != nil {
			return nil, fmt.Errorf("failed to read topic name: %v", err)
		}

		// Read number of partitions for the topic
		var partitionCount int32
		if err := binary.Read(buf, binary.BigEndian, &partitionCount); err != nil {
			return nil, fmt.Errorf("failed to read partition count: %v", err)
		}

		// Read each partition
		partitions := make([]int32, partitionCount)
		for j := 0; j < int(partitionCount); j++ {
			if err := binary.Read(buf, binary.BigEndian, &partitions[j]); err != nil {
				return nil, fmt.Errorf("failed to read partition: %v", err)
			}
		}

		// Add topic and partitions to the assignment
		assignment[string(topicName)] = partitions
	}

	return assignment, nil
}
func getConsumerGroupDescription(groupId string, admin sarama.ClusterAdmin) (*sarama.GroupDescription, error) {
	// Describe the consumer group
	groupDescriptions, err := admin.DescribeConsumerGroups([]string{groupId})
	if err != nil {
		return nil, fmt.Errorf("error describing consumer group: %w", err)
	}

	// Return the description of the first (and only) group
	if len(groupDescriptions) > 0 {
		return groupDescriptions[0], nil
	}

	return nil, fmt.Errorf("no description found for group: %s", groupId)
}
func getCommittedOffset(admin sarama.ClusterAdmin, groupId string, topic string, partition int32) (int64, error) {
	// Describe consumer group offsets for the specified partition
	offsets, err := admin.ListConsumerGroupOffsets(groupId, map[string][]int32{
		topic: {partition},
	})
	if err != nil {
		return 0, fmt.Errorf("error retrieving committed offset for group %s: %w", groupId, err)
	}

	// Fetch the offset and metadata for the specified partition
	offsetInfo, exists := offsets.Blocks[topic][partition]
	if !exists {
		return 0, fmt.Errorf("no offset information for topic %s, partition %d", topic, partition)
	}

	return offsetInfo.Offset, nil
}
func getLatestOffset(client sarama.Client, topic string, partition int32) (int64, error) {
	// Fetch the latest offset for the specified topic and partition
	latestOffset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return 0, fmt.Errorf("error retrieving latest offset for topic %s, partition %d: %w", topic, partition, err)
	}

	return latestOffset, nil
}
func getCoordinator(client sarama.Client, groupId string) (ConsumerGroupCoordinator, error) {
	// Use the client to find the coordinator for the given group ID
	coordinator, err := client.Coordinator(groupId)
	if err != nil {
		return ConsumerGroupCoordinator{}, fmt.Errorf("error finding coordinator for group %s: %w", groupId, err)
	}

	// Since we cannot fetch the rack directly, we only fill the ID, Host, and Port
	coordinatorStruct := ConsumerGroupCoordinator{
		ID:   int32(coordinator.ID()), // Convert ID to int32
		Host: coordinator.Addr(),      // Host address of the coordinator
		Port: 0,
		Rack: nil, // Rack is not directly retrievable here
	}

	return coordinatorStruct, nil
}
func getConsumerGroupInfo(clusterId string, groupDescription *sarama.GroupDescription, client sarama.Client, admin sarama.ClusterAdmin) ConsumerGroupInfo {
	var totalLag int64
	var totalAssignedPartitions int32
	podMap := sync.Map{} // Concurrent map for storing hosts (pod)
	membersByTopic := make(map[string][]ConsumerGroupMember)

	// Fill the coordinator data
	coordinator, _ := getCoordinator(client, groupDescription.GroupId)

	// Loop over group members and calculate lag, committed offsets, and other details

	for _, member := range groupDescription.Members {
		assignment, _ := decodeMemberAssignment(member.MemberAssignment)

		for topic, partitions := range assignment {
			for _, partition := range partitions {
				// Get the committed and latest offsets for the topic and partition
				committedOffset, _ := getCommittedOffset(admin, groupDescription.GroupId, topic, partition)
				latestOffset, _ := getLatestOffset(client, topic, partition)

				// Calculate lag
				lag := latestOffset - committedOffset
				atomic.AddInt64(&totalLag, lag)

				// Increment total assigned partitions
				atomic.AddInt32(&totalAssignedPartitions, 1)

				// Store the host in podMap (as a set)
				podMap.Store(member.ClientHost, struct{}{})

				// Create a ConsumerGroupMember object
				memberInfo := ConsumerGroupMember{
					Lag:             lag,
					ConsumerID:      member.ClientId,
					MemberID:        member.ClientId,
					ClientID:        member.ClientId,
					Host:            member.ClientHost,
					Topic:           topic,
					Partition:       partition,
					CommittedOffset: committedOffset,
					LatestOffset:    latestOffset,
				}

				// Group members by topic
				membersByTopic[topic] = append(membersByTopic[topic], memberInfo)
			}
		}
	}

	// Count the number of unique hosts (pods)
	podCount := 0
	podMap.Range(func(_, _ interface{}) bool {
		podCount++
		return true
	})

	// Create the final ConsumerGroupInfo object
	return ConsumerGroupInfo{
		GroupID:                 groupDescription.GroupId,
		Coordinator:             coordinator,
		State:                   groupDescription.State,
		PartitionAssignor:       groupDescription.Protocol,
		MembersByTopic:          membersByTopic,
		MemberCount:             podCount,
		PodCount:                podCount,
		AssignedTopicCount:      len(membersByTopic),
		AssignedPartitionsCount: int(totalAssignedPartitions),
		TotalLag:                totalLag,
	}
}
