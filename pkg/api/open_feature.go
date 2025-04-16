package api

import (
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/util/proxyutil"
	"github.com/grafana/grafana/pkg/web"
	"net/http"
)

// TODO: figure out if new config item is needed for this:
// >we might need to be able to change the root (/foo/bar/ofrep/v1/...) - this'll likely just need to be a config item.
const section = ""

func (hs *HTTPServer) registerOpenFeatureRoutes(apiRoute routing.RouteRegister) {
	if hs.openFeature.ProviderType != featuremgmt.StaticProviderType {
		apiRoute.Group("/ofrep/v1", func(apiRoute routing.RouteRegister) {
			apiRoute.Post("/evaluate/flags", hs.handleProxyRequest)
			apiRoute.Post("/evaluate/flags/:flagKey", hs.handleProxyRequest)
		})
	} else {
		apiRoute.Group("/ofrep/v1", func(apiRoute routing.RouteRegister) {
			apiRoute.Post("/evaluate/flags", hs.allFlagsStaticProvider)
			apiRoute.Post("/evaluate/flags/:flagKey", hs.evalFlagStaticProvider)
		})
	}
}

func (hs *HTTPServer) handleProxyRequest(c *contextmodel.ReqContext) {
	proxyPath := c.Req.URL.Path
	if proxyPath == "" {
		c.JsonApiErr(http.StatusBadRequest, "proxy path is required", nil)
		return
	}

	u := hs.openFeature.URL
	director := func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.URL.Path = proxyPath
	}

	proxy := proxyutil.NewReverseProxy(c.Logger, director)
	proxy.ServeHTTP(c.Resp, c.Req)
}

func (hs *HTTPServer) evalFlagStaticProvider(c *contextmodel.ReqContext) response.Response {
	flagKey := web.Params(c.Req)[":flagKey"]
	if flagKey == "" {
		return response.Error(http.StatusBadRequest, "flagKey is required", nil)
	}

	flags, err := hs.openFeature.EvalFlagWithStaticProvider(c.Req.Context(), flagKey)
	if err != nil {
		return response.Error(http.StatusInternalServerError, "failed to evaluate feature flag", err)
	}

	return response.JSON(http.StatusOK, flags)
}

func (hs *HTTPServer) allFlagsStaticProvider(c *contextmodel.ReqContext) response.Response {
	flags, err := hs.openFeature.EvalAllFlagsWithStaticProvider(c.Req.Context())
	if err != nil {
		return response.Error(http.StatusInternalServerError, "failed to evaluate feature flags", err)
	}

	return response.JSON(http.StatusOK, flags)
}
