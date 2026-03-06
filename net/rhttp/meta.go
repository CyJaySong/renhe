package rhttp

import (
	"reflect"
)

type HttpApiMeta struct{}

// metaType holds the reflection. Type of HttpApiMeta, used for efficient type comparison.
var metaType = reflect.TypeOf(HttpApiMeta{})

const metaNameTagName = "name"
const metaPathTagName = "path"
const metaMethodTagName = "method"

type tagData struct {
	Name   string
	Path   string
	Method string
}

// HttpApiMetaData retrieves and returns all metadata from `object`.
func HttpApiMetaData(object any) *tagData {
	reflectType, ok := object.(reflect.Type)
	if !ok {
		reflectType = reflect.TypeOf(object)
	}
	if reflectType.Kind() == reflect.Pointer {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return nil
	}
	for field := range reflectType.Fields() {
		if field.Type == metaType {
			return &tagData{
				Name:   field.Tag.Get(metaNameTagName),
				Path:   field.Tag.Get(metaPathTagName),
				Method: field.Tag.Get(metaMethodTagName),
			}
		}
	}
	return nil
}
