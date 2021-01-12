package jsonapirouter

import (
	"fmt"

	"github.com/mfcochauxlaberge/jsonapi"
)

// NewCollection creates a resource collection for the given type
func NewCollection(typ jsonapi.Type) jsonapi.Collection {
	return &Collection{
		typ: typ,
		col: make([]jsonapi.Resource, 0),
	}
}

// Collection implements the jsonapi.Collection interface
type Collection struct {
	typ jsonapi.Type
	col []jsonapi.Resource
}

// GetType returns the Type of the collection
func (c *Collection) GetType() jsonapi.Type {
	return c.typ
}

// Len retuns teh length of the collection
func (c *Collection) Len() int {
	return len(c.col)
}

// At returns the Resource at index i
func (c *Collection) At(i int) jsonapi.Resource {
	if len(c.col) > i {
		return c.col[i]
	}
	return nil
}

// Add a resource to the collection
func (c *Collection) Add(r jsonapi.Resource) {
	if r.GetType().Name != c.typ.Name {
		panic(fmt.Sprintf("Can not add resource of type %v to collection of type %v", r.GetType().Name, c.typ.Name))
	}
	c.col = append(c.col, r)
}
