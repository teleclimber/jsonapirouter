package jsonapirouter

import (
	"errors"
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

	getCollectionHandlers    map[string]JSONAPIRouteHandler
	getResourceHandlers      map[string]JSONAPIRouteHandler
	getRelatedHandlers       map[string]map[string]JSONAPIRouteHandler
	getRelationshipsHandlers map[string]map[string]JSONAPIRouteHandler
	// ... all the types, possibly subdivided
}

// NewJSONAPIRouter returns a JSONAPIRouters initialized
func NewJSONAPIRouter(schema *jsonapi.Schema) *JSONAPIRouter {
	return &JSONAPIRouter{
		schema:                   schema,
		getCollectionHandlers:    make(map[string]JSONAPIRouteHandler),
		getResourceHandlers:      make(map[string]JSONAPIRouteHandler),
		getRelatedHandlers:       make(map[string]map[string]JSONAPIRouteHandler),
		getRelationshipsHandlers: make(map[string]map[string]JSONAPIRouteHandler),
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
		handler, ok = r.getCollectionHandler(apiReq.URL.ResType) // might replace these with getHandler(r.cetCollenctionHandlers, ...)
	case getResource:
		handler, ok = r.getResourceHandler(apiReq.URL.ResType)
	case getRelated:
		handler, ok = r.getRelatedHandler(apiReq.URL.BelongsToFilter.Type, apiReq.URL.Rel.FromName)
	case getRelationships:
		handler, ok = r.getRelationshipsHandler(apiReq.URL.BelongsToFilter.Type, apiReq.URL.Rel.FromName)
	// TODO: other handlers...
	default:
		panic("not handled yet.")
	}
	return
}

// GetCollection sets the handler for requests like
// GET /articles
func (r *JSONAPIRouter) GetCollection(resType string, handler JSONAPIRouteHandler) {
	err := setHandler(r.getCollectionHandlers, resType, handler)
	if err != nil {
		panic(err)
	}
}
func (r *JSONAPIRouter) getCollectionHandler(resType string) (JSONAPIRouteHandler, bool) {
	h, ok := r.getCollectionHandlers[resType]
	return h, ok
}

// GetResource sets the handler for requests like
// GET /articles/1
func (r *JSONAPIRouter) GetResource(resType string, handler JSONAPIRouteHandler) {
	err := setHandler(r.getResourceHandlers, resType, handler)
	if err != nil {
		panic(err)
	}
}
func (r *JSONAPIRouter) getResourceHandler(resType string) (JSONAPIRouteHandler, bool) {
	h, ok := r.getResourceHandlers[resType]
	return h, ok
}

// GetRelated sets the handler for requests like
// GET /articles/1/author
// This one is a bit weird because you're returning a relType, filtered by belonging to resType
func (r *JSONAPIRouter) GetRelated(resType string, relName string, handler JSONAPIRouteHandler) {
	err := setHandlerRel(r.getRelatedHandlers, resType, relName, handler)
	if err != nil {
		panic(err)
	}
}
func (r *JSONAPIRouter) getRelatedHandler(resType string, relName string) (handler JSONAPIRouteHandler, ok bool) {
	handler, ok = getHandlerRel(r.getRelatedHandlers, resType, relName)
	return
}

// GetRelationships sets the handler for requests like
// GET /articles/1/relationships/author
func (r *JSONAPIRouter) GetRelationships(resType string, relName string, handler JSONAPIRouteHandler) {
	err := setHandlerRel(r.getRelationshipsHandlers, resType, relName, handler)
	if err != nil {
		panic(err)
	}
}
func (r *JSONAPIRouter) getRelationshipsHandler(resType string, relName string) (handler JSONAPIRouteHandler, ok bool) {
	handler, ok = getHandlerRel(r.getRelationshipsHandlers, resType, relName)
	return
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

// common handler management code

func setHandler(handlers map[string]JSONAPIRouteHandler, resType string, handler JSONAPIRouteHandler) error {
	_, ok := handlers[resType]
	if ok {
		return errors.New("GetResource handler already exists for " + resType) // sentinel error so we can deal more effectively?
	}
	handlers[resType] = handler
	return nil
}

func setHandlerRel(handlers map[string]map[string]JSONAPIRouteHandler, resType string, relName string, handler JSONAPIRouteHandler) error {
	_, ok := handlers[resType]
	if !ok {
		handlers[resType] = make(map[string]JSONAPIRouteHandler)
	}
	_, ok = handlers[resType][relName]
	if ok {
		return fmt.Errorf("handler already exists for %v -> %v", resType, relName)
	}
	handlers[resType][relName] = handler
	return nil
}
func getHandlerRel(handlers map[string]map[string]JSONAPIRouteHandler, resType string, relName string) (handler JSONAPIRouteHandler, ok bool) {
	resTypeHandlers, ok := handlers[resType]
	if !ok {
		return
	}
	handler, ok = resTypeHandlers[relName]
	return
}
