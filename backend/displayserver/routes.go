package displayserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		//handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	//http.FileServer(http.Dir("../api/index.html"))
	fmt.Fprint(w, "abc")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},

	Route{
		"startCharging",
		strings.ToUpper("Post"),
		"/start/{EVSEID}",
		onStart,
	},

	Route{
		"stopCharging",
		strings.ToUpper("Post"),
		"/stop/{EVSEID}",
		onStop,
	},

	Route{
		"stationStatus",
		strings.ToUpper("Get"),
		"/chargestatus/{EVSEID}",
		onGetChargeStatus,
	},

	Route{
		"ids",
		strings.ToUpper("Get"),
		"/evses/active/ids",
		onGetEVSEsActiveIds,
	},
}
