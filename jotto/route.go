package jotto

// Route represents a HTTP or TCP route
type Route struct {
	id     uint32
	uri    string
	method string
	group  string
}

// ID returns the command ID (used in TCP router)
func (r *Route) ID() uint32 {
	return r.id
}

// URI returns the URI of this route (under HTTP mode)
func (r *Route) URI() string {
	return r.uri
}

// Method returns the HTTP method of this route
func (r *Route) Method() string {
	return r.method
}

// Group returns the API group of this route
func (r *Route) Group() string {
	return r.group
}

// NewRoute creates a new route
func NewRoute(id uint32, method string, uri string, group string) (route Route) {
	return Route{
		id:     id,
		uri:    uri,
		method: method,
		group:  group,
	}
}
