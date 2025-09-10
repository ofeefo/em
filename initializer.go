package em

import (
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
)

func MustInit[T any](attrs ...attribute.KeyValue) *T {
	res, err := Init[T](attrs...)
	if err != nil {
		panic(err)
	}
	return res
}

func Init[T any](attrs ...attribute.KeyValue) (*T, error) {
	base := new(T)
	if err := initRef(base, attrs...); err != nil {
		return nil, err
	}
	return base, nil
}

func initRef(base any, attrs ...attribute.KeyValue) error {
	bType := reflect.TypeOf(base)
	bVal := reflect.ValueOf(base)

	if bType.Kind() == reflect.Pointer {
		bType = bType.Elem()
		bVal = bVal.Elem()
	}

	if bType.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct type, got %s", bType.Kind().String())
	}

	for i := 0; i < bType.NumField(); i++ {
		field := bType.Field(i)
		fieldVal := bVal.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldIface := fieldVal.Interface()
		builder, ok := fieldIface.(buildable)

		switch {
		case ok:
			if err := builder.init(field, attrs...); err != nil {
				return err
			}
			fieldVal.Set(reflect.ValueOf(builder))

		case fieldVal.Kind() == reflect.Struct:
			if fieldVal.CanAddr() {
				p := fieldVal.Addr()
				if err := initNested(field, p, fieldVal, attrs...); err != nil {
					return err
				}
			} else {
				p := reflect.New(fieldVal.Type())
				if err := initNested(field, p, fieldVal, attrs...); err != nil {
					return err
				}
			}

		case fieldVal.Kind() == reflect.Ptr:
			n := reflect.New(field.Type.Elem())
			if err := initNested(field, n, fieldVal, attrs...); err != nil {
				return err
			}
		}
	}
	return nil
}

func initNested(field reflect.StructField, newVal, fieldVal reflect.Value, attrs ...attribute.KeyValue) error {
	innerAttrs, err := getAttrs(field)
	if err != nil {
		return err
	}

	if err = initRef(newVal.Interface(), append(attrs, innerAttrs...)...); err != nil {
		return err
	}

	if fieldVal.Kind() == reflect.Ptr {
		fieldVal.Set(newVal)
	} else {
		fieldVal.Set(newVal.Elem())
	}

	return nil
}
