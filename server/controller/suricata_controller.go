package controller

import (
    "net/http"
    "time"

    "github.com/HUAHUAI23/simple-waf/server/service"
    "github.com/HUAHUAI23/simple-waf/server/utils/response"
    "github.com/gin-gonic/gin"
)

type SuricataController struct {
    svc service.SuricataService
}

func NewSuricataController(svc service.SuricataService) *SuricataController {
    return &SuricataController{svc: svc}
}

// 注册路由
func RegisterSuricataRoutes(r *gin.Engine) {
    repo := service.NewSuricataService(
        repository.NewSuricataRepository(),
    )
    ctrl := NewSuricataController(repo)

    api := r.Group("/api/v1/suricata")
    api.GET("/events", ctrl.ListEvents)
}

// 事件查询接口
func (s *SuricataController) ListEvents(c *gin.Context) {
    severity := c.Query("severity")
    srcIP := c.Query("src_ip")
    dstIP := c.Query("dst_ip")
    limit := int64(100)

    if l := c.Query("limit"); l != "" {
        if parsed, err := ParseInt64(l); err == nil {
            limit = parsed
        }
    }

    // 时间范围
    start, _ := time.Parse(time.RFC3339, c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339)))
    end, _ := time.Parse(time.RFC3339, c.DefaultQuery("end", time.Now().Format(time.RFC3339)))

    events, err := s.svc.GetEvents(c.Request.Context(), severity, srcIP, dstIP, start, end, limit)
    if err != nil {
        response.InternalServerError(c, err, true)
        return
    }

    response.Success(c, "查询成功", events)
}

func ParseInt64(s string) (int64, error) {
    var i int64
    _, err := fmt.Sscan(s, &i)
    return i, err
}
