package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// AdaptJSONForDTO 将json转换为dto适配的格式, 并反序列化
func AdaptJSONForDTO(jsonStr string, objPtr interface{}) error {
	// 首先将jsonStr转换为map
	adaptMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonStr), &adaptMap)
	if err != nil {
		return err
	}

	// 递归的比较对象json字段信息和map的差异, 将字符和数字类型进行转换
	convertedResult, err := convert(adaptMap, objPtr)
	if err != nil {
		json.Unmarshal([]byte(jsonStr), objPtr)
	} else {
		bts, err := json.Marshal(&convertedResult)
		if err != nil {
			json.Unmarshal([]byte(jsonStr), objPtr)
		} else {
			json.Unmarshal(bts, objPtr)
		}
	}

	return nil
}

func convert(fieldObj interface{}, obj interface{}) (interface{}, error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	fieldType := reflect.TypeOf(fieldObj)
	if objType.Kind() == fieldType.Kind() {
		// the type is same, there is no need to do convertion.
		return fieldObj, nil
	}

	fieldValue := reflect.ValueOf(fieldObj)
	switch fieldType.Kind() {
	case reflect.Float64:
		if objType.Kind() == reflect.String {
			return ConvertFloat64ToString(fieldObj.(float64)), nil
		}
	case reflect.String:
		if objType.Kind() == reflect.Int || objType.Kind() == reflect.Int32 || objType.Kind() == reflect.Int64 ||
			objType.Kind() == reflect.Int8 || objType.Kind() == reflect.Int16 {
			return ConvertStringToInt64(fieldObj.(string))
		} else if objType.Kind() == reflect.Uint || objType.Kind() == reflect.Uint32 || objType.Kind() == reflect.Uint64 ||
			objType.Kind() == reflect.Uint8 || objType.Kind() == reflect.Uint16 {
			return ConvertStringToUInt64(fieldObj.(string))
		} else if objType.Kind() == reflect.Float32 || objType.Kind() == reflect.Float64 {
			return ConvertStringToFloat64(fieldObj.(string))
		}
	case reflect.Map:
		// get field json name map
		fieldJSONMap := getFieldJSONMap(obj)

		realMap := make(map[string]interface{})
		rg := fieldValue.MapRange()
		for rg.Next() {
			fieldJSONName := rg.Key().Interface().(string)
			if val, ok := fieldJSONMap[fieldJSONName]; ok {
				convertResult, err := convert(rg.Value().Interface(), val)
				if err != nil {
					return fieldObj, err
				}
				realMap[fieldJSONName] = convertResult
			} else {
				fmt.Println(fieldJSONName, " ", rg.Value().Interface())
				realMap[fieldJSONName] = rg.Value().Interface()
			}
		}
		return realMap, nil
	}

	return fieldObj, nil
}

func getFieldJSONMap(obj interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	objValue := reflect.ValueOf(obj)
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
		if !objValue.Elem().IsValid() {
			objValue = reflect.New(objType).Elem()
		} else {
			objValue = objValue.Elem()
		}
	}
	for j := 0; j < objValue.NumField(); j++ {
		jsonTag := objType.Field(j).Tag.Get("json")
		if strings.Contains(jsonTag, "string") {
			continue
		}

		jsonTag = strings.ReplaceAll(jsonTag, "omitempty", "")
		jsonTag = strings.ReplaceAll(jsonTag, ",", "")
		if jsonTag == "" {
			jsonTag = FirstToLower(objType.Field(j).Name)
		}
		ret[jsonTag] = objValue.Field(j).Interface()
	}
	return ret
}
