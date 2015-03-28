package rest

import (
	"log"
	"errors"
	"reflect"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/xml"

	"github.com/julienschmidt/httprouter"
)

const (
	formatJSON = iota
	formatXML
	formatFORM
)

// Router overload the httprouter.Router in order to add default behaviors
type Router struct {
	*httprouter.Router
}

// Params contain an httprouter.Param, in order to avoid useless import of httprouter
type Params struct {
	httprouter.Params
}

// Resp is an interface allowing to return custom statusCode (200 will be used otherwise)
type Resp interface{
	StatusCode() int
}

// Controller is the function signature to be used with the GET/POST/... functions.
// A response of type Resp can be returned in order to overwrite the default 200 response.
// An error of type Error can be returned in order to overwrite the default error message.
type Controller func(r *http.Request, p Params) (interface{}, error)

func parseForm(form map[string][]string, v interface{}) error {
	val := reflect.ValueOf(v)
	t := val.Type()
	if t.Kind() != reflect.Ptr || val.IsNil(){
		return errors.New("Cannot parse form to non-pointer types")
	}
	val = val.Elem()
	for k, v := range(form) {
		if len(v) == 0 {
			continue
		}
		field := val.FieldByNameFunc(func (s string) bool {
			key := strings.ToLower(k)
			str := strings.ToLower(s)
			return key == str
		})
		if field.Kind() == reflect.String {
			field.SetString(v[0])
		}
	}
	return nil
}


// Parse is an helper function to parse the body according to its content-type. It supports json, xml and www-form-urlencoded
func Parse(r *http.Request, v interface{}) error {
	var err error

	outputFormat, _ := getFormat(r, "Accept")
	inputFormat, found := getFormat(r, "Content-Type")
	if found == false {
		if header, ok := r.Header["Content-Type"]; ok == true && len(header) != 0 {
		// 	return Error500{"unsupported Content-Type: " + header[0]}
		}
		inputFormat = outputFormat
	}

	if inputFormat == formatJSON {
		chunk, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return Error500{"failed to read body"}
		}

		err = json.Unmarshal(chunk, v)
	} else if inputFormat == formatXML {
		chunk, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return Error500{"failed to read body"}
		}

		err = xml.Unmarshal(chunk, v)
	} else if inputFormat == formatFORM {
		err = r.ParseForm()
		if err == nil {
			err = parseForm(r.PostForm, v)
		}
	} else {
		return errors.New("unknown output format")
	}
	if err != nil {
		return Error500{"failed to parse body: " + err.Error()}
	}
	return nil
}

func getFormat(r *http.Request, field string) (format int, found bool) {
	if header, ok := r.Header[field]; ok == true {
		for _, format := range(header) {
			if format == "application/json" {
				return formatJSON, true
			} else if format == "application/xml" {
				return formatXML, true
			} else if format == "application/x-www-form-urlencoded" {
				return formatFORM, true
			}
		}
	}
	return formatJSON, false
}

func output(w http.ResponseWriter, code int, data interface{}, format int) error {
	var chunk []byte
	var err error

	if format == formatJSON {
		chunk, err = json.Marshal(data)
		w.Header().Set("Content-Type", "aplication/json")
	} else if format == formatXML {
		chunk, err = xml.Marshal(data)
		w.Header().Set("Content-Type", "aplication/xml")
	} else {
		return errors.New("unknown output format")
	}
	if err != nil {
		return err
	}
	w.WriteHeader(code)
	_, err = w.Write(chunk)

	return err
}

func handler(fn Controller) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		outputFormat, _ := getFormat(r, "Accept")
		resp, err := fn(r, Params{p})
		if err != nil {
			if err2, ok := err.(Error); ok == true {
				err3 := output(w, err2.StatusCode(), err2, outputFormat)
				if err3 != nil {
					log.Println("error while writing error:", err3)
				}
				return
			}
			err2 := output(w, 500, NewError500(), outputFormat)
			if err2 != nil {
				log.Println("error while writing error:", err2)
			}
			return
		}
		statusCode := 200
		if resp2, ok := resp.(Resp); ok == true {
			statusCode = resp2.StatusCode()
		}
		err = output(w, statusCode, resp, outputFormat)
		if err != nil {
			log.Println("error while writing data:", err)
		}
	}
}

// GET is an overload to httprouter. Please refer to httprouter.GET for more details about the path
func (r *Router) GET(path string, ctrl Controller) {
	r.Router.GET(path, handler(ctrl))
}

// RawGET is an overload to httprouter. Please refer to httprouter.GET for more details about the path
func (r *Router) RawGET(path string, ctrl httprouter.Handle) {
	r.Router.GET(path, ctrl)
}

// POST is an overload to httprouter. Please refer to httprouter.POST for more details about the path
func (r *Router) POST(path string, ctrl Controller) {
	r.Router.POST(path, handler(ctrl))
}

// RawPOST is an overload to httprouter. Please refer to httprouter.POST for more details about the path
func (r *Router) RawPOST(path string, ctrl httprouter.Handle) {
	r.Router.POST(path, ctrl)
}

// PUT is an overload to httprouter. Please refer to httprouter.PUT for more details about the path
func (r *Router) PUT(path string, ctrl Controller) {
	r.Router.PUT(path, handler(ctrl))
}

// RawPUT is an overload to httprouter. Please refer to httprouter.PUT for more details about the path
func (r *Router) RawPUT(path string, ctrl httprouter.Handle) {
	r.Router.PUT(path, ctrl)
}

// DELETE is an overload to httprouter. Please refer to httprouter.DELETE for more details about the path
func (r *Router) DELETE(path string, ctrl Controller) {
	r.Router.DELETE(path, handler(ctrl))
}

// RawDELETE is an overload to httprouter. Please refer to httprouter.DELETE for more details about the path
func (r *Router) RawDELETE(path string, ctrl httprouter.Handle) {
	r.Router.DELETE(path, ctrl)
}

// New creates a new router.
func New() *Router {
	r := new(Router)
	r.Router = httprouter.New()
	return r
}

