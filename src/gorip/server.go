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
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	method := request.Method

	resourceContext := ResourceContext{}

	log.Printf("Requesting %s %s", method, urlPath)

	// Find route node and associated route variables
	node, routeVariables, err := s.router.FindNodeByRoute(urlPath)
	if err != nil {
		log.Printf("Warning : %s", err.Error())
	}

	if node == nil {
		log.Printf("Warning : Could not find route for %s", urlPath)
	} else {

		// Route was found:

		// Add route variables to the context
		resourceContext.RouteVariables = routeVariables

		if node.GetEndpoint() == nil {
			log.Printf("Warning : No endpoint found for this route")
		} else {

			// Endpoint was found:

			// Parse Content-Type and Accept headers

			contentTypeParser, err := newContentTypeHeaderParser(request.Header.Get(`Content-Type`))
			if err != nil {
				log.Printf(`Invalid Content-Type header : ` + err.Error())
			}

			acceptParser, err := newAcceptHeaderParser(request.Header.Get(`Accept`))
			if err != nil {
				log.Printf(`Invalid Accept header : ` + err.Error())
			}

			if !acceptParser.HasAcceptElement() {
				log.Printf(`No valid Accept header was given`)
			} else {

				// Headers are OK:

				endp := node.GetEndpoint()
				availableResourceImplementations := endp.GetResources()

				if len(availableResourceImplementations) == 0 {
					log.Printf(`No resource found on this route`)
				} else {

					matchingResource, contentTypeIn, contentTypeOut := endp.FindMatchingResource(method, &contentTypeParser, &acceptParser)

					if matchingResource == nil {
						log.Printf(`No available resource for this Content-Type`)
					} else {

						// Found a matching resource implementation: 

						// Add expected content type to the context 
						resourceContext.ContentTypeIn = contentTypeIn
						resourceContext.ContentTypeOut = contentTypeOut

						// Read request body

						bodyInBytes, err := ioutil.ReadAll(request.Body)
						if err != nil {
							log.Printf(`Could not read request body`)
						} else {
							if resourceContext.ContentTypeIn == nil && len(bodyInBytes) > 0 {
								log.Printf(`Body is not allowed for this resource`)
							} else {
								resourceContext.Body = bytes.NewBuffer(bodyInBytes)
							}
						}

						// Create a new instance from factory and executes it
						resource := matchingResource.Factory()
						if resource == nil {
							log.Printf(`Resource factory must instanciate a valid Resource`)
						} else {

							result := resource.Execute(&resourceContext)

							// Http response
							bodyOutLen := 0
							if result.Body != nil {
								bodyOutLen = result.Body.Len()
							}

							writer.Header().Set(`Content-Length`, strconv.Itoa(bodyOutLen))

							if bodyOutLen > 0 {
								writer.Header().Add(`Content-Type`, *resourceContext.ContentTypeOut)
							}

							writer.WriteHeader(result.HttpStatus)

							if bodyOutLen > 0 {
								_, err := result.Body.WriteTo(writer)
								if err != nil {
									log.Printf(`Error while writing the body %s`, err.Error())
								}
							}

						}
					}
				}
			}
		}
	}

	timeEnd := time.Now()
	durationMs := timeEnd.Sub(timeStart).Seconds() * 1000

	log.Printf("Response time : %2.2f ms", durationMs)

}

func (s *Server) RegisterEndpoint(e *endpoint) error {

	if e == nil {
		panic(`Endpoint cannot be nil`)
	}

	log.Printf("Registering endpoint : %s\n", e.GetRoute())
	return s.router.RegisterEndpoint(e)
}

func (s *Server) RegisterRouteVariableValidator(kind string, validator RouteVariableValidator) error {
	return s.router.RegisterRouteVariableValidator(kind, validator)
}
