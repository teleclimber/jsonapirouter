# jsonapirouter

A router and more for [mfcochauxlaberge/jsonapi](https://github.com/mfcochauxlaberge/jsonapi).

`jsonapirouter` builds on jsonapi with the goal of making it easier to create correct [JSON:API](https://jsonapi.org/format/) responses.

This code assumes that the resources that make up the responses are available in a separate part of your code. In other words, the database layer is completely decoupled from the API handlers.

## Experimental Code 

This code is experimental and rough and lacking in tests and robustness. Don't expect this to change very soon.

## Router

The router helps you divide your API code into different handlers. This code assumes there are eight types of handlers, as determined by looking at the JSON:API spec. These are:

- getCollection
- getResource
- getRelated
- getRelationships
- createResource
- updateResource
- updateRelationships
- deleteResource

For each of the types of your API, you can associate a handler as follows:

```
router.GetCollection("articles", getArticles)
```

The idea is that you set your handlers as you build your schema, and this keeps related things close together:

```
schema.AddType(jsonapi.MustBuildType(Article{}))
router.GetCollection("articles", getArticles)
router.GetResource("articles", getArticle)
// etc...
```

Handlers are `jsonapirouter.JSONAPIRouteHandler`:

```
type JSONAPIRouteHandler func(res http.ResponseWriter, httpReq *http.Request, rReq *RouterReq) Status
```

Where `RouterReq` is a struct that has the `URL`, `Document`, and `Includes` (see below).

## Includes

Now that requests are dispatched to a handler, the next problem is minimizing code duplication for handling included resources.JSON:API handlers are prone to duplicate code, because all handlers for types that are related to another type may have to load that related type.

`jsonapirouter.Includes` can

- determine which resource ids need to be included based on the requests parameters and the value of `jsonapi.Document.Data`
- hold on to Resources that have been loaded from the DB so they can be included automatically if required
- load missing resources via loader functions

Depending on your DB and models code, fetching a primary resource may include the id of a related resource that ought to be included. This is common with to-one relationships.

In other cases, to get related ids, you will have to perform an additional DB request, and this request may include complete data for related resources, not just ids.

In either case, `Includes` can help.

### Data Loaders

In addition to setting route handlers as above, we can also set data loaders when we build the schema:

```
schema.AddType(jsonapi.MustBuildType(Article{}))
router.AddLoader("articles", loadArticles)	// <--- data loader
router.GetCollection("articles", getArticles)
router.GetResource("articles", getArticle)
// etc...
```

Data loaders accept a list of IDs and return the actual resources.

### Got An ID

If in the process of loading the data for `Document.Data` you get the id(s) of the related resources to include, and you have set up a data loader as above, then you are done.

```
rReq.Doc.Data = myResource

return // you're done.
```

After the route handler runs the router will process the `Data` and concoct a list of ids it needs, feed it to the loader and include the resources in the `Document`.

### Got the Related Resources

If, as is often the case in to-many relationships, your route handler had to load all the resources of that relationship just to get the related ids, then you can "stash" them in `Includes`:

```
rReq.Includes.HoldResource(r)
```

The ids of Resources that are "held" will not be sent to the loader. This prevents unnecessary DB calls.

Basically the point of `jsonapirouter.Includes` is to collect the *ids* that need to be included, then load only the missing ones, as opposed to loading a lot of data from the DB and ignoring duplicates.


## Collection

Collection is a small helper that makes it easier to create collections in your handlers.

A `jsonapirouter.Collection` implements the `jsonapi.Collection` and can be initialized without passing an instance of an item of the collection.

A collection can be initialized by passing a `jsonapi.Type`:

```
myCol := NewCollection(articlesTyp)
```

If you have a router instance, it's even easier to create a collection, by passing the type's name as a string:

```
myCol := router.NewCollection("articles")
```

## License

MIT