package jsonapirouter

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mfcochauxlaberge/jsonapi"
)

type RouterReq struct {
	URL      *jsonapi.URL
	Doc      *jsonapi.Document
	Includes *Includes
	// we could add errors here
}

type Status int

const (
	OK Status = iota
	NotFound
	Unauthorized
	Error
	// check json:api spec
)

// JSONAPIRouteHandler responds to the request
type JSONAPIRouteHandler func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status

// JSONAPIDataLoader functions return resources for the provided ids
type JSONAPIDataLoader func(ids []string, rReq *RouterReq) ([]jsonapi.Resource, error)

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

	loaders map[string]JSONAPIDataLoader

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
		loaders:                  make(map[string]JSONAPIDataLoader),
		getCollectionHandlers:    make(map[string]JSONAPIRouteHandler),
		getResourceHandlers:      make(map[string]JSONAPIRouteHandler),
		getRelatedHandlers:       make(map[string]map[string]JSONAPIRouteHandler),
		getRelationshipsHandlers: make(map[string]map[string]JSONAPIRouteHandler),
	}
}

// Handle the request
func (r *JSONAPIRouter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	rReq := &RouterReq{}

	su, err := jsonapi.NewSimpleURL(req.URL)
	if err != nil {
		//return nil, err
	}

	rReq.URL, err = jsonapi.NewURL(r.schema, su)
	if err != nil {
		//todo
	}

	rReq.Doc = &jsonapi.Document{}

	rReq.Includes = NewIncludes(rReq.Doc)

	hType := r.getHandleType(req.Method, rReq.URL)
	handler, ok := r.getHandler(hType, rReq.URL)

	if !ok {
		// The actual return code depends a bit. See the spec
		http.Error(res, "not implemented", 500)
		return
	}

	status := handler(res, req, rReq)
	if status == Error {
		http.Error(res, "some error", 500)
		return
	} else if status == Unauthorized {
		http.Error(res, "unauthorized", 403)
		return
	}
	// more ways to handle, and do it appropriately wrt json:api spec

	rReq.Includes.extractIDs(rReq)

	err = r.loadIncludes(rReq)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	rReq.Includes.AddAllToDoc()

	payload, err := jsonapi.MarshalDocument(rReq.Doc, rReq.URL)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	res.Header().Set("Content-Type", "application/json") // this is the wrong content type for json api.
	res.Write(payload)

}

func (r *JSONAPIRouter) getHandleType(method string, u *jsonapi.URL) handlerType {
	switch method {
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

func (r *JSONAPIRouter) getHandler(hType handlerType, u *jsonapi.URL) (handler JSONAPIRouteHandler, ok bool) {
	switch hType {
	case getCollection:
		handler, ok = r.getCollectionHandler(u.ResType) // might replace these with getHandler(r.cetCollenctionHandlers, ...)
	case getResource:
		handler, ok = r.getResourceHandler(u.ResType)
	case getRelated:
		handler, ok = r.getRelatedHandler(u.BelongsToFilter.Type, u.Rel.FromName)
	case getRelationships:
		handler, ok = r.getRelationshipsHandler(u.BelongsToFilter.Type, u.Rel.FromName)
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

// AddLoader sets the data loading function for a type
func (r *JSONAPIRouter) AddLoader(t string, loader JSONAPIDataLoader) {
	if _, ok := r.loaders[t]; ok {
		panic("loader exists for " + t)
	}
	// check if t is a legit type?
	r.loaders[t] = loader
}
func (r *JSONAPIRouter) loadIncludes(rReq *RouterReq) error {
	for t, loader := range r.loaders {
		ids := rReq.Includes.getLoadIDs(t)
		resources, err := loader(ids, rReq)
		if err != nil {
			return err
		}
		rReq.Includes.HoldResources(t, resources)
	}
	return nil
}

// NewCollection creates a resource collection for the given type
// This could be moved to any object that has the schema.
func (r *JSONAPIRouter) NewCollection(typ string) jsonapi.Collection {
	schemaType := r.schema.GetType(typ)
	if schemaType.Name == "" {
		panic("type not reocgnized: " + typ)
	}
	return NewCollection(schemaType)
}
