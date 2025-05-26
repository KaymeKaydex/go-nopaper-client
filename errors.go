package go_nopaper_client

import "fmt"

type Error string

func (e Error) String() string {
	return string(e)
}

func (e Error) Error() string {
	return e.String()
}

var (
	ProfileByPhoneNotFound            Error = "profile by phone not found"
	RequestBodyWasNotConvertedToModel Error = "request body was not converted to model"
)

var errorMap = map[string]Error{
	"NOPAPERPARTNER.10401":         ProfileByPhoneNotFound,
	"NOPAPERPARTNERLIB.10401":      ProfileByPhoneNotFound,
	"NOPAPERPARTNERAPI.CORE.41116": RequestBodyWasNotConvertedToModel,
}

func errorByCode(code string) error {
	var err error

	err, exists := errorMap[code]
	if !exists {
		err = fmt.Errorf("unknown bad response error code from Nopaper: %s", code)
	}

	return err
}
