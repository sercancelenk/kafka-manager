package manager

import (
	"fmt"
	"github.com/IBM/sarama"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/config"
)

type AdminClientService interface {
	GetAdminClient(clusterId string) (sarama.Client, sarama.ClusterAdmin, error)
}

type adminClientService struct {
	Clusters map[string]config.Cluster
}

func NewAdminClientService(clusters map[string]config.Cluster) *adminClientService {
	return &adminClientService{
		Clusters: clusters,
	}
}

func (s *adminClientService) GetAdminClient(clusterId string) (sarama.Client, sarama.ClusterAdmin, error) {
	// Create a new Sarama client to get partition offsets
	client, err := sarama.NewClient(s.Clusters[clusterId].Brokers, sarama.NewConfig())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client: %v", err)
	}

	adminClient, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cluster admin: %v", err)
	}
	return client, adminClient, nil
}
