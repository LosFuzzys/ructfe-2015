package main

import (
	"io"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const STATIC_URL string = "/static/"
const STATIC_ROOT string = "static/"

type Link struct {
	LinkHref string
	LinkText string
}

type Context struct {
    LoggedIn bool
	Action string
	Text string
	Metrics []HealthMetrics
}

func addHealthMetricsHandler(w http.ResponseWriter, request *http.Request) {

	status, response := addHealthMetrics(request)
	w.WriteHeader(status)

	context := Context{LoggedIn: loggedin(request), Action: "", Text: response}
	render(w, "text", context)
}

func addHealthMetricsFormHandler(w http.ResponseWriter, request *http.Request) {
	context := Context{LoggedIn: loggedin(request), Action: "/addhealthmetrics", Text: ""}
	render(w, "metrics", context)
}

func healthMetricsHandler(w http.ResponseWriter, request *http.Request) {

	status, response, metrics := getHealthMetrics(request)
	w.WriteHeader(status)
	
	if status == http.StatusOK {
		context := Context{LoggedIn: loggedin(request), Metrics: metrics}
		render(w, "table", context)
	} else {
		context := Context{LoggedIn: loggedin(request), Text: response}
		render(w, "text", context)
	}
}

func addUserHandler(w http.ResponseWriter, request *http.Request) {

	status, response := addUser(request)
	w.WriteHeader(status)
	
	context := Context{LoggedIn: loggedin(request), Action: "", Text: response}
	render(w, "text", context)
}

func loginHandler(w http.ResponseWriter, request *http.Request) {
	
	status, response, c1, c2 := login(request)
	http.SetCookie(w, &c1) 
	http.SetCookie(w, &c2) 
	w.WriteHeader(status)
	
	loggedIn := status == http.StatusOK 

	context := Context{LoggedIn: loggedIn, Action: "", Text: response}
	render(w, "text", context)
}

func logoutHandler(w http.ResponseWriter, request *http.Request) {
	
	status, response, c1, c2 := logout(request)
	http.SetCookie(w, &c1) 
	http.SetCookie(w, &c2) 
	w.WriteHeader(status)

	context := Context{LoggedIn: loggedin(request), Action: "", Text: response}
	render(w, "text", context)
}

func signupFormHandler(w http.ResponseWriter, request *http.Request) {
	context := Context{LoggedIn: loggedin(request), Action: "/newuser"}
	render(w, "login", context)
}

func loginformHandler(w http.ResponseWriter, request *http.Request) {
	context := Context{LoggedIn: loggedin(request), Action: "/login"}
	render(w, "login", context)
}

func homeHandler(w http.ResponseWriter, request *http.Request) {
	context := Context{LoggedIn: loggedin(request)}
	render(w, "index", context)
}

func render(w http.ResponseWriter, tmpl string, context Context) {
    tmpl_list := []string{"base.html",
        fmt.Sprintf("%s.html", tmpl)}
    t, err := template.ParseFiles(tmpl_list...)
    if err != nil {
        fmt.Println("template parsing error: ", err)
    }
    err = t.Execute(w, context)
    if err != nil {
        fmt.Println("template executing error: ", err)
    }
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
    static_file := req.URL.Path[len(STATIC_URL):]
    if len(static_file) != 0 {
        f, err := http.Dir(STATIC_ROOT).Open(static_file)
        if err == nil {
            content := io.ReadSeeker(f)
            http.ServeContent(w, req, static_file, time.Now(), content)
            return
        }
    }
    http.NotFound(w, req)
}

var mux map[string]func(http.ResponseWriter, *http.Request)

type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := "/" + strings.Split(r.URL.String(), "/")[1]
	fmt.Println(path)
	if h, ok := mux[path]; ok {
		h(w, r)
		return
	}

	io.WriteString(w, "My server: "+path)
}

var links map[string]Link

func initService(){
	prepareDb()
}

func main() {

	initService()
	
	var port = flag.String("port", "8000", "please specify the port to start server on")
	flag.Parse()
	fmt.Println("Port to start on: " + *port)
	server := http.Server{
		Addr:    ":" + *port,
		Handler: &myHandler{},
	}

	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = homeHandler
	mux["/healthmetrics"] = healthMetricsHandler
	mux["/addhealthmetrics"] = addHealthMetricsHandler
	mux["/addhealthmetricsform"] = addHealthMetricsFormHandler
	mux["/newuser"] = addUserHandler
	mux["/signupform"] = signupFormHandler
	mux["/login"] = loginHandler
	mux["/loginform"] = loginformHandler
	mux["/logout"] = logoutHandler
	mux["/static"] = staticHandler
	//todo: about

	server.ListenAndServe()
}