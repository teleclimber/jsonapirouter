package jsonapirouter

import (
	"fmt"
	"net/http"

	"github.com/mfcochauxlaberge/jsonapi"
)

// JSONAPIRouteHandler responds to the request
type JSONAPIRouteHandler func(res http.ResponseWriter, httpReq *http.Request, apiReq *jsonapi.Request)

// How to structure this?
// Here's what we know from the Request:
// - Method (this gives us Get/Update/Delete)
// - ResType
// - whether it is a collection
// -
type handlerType int

const (
	getCollection handlerType = iota
	getResource
	getRelated
	getRelationships
	createResource
	updateResource
	updateRelationships
	deleteResource
)

// JSONAPIRouter directs jsonapi requests to handlers of your choice
type JSONAPIRouter struct {
	schema *jsonapi.Schema

	getCollectionHandlers map[string]JSONAPIRouteHandler
	getResourceHandlers   map[string]JSONAPIRouteHandler
	getRelatedHandlers    map[string]map[string]JSONAPIRouteHandler
	// ... all the types, possibly subdivided
}

// NewJSONAPIRouter returns a JSONAPIRouters initialized
func NewJSONAPIRouter(schema *jsonapi.Schema) *JSONAPIRouter {
	return &JSONAPIRouter{
		schema:                schema,
		getCollectionHandlers: make(map[string]JSONAPIRouteHandler),
		getResourceHandlers:   make(map[string]JSONAPIRouteHandler),
		getRelatedHandlers:    make(map[string]map[string]JSONAPIRouteHandler),
	}
}

// Handle the request
func (r *JSONAPIRouter) Handle(res http.ResponseWriter, req *http.Request, apiReq *jsonapi.Request) {
	hType := r.getHandleType(apiReq)
	handler, ok := r.getHandler(hType, apiReq)

	if !ok {
		// The actual return code depends a bit. See the spec
		http.Error(res, "not implemented", 500)
	} else {
		handler(res, req, apiReq)
	}
}

func (r *JSONAPIRouter) getHandleType(apiReq *jsonapi.Request) handlerType {
	u := apiReq.URL
	switch apiReq.Method {
	case http.MethodGet:
		if !u.IsCol && u.ResID != "" && u.RelKind == "" {
			return getResource
		} else if u.IsCol && u.RelKind == "" {
			return getCollection
		} else if u.RelKind == "related" {
			return getRelated
		} else if u.RelKind == "self" {
			return getRelationships
		} else {
			panic("i dont understand this url")
		}

	case http.MethodPatch:

	case http.MethodDelete:

	}

	panic("method not supported")
}

func (r *JSONAPIRouter) getHandler(hType handlerType, apiReq *jsonapi.Request) (handler JSONAPIRouteHandler, ok bool) {
	switch hType {
	case getCollection:
		handler, ok = r.getCollectionHandler(apiReq.URL.ResType)
	case getResource:
		handler, ok = r.getResourceHandler(apiReq.URL.ResType)
	case getRelated:
		handler, ok = r.getRelatedHandler(apiReq.URL.BelongsToFilter.Type, apiReq.URL.Rel.FromName)
	default:
		panic("not handled yet.")
	}
	return
}

// GetCollection GET /articles
func (r *JSONAPIRouter) GetCollection(resType string, handler JSONAPIRouteHandler) {
	_, ok := r.getCollectionHandlers[resType]
	if ok {
		panic("GetCollection handler already exists for " + resType)
	}
	r.getCollectionHandlers[resType] = handler
}
func (r *JSONAPIRouter) getCollectionHandler(resType string) (JSONAPIRouteHandler, bool) {
	h, ok := r.getCollectionHandlers[resType]
	return h, ok
}

// GetResource GET /articles/1
func (r *JSONAPIRouter) GetResource(resType string, handler JSONAPIRouteHandler) {
	_, ok := r.getResourceHandlers[resType]
	if ok {
		panic("GetResource handler already exists for " + resType)
	}
	r.getResourceHandlers[resType] = handler
}
func (r *JSONAPIRouter) getResourceHandler(resType string) (JSONAPIRouteHandler, bool) {
	h, ok := r.getResourceHandlers[resType]
	return h, ok
}

// GetRelated GET /articles/1/author
// This one is a bit weird because you're returning a relType, filtered by belonging to resType
func (r *JSONAPIRouter) GetRelated(resType string, relName string, handler JSONAPIRouteHandler) {
	_, ok := r.getRelatedHandlers[resType]
	if !ok {
		r.getRelatedHandlers[resType] = make(map[string]JSONAPIRouteHandler)
	}
	_, ok = r.getRelatedHandlers[resType][relName]
	if ok {
		panic(fmt.Sprintf("GetRelated handler already exists for %v -> %v", resType, relName))
	}
	r.getRelatedHandlers[resType][relName] = handler
}
func (r *JSONAPIRouter) getRelatedHandler(resType string, relName string) (handler JSONAPIRouteHandler, ok bool) {
	handlers, ok := r.getRelatedHandlers[resType]
	if !ok {
		return
	}
	handler, ok = handlers[relName]
	return
}

// GetRelationships GET /articles/1/relationships/author
func (r *JSONAPIRouter) GetRelationships(resType string, relType string, handler JSONAPIRouteHandler) {

}

// CreateResource POST /articles
func (r *JSONAPIRouter) CreateResource(resType string, handler JSONAPIRouteHandler) {

}

// UpdateResource PATCH /articles/1
func (r *JSONAPIRouter) UpdateResource(resType string, handler JSONAPIRouteHandler) {

}

// UpdateRelationships PATCH /articles/1/relationships/author
func (r *JSONAPIRouter) UpdateRelationships(resType string, relType string, handler JSONAPIRouteHandler) {

}

// DeleteResource DELETE /articles/1
func (r *JSONAPIRouter) DeleteResource(resType string, handler JSONAPIRouteHandler) {

}
