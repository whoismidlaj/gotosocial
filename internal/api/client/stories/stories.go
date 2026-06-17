package stories

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
	g.GET("/v1/stories", m.getStories)
	g.GET("/v1/accounts/:id/stories", m.getAccountStories)
}

func (m *Module) getStories(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (m *Module) getAccountStories(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
