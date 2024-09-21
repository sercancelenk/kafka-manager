package topic

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/domain/service/topic"
	"gitlab.trendyol.com/platform/messaging/kafka/kafka-stream-api/internal/http"
)

type TopicController interface {
	GetTopicInfo(ctx *fiber.Ctx, q *GetTopicInfoRequest) (*GetTopicInfoResponse, error)
	GetTopicList(ctx *fiber.Ctx, q *GetTopicListRequest) (*GetTopicListResponse, error)
}

type topicController struct {
	s topic.TopicService
}

func NewTopicController(r fiber.Router, s topic.TopicService) {
	c := &topicController{s}
	r.Route("/v1/topics", func(router fiber.Router) {
		router.Get("/", http.Serve[GetTopicListRequest, GetTopicListResponse](c.GetTopicList))
		router.Get("/info", http.Serve[GetTopicInfoRequest, GetTopicInfoResponse](c.GetTopicInfo))
		router.Get("/messages", http.Serve[GetTopicMessagesRequest, GetTopicMessagesResponse](c.GetTopMessages))
	})
}

// GetTopicInfo	 godoc
// @Summary      GetTopicInfo
// @Description  get topic info
// @Tags         topic services
// @Accept       json
// @Produce      json
// @Param        clusterId  query  string         true  "ClusterId"
// @Param        topic  query  string         true  "Topic"
// @Success      200  {object}  GetTopicInfoResponse
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /topics/info [get]
func (c *topicController) GetTopicInfo(ctx *fiber.Ctx, q *GetTopicInfoRequest) (*GetTopicInfoResponse, error) {
	info, err := c.s.GetTopicInfo(ctx, q.ClusterId, q.Topic)
	if err != nil {
		return nil, err
	}
	return &GetTopicInfoResponse{Info: info}, nil
}

// GetTopicList	 godoc
// @Summary      GetTopicList
// @Description  get topic list
// @Tags         topic services
// @Accept       json
// @Produce      json
// @Param        clusterId  query  string         true  "ClusterId"
// @Param        page  query    string     true  "Page"
// @Param        size  query    string     true  "Size"
// @Success      200  {object}  GetTopicListResponse
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /topics [get]
func (c *topicController) GetTopicList(ctx *fiber.Ctx, q *GetTopicListRequest) (*GetTopicListResponse, error) {
	paginatedList, err := c.s.GetTopicList(ctx, q.ClusterId, q.Page, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetTopicListResponse{
		Topics: paginatedList,
	}, nil
}

// GetTopMessages	 godoc
// @Summary      GetTopMessages
// @Description  get top messages
// @Tags         topic services
// @Accept       json
// @Produce      json
// @Param        clusterId  query  string         true  "ClusterId"
// @Param        topic  query  string         true  "Topic"
// @Param        size  query    string     true  "Size"
// @Success      200  {object}  GetTopicMessagesResponse
// @Failure      400  {object}  http.Error
// @Failure      404  {object}  http.Error
// @Failure      500  {object}  http.Error
// @Router       /topics/messages [get]
func (c *topicController) GetTopMessages(ctx *fiber.Ctx, q *GetTopicMessagesRequest) (*GetTopicMessagesResponse, error) {
	messages, err := c.s.FetchTopMessagesFromAllPartitions(ctx, q.ClusterId, q.Topic, q.Size)
	if err != nil {
		return nil, err
	}
	return &GetTopicMessagesResponse{
		Messages: messages,
	}, nil
}
