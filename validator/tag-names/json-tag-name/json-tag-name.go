package jsonTagName

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func New() validator.TagNameFunc {
	return func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	}
}
