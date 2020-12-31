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
			apiReq := jsonapi.Request{
				Method: c.method,
				URL:    apiURL}
			ht := router.getHandleType(&apiReq)
			if ht != c.handlerType {
				t.Errorf("Wrong handler type. Expected %v, got %v", c.handlerType, ht)
			}
		})
	}

}

func TestHandlers(t *testing.T) {
	schema := getTestSchema()
	router := NewJSONAPIRouter(schema)
	hitArticlesCollection := false
	router.GetCollection("articles", func(res http.ResponseWriter, httpReq *http.Request, apiReq *jsonapi.Request) {
		hitArticlesCollection = true
	})

	apiURL, err := jsonapi.NewURLFromRaw(schema, "/articles")
	if err != nil {
		t.Error(err)
	}
	router.Handle(nil, nil, &jsonapi.Request{
		Method: http.MethodGet,
		URL:    apiURL})

	if !hitArticlesCollection {
		t.Error("expected to hit articles collection")
	}
}

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
