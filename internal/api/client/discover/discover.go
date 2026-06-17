package discover

import (
	"net/http"

	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/router"
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

func (m *Module) Route(g router.Group) {
	g.GET("/v1/discover/posts", m.getDiscoverPosts)
	g.GET("/v1/discover/tags/trending", m.getDiscoverTags)
	g.GET("/v1/discover/accounts/popular", m.getDiscoverAccounts)
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
