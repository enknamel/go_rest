go-rest
=======

A very simple RESTful HTTP handler for Go

Simply create call NewRestRouter() to create a new RestRouter

Then call AddRoute with your HTTP method, url pattern and a handler.

AddRoute(method string, urlPattern string, handler func(http.ResponseWriter, *http.Request, *RestParams))

URL Parameters will be extracted if they are specified in your pattern like so "/user/:userId"