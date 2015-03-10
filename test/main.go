package main

import (
	"fmt"
	"net/http"

	"github.com/konek/rest"
)

// TestRet ...
type TestRet struct{
	Status	string
	Code	int
}

// StatusCode ...
func (t TestRet) StatusCode() int {
	return t.Code
}

func (t TestRet) Error() string {
	return t.Status
}

func testfunc(r *http.Request, p rest.Params) (interface{}, error) {
	return TestRet{
		"ok",
		200,
	}, nil
}

func test2func(r *http.Request, p rest.Params) (interface{}, error) {
	var data TestRet

	err := rest.Parse(r, &data)
	if err != nil {
		return nil, err
	}
	return nil, TestRet{
		"oops",
		403,
	}
}


func main() {
	router := rest.New()

	router.GET("/test", testfunc)
	router.POST("/test2", test2func)
	fmt.Println("listening on :8081")
	err := http.ListenAndServe(":8081", router)
	fmt.Println(err)
}
