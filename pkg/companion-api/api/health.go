package api

import (
	"github.com/emicklei/go-restful"
	"net/http"
)

func (cApi *CompanionAPI) healthWS() (ws *restful.WebService) {
	ws = &restful.WebService{}
	ws.Route(ws.GET("/health/ok").
		Doc("Simple API health check").
		To(func(req *restful.Request, res *restful.Response) {
			res.WriteHeader(http.StatusOK)
		}))

	return
}