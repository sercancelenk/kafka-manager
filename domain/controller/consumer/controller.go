package consumer

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/consumer"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/http"
)

type ConsumerController interface {
	GetConsumersByTopic(ctx *fiber.Ctx, q *GetConsumersRequest) (*GetConsumersResponse, error)
}

type consumerController struct {
	s consumer.ConsumerService
}

func NewConsumerController(r fiber.Router, s consumer.ConsumerService) {
	c := &consumerController{s}
	r.Route("/v1/consumers", func(router fiber.Router) {
		router.Get("/list", http.Serve[GetConsumerGroupListRequest, GetConsumerGroupListResponse[string]](c.GetConsumerGroupList))
		router.Get("/info-by-topic", http.Serve[GetConsumersRequest, GetConsumersResponse](c.GetConsumersByTopic))
		router.Get("/info-by-group", http.Serve[GetConsumerInfoRequest, GetConsumerInfoResponse](c.GetConsumerInfoByGroupId))
	})
}

// GetConsumersByTopic	 godoc
// @Summary      GetConsumersByTopic
// @Description  get consumers of topic
// @Tags         consumer services
// @Accept       json
// @Produce      json
// @Param        clusterId  query  string         true  "ClusterId"
// @Param        topic  query  string         true  "Topic"
// @Success      200  {object}  GetConsumersResponse
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /consumers/info-by-topic [get]
func (c *consumerController) GetConsumersByTopic(ctx *fiber.Ctx, q *GetConsumersRequest) (*GetConsumersResponse, error) {
	consumerGroupInfos, err := c.s.GetConsumerGroupInfoByTopic(ctx, q.ClusterId, q.Topic)
	if err != nil {
		return nil, err
	}
	return &GetConsumersResponse{
		Infos: consumerGroupInfos,
	}, nil
}

// GetConsumerInfoByGroupId godoc
// @Summary      Get Consumer Info by Group Id
// @Description  get consumer group details
// @Tags         consumer services
// @Accept       json
// @Produce      json
// @Param        clusterId  query    string     true  "ClusterId"
// @Param        groupId    query    string     true  "GroupId"
// @Success      200  {object}  GetConsumerInfoResponse
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /consumers/info-by-group [get]
func (c *consumerController) GetConsumerInfoByGroupId(ctx *fiber.Ctx, q *GetConsumerInfoRequest) (*GetConsumerInfoResponse, error) {
	consumerGroupInfo, err := c.s.GetConsumerGroupInfoByGroup(ctx, q.ClusterId, q.GroupId)
	if err != nil {
		return nil, err
	}
	return &GetConsumerInfoResponse{
		Info: consumerGroupInfo,
	}, nil
}

// GetConsumerGroupList godoc
// @Summary      GetConsumerGroupList
// @Description  get consumer group list
// @Tags         consumer services
// @Accept       json
// @Produce      json
// @Param        clusterId  query    string     true  "ClusterId"
// @Param        page  query    string     true  "Page"
// @Param        size  query    string     true  "Size"
// @Success      200  {object}  []string
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /consumers/list [get]
func (c *consumerController) GetConsumerGroupList(ctx *fiber.Ctx, q *GetConsumerGroupListRequest) (*GetConsumerGroupListResponse[string], error) {
	groupList, err := c.s.GetConsumerGroupList(ctx, q.ClusterId, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetConsumerGroupListResponse[string]{Groups: groupList}, nil
}
