package utils

import (
	"reflect"
)

type DbReflector struct {
	Val interface{}
}

func newDbReflector(val interface{}) DbReflector {
	return DbReflector{
		Val: val,
	}
}

func (r *DbReflector) getType() reflect.Type {
	typeOf := reflect.TypeOf(r.Val)
	switch typeOf.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return typeOf.Elem()
	default:
		return typeOf
	}
}

func (r *DbReflector) getValue() reflect.Value {
	val := reflect.ValueOf(r.Val)
	switch val.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return val.Elem()
	default:
		return val
	}
}

func (r *DbReflector) getTaggedIndices() []int {
	var indices []int

	typeElem := r.getType()
	numFields := typeElem.NumField()

	for i := 0; i <= numFields-1; i++ {
		structType := typeElem.Field(i)
		_, ok := structType.Tag.Lookup("db")
		if !ok {
			continue
		}

		indices = append(indices, i)
	}

	return indices
}

func (r *DbReflector) ReflectColumns() []string {
	indices := r.getTaggedIndices()
	var fields []string

	for _, index := range indices {
		structType := r.getType().Field(index)
		value := structType.Tag.Get("db")
		fields = append(fields, value)
	}

	return fields
}

func (r *DbReflector) ReflectValues() []interface{} {
	indices := r.getTaggedIndices()
	var values []interface{}

	for _, index := range indices {
		value := r.getValue().Field(index)
		values = append(values, value.Interface())
	}

	return values
}

func ReflectValues(v interface{}) []interface{} {
	reflector := newDbReflector(v)

	return reflector.ReflectValues()
}

func ReflectColumns(v interface{}) []string {
	reflector := newDbReflector(v)

	return reflector.ReflectColumns()
}
