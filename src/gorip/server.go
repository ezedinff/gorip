// author			sigu-399
// author-github	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gorip
// repository-desc	REST Server Framework - ( gorip: REST In Peace ) - Go language
// 
// description		Server implementation.
// 
// created			03-03-2013

package gorip

import (
	"log"
	"net/http"
	"time"
)

type Server struct {
	pattern string
	address string
	router  *router
}

func NewServer(pattern string, address string) *Server {
	return &Server{pattern: pattern, address: address, router: NewRouter()}
}

func (s *Server) ListenAndServe() error {

	http.Handle(s.pattern, s)
	return http.ListenAndServe(s.address, nil)
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	timeStart := time.Now()

	urlPath := request.URL.Path

	node, err := s.router.FindNodeByRoute(urlPath)
	if err != nil {
		log.Printf("Warning : %s", err.Error())
	}

	if node == nil {
		log.Printf("Warning : Could not find route for %s", urlPath)
	} else {
		log.Printf("%s", node)
	}

	timeEnd := time.Now()
	durationMs := timeEnd.Sub(timeStart).Seconds() * 1000

	log.Printf("Response : %2.2f ms", durationMs)

}

func (s *Server) RegisterEndpoint(e endpoint) error {
	log.Printf("Registering endpoint : %s\n", e.GetRoute())
	return s.router.RegisterRoute(e.GetRoute())
}
