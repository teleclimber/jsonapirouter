package jsonapirouter

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/mfcochauxlaberge/jsonapi"
)

func TestSchema(t *testing.T) {
	getTestSchema()
}

func TestGetHandleType(t *testing.T) {
	router := JSONAPIRouter{
		schema: getTestSchema()}

	cases := []struct {
		method      string
		path        string
		handlerType handlerType
	}{
		{http.MethodGet, "/articles", getCollection},
		{http.MethodGet, "/articles/1", getResource},
		{http.MethodGet, "/articles/1/author", getRelated},
		{http.MethodGet, "/articles/1/tags", getRelated},
		/////// invalid request {http.MethodGet, "/articles/1/relationships", getRelationships},
		{http.MethodGet, "/articles/1/relationships/tags", getRelationships},
		{http.MethodGet, "/tags/1/relationships/articles", getRelationships},
		{http.MethodPatch, "/articles/1", updateResource},
		// TODO: other methods
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			goURL, err := url.Parse(c.path)
			if err != nil {
				t.Error(err)
			}
			sURL, err := jsonapi.NewSimpleURL(goURL)
			if err != nil {
				t.Error(err)
			}
			apiURL, err := jsonapi.NewURL(router.schema, sURL)
			if err != nil {
				t.Error(err)
			}
			ht := router.getHandleType(c.method, apiURL)
			if ht != c.handlerType {
				t.Errorf("Wrong handler type. Expected %v, got %v", c.handlerType, ht)
			}
		})
	}

}

func TestGetCollectionHandler(t *testing.T) {
	router := NewJSONAPIRouter(nil)
	touched := false
	hIn := func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status {
		touched = true
		return OK
	}
	router.GetCollection("abc", hIn)
	hOut, ok := router.getCollectionHandler("abc")
	if !ok {
		t.Error("expected a handler")
	}
	hOut(nil, nil, nil)
	if !touched {
		t.Error("got the wrong handler?")
	}

	hOut, ok = router.getCollectionHandler("def")
	if ok {
		t.Error("def should not return a handler")
	}
	if hOut != nil {
		t.Error("hOut should be nil")
	}
}

func TestGetRelatedHandler(t *testing.T) {
	router := NewJSONAPIRouter(nil)
	touched := false
	hGood := func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status {
		touched = true
		return OK
	}
	hBad := func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status {
		return OK
	}

	router.GetRelated("abc", "xyz", hGood)
	router.GetRelated("abc", "other", hBad)
	router.GetRelated("bad", "other", hBad)

	hOut, ok := router.getRelatedHandler("abc", "xyz")
	if !ok {
		t.Error("Expcted tog et a handler")
	}
	hOut(nil, nil, nil)
	if !touched {
		t.Error("expected the right handler")
	}
	hOut, ok = router.getRelatedHandler("abc", "def")
	if ok {
		t.Error("should not have found handler")
	}
	hOut, ok = router.getRelatedHandler("zzz", "def")
	if ok {
		t.Error("should not have found handler")
	}
}

// func TestHandlers(t *testing.T) {
// 	schema := getTestSchema()
// 	router := NewJSONAPIRouter(schema)
// 	hitArticlesCollection := false
// 	router.GetCollection("articles", func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status {
// 		hitArticlesCollection = true
// 		return OK
// 	})

// 	apiURL, err := jsonapi.NewURLFromRaw(schema, "/articles")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	router.Handle(nil, nil, &RouterReq{
// 		URL: apiURL})

// 	if !hitArticlesCollection {
// 		t.Error("expected to hit articles collection")
// 	}
// }

// getTestSchema creates a sample schema that tests can use.
// - articles
//   "author" -> user
//   "tags" -> *tags
// - tags
//   "articles" -> *articles
// - users
//   "articles" -> *articles
//
func getTestSchema() *jsonapi.Schema {
	schema := &jsonapi.Schema{}

	articles := jsonapi.Type{
		Name: "articles",
	}
	articles.AddAttr(jsonapi.Attr{
		Name:     "title",
		Type:     jsonapi.AttrTypeString,
		Nullable: false,
	})
	schema.AddType(articles)

	tags := jsonapi.Type{
		Name: "tags",
	}
	tags.AddAttr(jsonapi.Attr{
		Name:     "name",
		Type:     jsonapi.AttrTypeString,
		Nullable: false,
	})
	schema.AddType(tags)

	users := jsonapi.Type{
		Name: "users",
	}
	users.AddAttr(jsonapi.Attr{
		Name:     "username",
		Type:     jsonapi.AttrTypeString,
		Nullable: false,
	})
	schema.AddType(users)

	schema.AddTwoWayRel(jsonapi.Rel{
		FromType: "articles",
		FromName: "author",
		ToOne:    true,
		ToType:   "users",
		ToName:   "articles",
		FromOne:  false,
	})

	schema.AddTwoWayRel(jsonapi.Rel{
		FromType: "articles",
		FromName: "tags",
		ToOne:    false,
		ToType:   "tags",
		ToName:   "articles",
		FromOne:  false,
	})

	errs := schema.Check()
	if len(errs) != 0 {
		for _, e := range errs {
			fmt.Printf("%v \n", e)
		}
		panic(errs)
	}

	return schema
}
