package api

import "reflect"

func getType(class interface{}) string {
	if t := reflect.TypeOf(class); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
