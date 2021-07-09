package main

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type ToDo struct {
	Id        uuid.UUID `json:"id" description:"uuid of the todo"`
	Title     string    `json:"title" description:"title of the todo"`
	Completed bool      `json:"completed" description:"status of the todo" default:"false"`
	Created   int64     `json:"created" description:"creation time of the todo"`
}

type ToDoResource struct {
	ToDos sync.Map
}

func (p ToDoResource) getToDo(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("id")

	todo, found := p.ToDos.Load(id)
	if !found {
		resp.WriteErrorString(http.StatusBadRequest, "todo with id "+id+" is not found ")
		return
	}
	log.Println("returned product with id:" + id)
	resp.WriteEntity(todo)
}

func (p ToDoResource) addToDo(req *restful.Request, resp *restful.Response) {
	newToDo := new(ToDo)
	err := req.ReadEntity(newToDo)
	if err != nil { // bad request
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}
	newToDo.Id = uuid.New()
	newToDo.Created = time.Now().UnixNano()
	p.ToDos.Store(newToDo.Id.String(), newToDo)
	log.Println("added product with id:" + newToDo.Id.String())
	resp.WriteEntity(newToDo)
}

func (p ToDoResource) updateTodo(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("id")
	loadedTodo, found := p.ToDos.Load(id)
	if !found {
		resp.WriteErrorString(http.StatusBadRequest, "todo with id "+id+" is not found ")
		return
	}

	updatedToDo := new(ToDo)
	err := req.ReadEntity(updatedToDo)
	if err != nil { // bad request
		resp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	todo := loadedTodo.(ToDo)
	todo.Title = updatedToDo.Title
	todo.Completed = updatedToDo.Completed
	p.ToDos.Store(id, todo)

	log.Println("updated product with id:" + id)
	resp.WriteEntity(todo)
}

func (p ToDoResource) deleteTodo(req *restful.Request, resp *restful.Response) {
	id := req.PathParameter("id")
	if _, found := p.ToDos.Load(id); !found {
		resp.WriteErrorString(http.StatusBadRequest, "todo with id "+id+" is not found ")
		return
	}

	p.ToDos.Delete(id)

	log.Println("deleted product with id:" + id)
}

func (p ToDoResource) getAllTodo(req *restful.Request, resp *restful.Response) {
	arr := []ToDo{}
	p.ToDos.Range(func(key, value interface{}) bool {
		v := value.(ToDo)
		arr = append(arr, v)
		return true
	})

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].Created < arr[j].Created
	})

	resp.WriteEntity(arr)
	log.Println("return all todos")
}

func (p ToDoResource) RegisterTo(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/products")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	p.ToDos = sync.Map{}

	ws.Route(ws.GET("/{id}").To(p.getToDo).
		Doc("get the product by its id").
		Param(ws.PathParameter("id", "identifier of the product").DataType("integer")))

	ws.Route(ws.POST("/").To(p.addToDo).
		Doc("update or create a product").
		Param(ws.BodyParameter("ToDo", "a ToDo (JSON)").DataType("main.ToDo")))

	ws.Route(ws.PUT("/{id}").To(p.updateTodo).
		Doc("get the product by its id").
		Param(ws.PathParameter("id", "identifier of the product").DataType("integer")).
		Param(ws.BodyParameter("ToDo", "a ToDo (JSON)").DataType("main.ToDo")))

	ws.Route(ws.DELETE("/{id}").To(p.deleteTodo).
		Doc("update or create a product").
		Param(ws.PathParameter("id", "identifier of the product").DataType("integer")))

	ws.Route(ws.GET("").To(p.getAllTodo).
		Doc("get all todos"))

	container.Add(ws)
}

func CORSFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader(restful.HEADER_AccessControlAllowOrigin, "*")
	chain.ProcessFilter(req, resp)
}

func main() {
	wsContainer := restful.NewContainer()
	t := ToDoResource{}
	t.RegisterTo(wsContainer)

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"PUT", "POST", "GET", "DELETE"},
		AllowedDomains: []string{"*"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	// Add container filter to respond to OPTIONS
	wsContainer.Filter(wsContainer.OPTIONSFilter)
	wsContainer.Filter(CORSFilter)

	log.Print("start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", wsContainer))
}
