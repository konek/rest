
package rest

//Error is the interface that needs to be implemented in order to return meaningfull errors to the client.
type Error interface{
	StatusCode() int
}

// Error500 is an easy way to return 500 errors
type Error500 struct{
	Message string
}

// NewError500 creates an Error500 with the following message : "an unexpected error occured, please contact an administrator"
func NewError500() Error500 {
	return Error500{
		"an unexpected error occured, please contact an administrator",
	}
}

func (e Error500) Error() string {
	return e.Message
}

// StatusCode returns 500
func (e Error500) StatusCode() int {
	return 500
}

