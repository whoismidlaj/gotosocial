package discover

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/gin-gonic/gin"
)

type Module struct {
	processor *processing.Processor
}

func New(processor *processing.Processor) *Module {
	return &Module{
		processor: processor,
	}
}

func (m *Module) Route(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, "/v1/discover/posts", m.getDiscoverPosts)
	attachHandler(http.MethodGet, "/v1/discover/tags/trending", m.getDiscoverTags)
	attachHandler(http.MethodGet, "/v1/discover/accounts/popular", m.getDiscoverAccounts)
}

func (m *Module) getDiscoverPosts(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (m *Module) getDiscoverTags(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (m *Module) getDiscoverAccounts(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
