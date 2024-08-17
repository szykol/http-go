package http

type RequestHandler func(ResponseWriter, *Request)

type handlerIdentifier struct {
	path   string
	method string
}
