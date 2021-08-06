package motto

type Route struct {
	id     int32
	uri    string
	method string
}

func (r *Route) ID() int32 {
	return r.id
}

func (r *Route) URI() string {
	return r.uri
}

func (r *Route) Method() string {
	return r.method
}

func NewRoute(id int32, uri string, method string) (route Route) {
	return Route{
		id:     id,
		uri:    uri,
		method: method,
	}
}
