package nopaper

import "fmt"

type Error string

func (e Error) String() string {
	return string(e)
}

func (e Error) Error() string {
	return e.String()
}

var (
	ErrProfileByPhoneNotFound            Error = "profile by phone not found"
	ErrRequestBodyWasNotConvertedToModel Error = "request body was not converted to model"
	// ErrNotFullUserProfile - errors is caused by not full user profile.
	ErrNotFullUserProfile Error = "cannot be created certificate without full name profile fl"
)

var errorMap = map[string]Error{
	"NOPAPERPARTNER.10401":         ErrProfileByPhoneNotFound,
	"NOPAPERPARTNERLIB.10401":      ErrProfileByPhoneNotFound,
	"NOPAPERPARTNERAPI.CORE.41116": ErrRequestBodyWasNotConvertedToModel,
	"NOPAPERPARTNER.10300":         ErrNotFullUserProfile,
}

func errorByCode(code string) error {
	var err error

	err, exists := errorMap[code]
	if !exists {
		err = fmt.Errorf("unknown bad response error code from Nopaper: %s", code)
	}

	return err
}
