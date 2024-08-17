package http

type RequestHandler func(ResponseWriter, *Request)

type handlerIdentifier struct {
	path   string
	method string
}

func NotFoundHandler(w ResponseWriter, r *Request) {
	_ = w.SetStatus(404)
}

func InternalServerErrorHandler(w ResponseWriter, r *Request) {
	_ = w.SetStatus(500)
}
