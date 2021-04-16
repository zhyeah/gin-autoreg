package util

import (
	"strconv"
	"strings"
)

// ConvertStringToInt string 转 int
func ConvertStringToInt(str string) (int, error) {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ConvertStringToUInt64 string 转 uint64
func ConvertStringToUInt64(str string) (uint64, error) {
	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ConvertStringToUInt32 string 转 uint32
func ConvertStringToUInt32(str string) (uint32, error) {
	val, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(val), nil
}

// ConvertStringToUInt16 string 转 uint16
func ConvertStringToUInt16(str string) (uint16, error) {
	val, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(val), nil
}

// ConvertStringToUInt8 string 转 uint8
func ConvertStringToUInt8(str string) (uint8, error) {
	val, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(val), nil
}

// ConvertStringToInt64 string 转 int64
func ConvertStringToInt64(str string) (int64, error) {
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ConvertStringToInt32 string 转 int32
func ConvertStringToInt32(str string) (int32, error) {
	val, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

// ConvertStringToInt16 string 转 int16
func ConvertStringToInt16(str string) (int16, error) {
	val, err := strconv.ParseInt(str, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(val), nil
}

// ConvertStringToInt8 string 转 int8
func ConvertStringToInt8(str string) (int8, error) {
	val, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(val), nil
}

// ConvertStringToFloat64 string 转 float64
func ConvertStringToFloat64(str string) (float64, error) {
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ConvertStringToFloat32 string 转 float32
func ConvertStringToFloat32(str string) (float32, error) {
	val, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, err
	}
	return float32(val), nil
}

// ConvertStringToBool 转换 string 到 bool
func ConvertStringToBool(str string) (bool, error) {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return val, nil
}

// ConvertStringToBoolDefault 转换 string 到 bool
func ConvertStringToBoolDefault(str string, def bool) bool {
	val, err := strconv.ParseBool(str)
	if err != nil {
		return def
	}
	return val
}

// ConvertIntToString 转换整数为字符串
func ConvertIntToString(n int) string {
	return strconv.Itoa(n)
}

// ConvertUint64ToString 转uint64到字符串
func ConvertUint64ToString(n uint64) string {
	return strconv.FormatUint(n, 10)
}

// FirstToLower 转换首字母为小写字母
func FirstToLower(input string) string {
	return strings.ToLower(input[0:1]) + input[1:]
}
