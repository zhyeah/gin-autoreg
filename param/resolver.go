package param

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/zhyeah/gin-autoreg/util"
)

const (
	FROM_QUERY    = "query"
	FROM_PATH     = "path"
	FROM_FORMDATA = "form"
	FROM_BODY     = "body"
	FROM_CONTEXT  = "context"
)

// FieldInfo 字段信息
type FieldInfo struct {
	Field        reflect.Value
	FieldName    string
	Name         string
	From         string
	DefaultValue string
	Type         reflect.Type
	MustHave     bool
}

// ResolvePostDataJson resolve json of controler post data
func ResolvePostDataJson(ctrl interface{}, methodName string) (string, error) {
	// get method by method name
	method := reflect.ValueOf(ctrl).MethodByName(methodName)

	// find out the specified param
	methodType := method.Type()
	for i := 0; i < methodType.NumIn(); i++ {
		inType := methodType.In(i)
		if inType.Elem().Kind() == reflect.Struct {
			paramInstancePtr := reflect.New(inType.Elem())
			instanceType := paramInstancePtr.Elem().Type()
			for j := 0; j < instanceType.NumField(); j++ {
				from := instanceType.Field(j).Tag.Get("from")
				if from == FROM_BODY {
					fVal := paramInstancePtr.Elem().Field(j)
					fType := instanceType.Field(j).Type
					switch fType.Kind() {
					case reflect.Slice, reflect.Map, reflect.Struct:
						bts, err := json.Marshal(fVal.Addr().Interface())
						return string(bts), err
					case reflect.Ptr:
						if fType.Elem().Kind() == reflect.Struct {
							val := reflect.New(fType.Elem()).Interface()
							bts, err := json.Marshal(val)
							return string(bts), err
						}
					}
				}
			}
		}
	}
	return "", nil
}

// ResolveParams 解析controller action需要的参数
func ResolveParams(ctrl interface{}, methodName string, ctx *gin.Context) ([]interface{}, error) {
	ret := make([]interface{}, 0)

	// 获取ctrl的methodName的方法
	method := reflect.ValueOf(ctrl).MethodByName(methodName)

	// 遍历method的参数，如果是*struct类型，对字段进行解析，如果是*gin.Context类型，直接塞入，其他类型则抛异常
	methodType := method.Type()
	for i := 0; i < methodType.NumIn(); i++ {
		inType := methodType.In(i)
		if inType.Kind().String() != reflect.Ptr.String() {
			return nil, errors.New("the parameters of method should be ptr")
		}
		if inType.Elem().PkgPath() == "github.com/gin-gonic/gin" && inType.Elem().Name() == "Context" {
			ret = append(ret, ctx)
		} else if inType.Elem().Kind().String() == reflect.Struct.String() {
			// 首先new一个instance
			paramInstancePtr := reflect.New(inType.Elem())
			instanceType := paramInstancePtr.Elem().Type()
			for j := 0; j < instanceType.NumField(); j++ {
				err := SetFieldValue(&FieldInfo{
					Field:        paramInstancePtr.Elem().Field(j),
					FieldName:    instanceType.Field(j).Name,
					Name:         instanceType.Field(j).Tag.Get("field"),
					From:         instanceType.Field(j).Tag.Get("from"),
					DefaultValue: instanceType.Field(j).Tag.Get("default"),
					MustHave:     util.ConvertStringToBoolDefault(instanceType.Field(j).Tag.Get("must"), true),
					Type:         instanceType.Field(j).Type,
				}, ctx)

				if err != nil {
					return nil, err
				}
			}
			ret = append(ret, paramInstancePtr.Interface())
		} else {
			return nil, fmt.Errorf("unsupport controller param ptr type %s", inType.Elem().Name())
		}
	}

	return ret, nil
}

