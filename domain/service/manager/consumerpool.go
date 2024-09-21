package manager

import (
	"fmt"
	"github.com/IBM/sarama"
	"sync"
)

type ConsumerPoolManager struct {
	consumerPoolMap map[string]ConsumerPool // map[cluster][]*sarama.Consumer
}

func NewConsumerPoolManager(consumerPoolMap map[string]ConsumerPool) *ConsumerPoolManager {
	return &ConsumerPoolManager{
		consumerPoolMap: consumerPoolMap,
	}
}

// GetConsumerByCluster retrieves the first available consumer for a specific cluster
func (pool *ConsumerPoolManager) GetConsumerPool(cluster string) (ConsumerPool, error) {
	clusterPool, exists := pool.consumerPoolMap[cluster]
	if !exists {
		return nil, fmt.Errorf("no consumers found for cluster %s", cluster)
	}

	return clusterPool, nil
}

type ConsumerPool interface {
	GetConsumer() (*sarama.Consumer, error)
	ReturnConsumer(consumer *sarama.Consumer)
}

// ConsumerPool to hold reusable consumers
type consumerPool struct {
	pool    []*sarama.Consumer
	mu      sync.Mutex
	config  *sarama.Config
	brokers []string
}

// NewConsumerPool creates a new pool of consumers
func NewConsumerPool(size int, brokers []string, config *sarama.Config) (*consumerPool, error) {
	cp := &consumerPool{
		pool:    make([]*sarama.Consumer, 0, size),
		config:  config,
		brokers: brokers,
	}
	for i := 0; i < size; i++ {
		consumer, err := sarama.NewConsumer(brokers, config)
		if err != nil {
			return nil, fmt.Errorf("error creating consumer: %v", err)
		}
		cp.pool = append(cp.pool, &consumer)
	}
	return cp, nil
}

func (cp *consumerPool) GetConsumer() (*sarama.Consumer, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if len(cp.pool) == 0 {
		return nil, fmt.Errorf("no available consumers in the pool")
	}

	consumer := cp.pool[0]
	cp.pool = cp.pool[1:]
	return consumer, nil
}

// ReturnConsumer returns a consumer to the pool
func (cp *consumerPool) ReturnConsumer(consumer *sarama.Consumer) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.pool = append(cp.pool, consumer)
}
