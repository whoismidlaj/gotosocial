package timelines

import (
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"github.com/gin-gonic/gin"
)

// DirectTimelineGETHandler handles GET requests for the direct timeline.
// It retrieves the user's conversations and returns their last statuses.
func (m *Module) DirectTimelineGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, errWithCode := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	page, errWithCode := paging.ParseIDPage(c,
		1,  // min limit
		80, // max limit
		40, // default limit
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	resp, errWithCode := m.processor.Conversations().GetAll(
		c.Request.Context(),
		authed.Account,
		page,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// Format the items from conversations to status models.
	statuses := make([]interface{}, 0, len(resp.Items))
	for _, item := range resp.Items {
		if conv, ok := item.(*apimodel.Conversation); ok && conv.LastStatus != nil {
			statuses = append(statuses, conv.LastStatus)
		} else if convMap, ok := item.(map[string]interface{}); ok {
			// In case type assertion fails but it was serialized to a map/struct
			if lastStatus, exists := convMap["last_status"]; exists && lastStatus != nil {
				statuses = append(statuses, lastStatus)
			}
		}
	}

	if resp.LinkHeader != "" {
		c.Header("Link", resp.LinkHeader)
	}

	apiutil.JSON(c, http.StatusOK, statuses)
}
