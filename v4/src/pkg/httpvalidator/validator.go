package httpvalidator

import (
	"net/http"
	"strconv"
	"strings"

	valid "github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v4"
)

func init() { //nolint: gochecknoinits
	valid.SetFieldsRequiredByDefault(true)

	valid.TagMap["positive_uint"] = func(str string) bool {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return false
		}

		if v == 0 {
			return false
		}

		return true
	}
}

type CustomValidator struct{}

func (cv *CustomValidator) Validate(i any) error {
	result, err := valid.ValidateStruct(i)
	if !result || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func ParseErrors(err error) []ValidationError {
	tmp := any(err)

	internal, ok := tmp.(*echo.HTTPError)
	if !ok {
		return []ValidationError{{"internal", err.Error()}}
	}

	var errs []ValidationError

	for _, str := range strings.Split(internal.Message.(string), ";") { //nolint: forcetypeassert
		data := strings.SplitN(str, ":", 2) //nolint: gomnd

		errs = append(errs, ValidationError{strings.TrimSpace(data[0]), strings.TrimSpace(data[1])})
	}

	return errs
}
