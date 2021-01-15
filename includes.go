package jsonapirouter

import (
	"sort"

	"github.com/mfcochauxlaberge/jsonapi"
)

// includes management
// We need to know if something already *has* been included?
// We also would like to *not* load resources that are already included

type incStruct struct {
	id       string
	required bool
	included bool
	resource jsonapi.Resource
}

// Includes can track missing includes
// and hold on to resources for later inclusion
type Includes struct {
	doc *jsonapi.Document
	// typeIncs of all resources that should be included
	typeIncs map[string][]incStruct
}

// NewIncludes returns an Includes
func NewIncludes(doc *jsonapi.Document) *Includes {
	return &Includes{
		doc:      doc,
		typeIncs: map[string][]incStruct{},
	}
}

// HoldResource to the cache of resources that can be included in a doc
func (incs *Includes) HoldResource(r jsonapi.Resource) {
	typ := r.GetType().Name
	i := incs.getIndex(typ, r.Get("id").(string))
	incs.typeIncs[typ][i].resource = r
}

// HoldResources adds multiple resources of the given type
// to the cache of resources
func (incs *Includes) HoldResources(typ string, rs []jsonapi.Resource) {
	_, ok := incs.typeIncs[typ]
	if !ok {
		incs.typeIncs[typ] = []incStruct{}
	}
	for _, r := range rs {
		id := r.Get("id").(string)
		i := incs.getIndex(typ, id)
		incs.typeIncs[typ][i].resource = r
	}
}

// AddAllToDoc adds the required resources to the doc's includes
// Returns true if all required includes are added to the doc
func (incs *Includes) AddAllToDoc() bool {
	ret := true
	for typ, incDatas := range incs.typeIncs {
		for i, incData := range incDatas {
			if incData.required && !incData.included {
				if incData.resource != nil {
					incs.doc.Include(incData.resource)
					incs.typeIncs[typ][i].included = true
				} else {
					ret = false
				}
			}
		}
	}
	return ret
}

// extractIDs looks at what is in data and what is meant to be included
// and adds the ids of to-be included resources in the local map
// This code is modeled after jsonapi's MarshalDocument
func (incs *Includes) extractIDs(rReq *RouterReq) {
	switch d := rReq.Doc.Data.(type) {
	case jsonapi.Resource:
		dataTypeName := d.GetType().Name
		incs.appendResourceIncludes(d, rReq.URL.Params.Fields[d.GetType().Name], rReq.Doc.RelData[dataTypeName])
	case jsonapi.Collection:
		fields := rReq.URL.Params.Fields[d.GetType().Name]
		dataTypeName := d.GetType().Name
		relData := rReq.Doc.RelData[dataTypeName]
		for i := 0; i < d.Len(); i++ {
			incs.appendResourceIncludes(d.At(i), fields, relData)
		}
	}
}

func (incs *Includes) appendResourceIncludes(r jsonapi.Resource, fields []string, relData []string) {
	for _, rel := range r.Rels() {
		include := false

		for _, field := range fields {
			if field == rel.FromName {
				include = true
				break
			}
		}

		if include {
			if rel.ToOne {
				for _, n := range relData {
					if n == rel.FromName {
						id := r.Get(rel.FromName)
						incs.appendInclude(rel.ToType, id.(string))
						break
					}
				}
			} else {
				for _, n := range relData {
					if n == rel.FromName {
						ids := r.Get(rel.FromName).([]string)
						sort.Strings(ids)
						for _, id := range ids {
							incs.appendInclude(rel.ToType, id)
						}
						break
					}
				}
			}
		}
	}
}

func (incs *Includes) appendInclude(typ string, id string) {
	i := incs.getIndex(typ, id)
	incs.typeIncs[typ][i].required = true
}

func (incs *Includes) getLoadIDs(t string) []string {
	ids, ok := incs.typeIncs[t]
	if !ok {
		return []string{}
	}
	ret := make([]string, 0)
	for _, inc := range ids {
		if inc.required && inc.resource == nil {
			ret = append(ret, inc.id)
		}
	}
	return ret
}

func (incs *Includes) getIndex(typ string, id string) int {
	_, ok := incs.typeIncs[typ]
	if !ok {
		incs.typeIncs[typ] = []incStruct{{id: id}}
		return 0
	}
	for i, incID := range incs.typeIncs[typ] {
		if incID.id == id {
			return i
		}
	}
	incs.typeIncs[typ] = append(incs.typeIncs[typ], incStruct{id: id})
	return len(incs.typeIncs[typ]) - 1
}
