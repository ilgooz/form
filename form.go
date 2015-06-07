package form

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/ilgooz/ersp"
)

//todo: slices

type Form struct {
	schema    interface{}
	w         http.ResponseWriter
	r         *http.Request
	Error     *ersp.Response
	existence map[string]bool
}

func Parse(schema interface{}, w http.ResponseWriter, r *http.Request) (*Form, error) {
	form := &Form{
		schema:    schema,
		w:         w,
		r:         r,
		Error:     ersp.New(w),
		existence: make(map[string]bool, 0),
	}
	return form, form.parse()
}

func (form *Form) parse() error {
	if err := form.r.ParseForm(); err != nil {
		form.Error.SendParseFormError()
		return nil
	}

	if err := form.parseValues(); err != nil {
		return err
	}

	return nil
}

func (form *Form) ApplyTo(out interface{}) {

}

type Rule struct {
	As       string
	Required bool
	Min      int
	Email    bool
	Comma    bool
}

func (form *Form) parseValues() error {
	t := reflect.Indirect(reflect.ValueOf(form.schema)).Type()
	for i := 0; i < t.NumField(); i++ {
		rule, err := form.rule(t.Field(i).Tag.Get("form"))
		if err != nil {
			return err
		}
		form.convert(rule, reflect.ValueOf(form.schema).Elem().Field(i))
	}
	return nil
}

func (form *Form) rule(s string) (Rule, error) {
	r := Rule{}
	a := strings.Split(s, ",")
	for _, c := range a {
		b := strings.Split(c, ":")
		switch b[0] {
		case "as":
			r.As = b[1]
		case "required":
			r.Required = true
		case "min":
			i, err := strconv.Atoi(b[1])
			if err != nil {
				return r, TagError{err}
			}
			r.Min = i
		case "email":
			r.Email = true
		case "comma":
			r.Comma = true
		default:
			return r, TagError{errors.New("Unknown rule: " + b[0])}
		}
	}
	return r, nil
}

func (form *Form) convert(rule Rule, field reflect.Value) {
	value, exists := form.r.Form[rule.As]
	form.existence[rule.As] = exists
	if !exists {
		if rule.Required {
			form.Error.Field(rule.As, "required")
		}
		return
	}

	l := len(value)

	switch field.Type().String() {
	case "string":
		if rule.Min > 0 && len(value[0]) < rule.Min {
			form.Error.Field(rule.As, fmt.Sprintf("must be at least %d chars long", rule.Min))
			return
		}
		if rule.Email && !govalidator.IsEmail(value[0]) {
			form.Error.Field(rule.As, "not valid")
			return
		}
		field.SetString(value[0])
	case "int64":
		i, err := strconv.ParseInt(value[0], 0, 64)
		if err != nil {
			form.Error.Field(rule.As, "must be a number")
		}
		field.SetInt(i)
	case "float32":
		i, err := strconv.ParseFloat(value[0], 32)
		if err != nil {
			form.Error.Field(rule.As, "must be a number")
		}
		field.SetFloat(i)
	case "bool":
		switch value[0] {
		case "true":
			field.SetBool(true)
		case "false":
			field.SetBool(false)
		default:
			form.Error.Field(rule.As, "must be true or false")
		}
	case "time.Time":
		t := time.Time{}
		err := t.UnmarshalText([]byte(value[0]))
		if err != nil {
			form.Error.Field(rule.As, "must be UTC")
		}
		field.Set(reflect.ValueOf(t))
	case "*time.Time":
		t := &time.Time{}
		err := t.UnmarshalText([]byte(value[0]))
		if err != nil {
			form.Error.Field(rule.As, "must be UTC")
		}
		field.Set(reflect.ValueOf(t))
	case "[]int64":
		if rule.Comma {
			values := strings.Split(value[0], ",")
			l = len(values)
			s := reflect.MakeSlice(reflect.TypeOf([]int64{}), l, l)
			for i, v := range values {
				val := strings.Trim(v, " ")
				in, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					form.Error.Field(rule.As, "must be comma separated numbers")
					return
				}
				s.Index(i).Set(reflect.ValueOf(in))
			}
			field.Set(s)
		} else {
			s := reflect.MakeSlice(reflect.TypeOf([]int64{}), l, l)
			for i, v := range value {
				in, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					form.Error.Field(rule.As, "must be numbers")
					return
				}
				s.Index(i).Set(reflect.ValueOf(in))
			}
			field.Set(s)
		}
	case "[]string":
		s := reflect.MakeSlice(reflect.TypeOf([]string{}), l, l)
		for i, v := range value {
			s.Index(i).Set(reflect.ValueOf(v))
		}
		field.Set(s)
	}
}

func (form *Form) HasError() bool {
	return form.Error.HasError()
}

func (form *Form) Exists(s string) bool {
	return form.existence[s]
}

func Time(ts string) (time.Time, error) {
	t := time.Time{}
	err := t.UnmarshalText([]byte(ts))
	return t, err
}
