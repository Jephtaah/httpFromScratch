package server

type Handler func(method, path string, headers map[string]string, body []byte) (int, string)

var routes = map[string]Handler{
	"GET /health": func(method, path string, headers map[string]string, body []byte) (int, string) {
		return 200, "OK"
	},
	"POST /echo": func(method, path string, headers map[string]string, body []byte) (int, string) {
		return 200, string(body)
	},
}

func dispatch(req *Request) (int, string) {
	handler, found := routes[req.Method+" "+req.Path]
	if !found {
		return 404, "Not Found"
	}
	return handler(req.Method, req.Path, req.Headers, req.Body)
}
