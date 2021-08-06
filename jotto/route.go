package jotto

type Route struct {
	id     uint32
	uri    string
	method string
}

func (r *Route) ID() uint32 {
	return r.id
}

func (r *Route) URI() string {
	return r.uri
}

func (r *Route) Method() string {
	return r.method
}

func NewRoute(id uint32, method string, uri string) (route Route) {
	return Route{
		id:     id,
		uri:    uri,
		method: method,
	}
}
