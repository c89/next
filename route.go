package next

import (
	"net/http"
	"reflect"
	"regexp"
)

type Routes struct {
	data []Route
}
type Route struct {
	r           string
	cr          *regexp.Regexp
	method      string
	handler     reflect.Value
	httpHandler http.Handler
}

func NewRoutes() *Routes {
	return &Routes{}
}

func (rs *Routes) Add(r string, method string, handler interface{}) {
	cr, err := regexp.Compile(r)
	if err != nil {
		// TODO
		// s.Logger.Printf("Error in route regex %q\n", r)
		return
	}

	switch handler.(type) {
	case http.Handler:
		rs.data = append(rs.data, Route{r: r, cr: cr, method: method, httpHandler: handler.(http.Handler)})
	case reflect.Value:
		fv := handler.(reflect.Value)
		rs.data = append(rs.data, Route{r: r, cr: cr, method: method, handler: fv})
	default:
		fv := reflect.ValueOf(handler)
		rs.data = append(rs.data, Route{r: r, cr: cr, method: method, handler: fv})
	}
}

func (s *Routes) Match(r, method string) *Route {
	for i := 0; i < len(s.data); i++ {
		route := &s.data[i]
		cr := route.cr
		//if the methods don't match, skip this handler (except HEAD can be used in place of GET)
		if method != route.method && !(method == "HEAD" && route.method == "GET") {
			continue
		}
		if !cr.MatchString(r) {
			continue
		}
		match := cr.FindStringSubmatch(r)

		if len(match[0]) != len(r) {
			continue
		}

		return route
	}

	return nil
}