// SetFieldValue 根据字段信息, 设置gin.Context中的值
func SetFieldValue(fieldInfo *FieldInfo, ctx *gin.Context) error {
	if fieldInfo.From == "" {
		return nil
	}

	// 根据字段类型设置value
	switch fieldInfo.Type.Kind().String() {
	case reflect.Int.String(), reflect.Int8.String(), reflect.Int16.String(), reflect.Int32.String(), reflect.Int64.String():
		valStr := getValueFromContext(fieldInfo, ctx)
		intVal, err := util.ConvertStringToInt64(valStr)
		if err != nil && fieldInfo.MustHave {
			return fmt.Errorf("field '%s' val '%s' cannot convert to int", fieldInfo.Name, valStr)
		}
		fieldInfo.Field.SetInt(intVal)
	case reflect.Uint.String(), reflect.Uint8.String(), reflect.Uint16.String(), reflect.Uint32.String(), reflect.Uint64.String():
		valStr := getValueFromContext(fieldInfo, ctx)
		intVal, err := util.ConvertStringToUInt64(valStr)
		if err != nil && fieldInfo.MustHave {
			return fmt.Errorf("field '%s' val '%s' cannot convert to unsigned int", fieldInfo.Name, valStr)
		}
		fieldInfo.Field.SetUint(intVal)
	case reflect.Bool.String():
		valStr := getValueFromContext(fieldInfo, ctx)
		boolVal, err := util.ConvertStringToBool(valStr)
		if err != nil && fieldInfo.MustHave {
			return fmt.Errorf("field '%s' val '%s' cannot convert to bool", fieldInfo.Name, valStr)
		}
		fieldInfo.Field.SetBool(boolVal)
	case reflect.String.String():
		valStr := getValueFromContext(fieldInfo, ctx)
		if valStr == "" && fieldInfo.MustHave {
			return fmt.Errorf("field '%s' must have val, but now it's empty", fieldInfo.Name)
		}
		fieldInfo.Field.SetString(valStr)
	case reflect.Slice.String(), reflect.Map.String(), reflect.Struct.String():
		val := fieldInfo.Field.Addr().Interface()
		defer ctx.Request.Body.Close()
		body, _ := ioutil.ReadAll(ctx.Request.Body)
		err := util.AdaptJSONForDTO(string(body), val)
		if err != nil {
			return err
		}
		fieldInfo.Field.Set(reflect.ValueOf(val).Elem())
	case reflect.Ptr.String():
		if fieldInfo.Type.Elem().Kind().String() == reflect.Struct.String() {
			val := reflect.New(fieldInfo.Type.Elem()).Interface()
			// 这里手动强制适配
			defer ctx.Request.Body.Close()
			body, _ := ioutil.ReadAll(ctx.Request.Body)
			err := util.AdaptJSONForDTO(string(body), val)
			if err != nil {
				return err
			}
			fieldInfo.Field.Set(reflect.ValueOf(val))
		} else if fieldInfo.Type.Elem().Kind().String() == reflect.Map.String() {
			keyType := fieldInfo.Type.Elem().Key()
			valType := fieldInfo.Type.Elem().Elem()
			val := reflect.New(reflect.MapOf(keyType, valType)).Interface()
			defer ctx.Request.Body.Close()
			body, _ := ioutil.ReadAll(ctx.Request.Body)
			err := util.AdaptJSONForDTO(string(body), val)
			if err != nil {
				return err
			}
			fieldInfo.Field.Set(reflect.ValueOf(val))
		} else if fieldInfo.Type.Elem().Kind().String() == reflect.Slice.String() {
			listType := fieldInfo.Type.Elem().Elem()
			val := reflect.New(reflect.SliceOf(listType)).Interface()
			defer ctx.Request.Body.Close()
			body, _ := ioutil.ReadAll(ctx.Request.Body)
			err := util.AdaptJSONForDTO(string(body), val)
			if err != nil {
				return err
			}
			fieldInfo.Field.Set(reflect.ValueOf(val))
		}
	}
	return nil
}

func getValueFromContext(fieldInfo *FieldInfo, ctx *gin.Context) string {
	if fieldInfo.Name == "" {
		fieldInfo.Name = util.FirstToLower(fieldInfo.FieldName)
	}
	switch fieldInfo.From {
	case FROM_QUERY:
		return ctx.DefaultQuery(fieldInfo.Name, fieldInfo.DefaultValue)
	case FROM_PATH:
		return ctx.Param(fieldInfo.Name)
	case FROM_FORMDATA:
		return ctx.DefaultPostForm(fieldInfo.Name, fieldInfo.DefaultValue)
	case FROM_CONTEXT:
		return ctx.GetString(fieldInfo.Name)
	}
	return fieldInfo.DefaultValue
}
