package consumer

type ConsumerGroupCoordinator struct {
	ID       int32
	IDString string
	Host     string
	Port     int32
	Rack     *string
}

type ConsumerGroupMember struct {
	Lag             int64
	ConsumerID      string
	MemberID        string
	ClientID        string
	Host            string
	Topic           string
	Partition       int32
	CommittedOffset int64
	LatestOffset    int64
}

type ConsumerGroupInfo struct {
	GroupID                 string
	Coordinator             ConsumerGroupCoordinator
	State                   string
	PartitionAssignor       string
	MembersByTopic          map[string][]ConsumerGroupMember
	MemberCount             int
	PodCount                int
	AssignedTopicCount      int
	AssignedPartitionsCount int
	TotalLag                int64
}
