package collections

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
	attachHandler(http.MethodGet, "/v1/collections", m.getCollections)
}

func (m *Module) getCollections(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
