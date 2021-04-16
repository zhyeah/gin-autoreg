package util

import (
	"reflect"
)

// ReflectInvokeMethod 通过反射调用方法
func ReflectInvokeMethod(object interface{}, methodName string, args ...interface{}) []interface{} {

	inputs := make([]reflect.Value, len(args))
	for i, arg := range args {
		inputs[i] = reflect.ValueOf(arg)
	}

	objectValue := reflect.ValueOf(object)
	method := objectValue.MethodByName(methodName)
	ret := method.Call(inputs)

	retList := []interface{}{}
	for _, retItem := range ret {
		retList = append(retList, retItem.Interface())
	}

	return retList
}
