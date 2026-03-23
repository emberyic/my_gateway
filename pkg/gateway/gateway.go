package gateway

import(
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"github.com/emberyic/my_gateway/pkg/config"
)

type Gateway struct{
	routes []config.Route
}

func NewGateway(routes []config.Route) *Gateway{
	return &Gateway{routes: routes}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request){
	for _, route := range g.routes{
		if strings.HasPrefix(r.URL.Path, route.Path){
			backendUrl, _ := url.Parse(route.Backend)

			proxy := httputil.NewSingleHostReverseProxy(backendUrl)

			proxy.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}