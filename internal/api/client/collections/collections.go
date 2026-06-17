package collections

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
	g.GET("/v1/collections", m.getCollections)
}

func (m *Module) getCollections(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
