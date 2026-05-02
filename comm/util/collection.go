/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package util

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strings"
)

// ListTo 转换切片数据类型。
func ListTo[T1, T2 any](arr []T1, fn func(T1) T2) []T2 {
	if len(arr) <= 0 {
		return []T2{}
	}
	arr2 := make([]T2, 0, len(arr))
	for i := range arr {
		arr2 = append(arr2, fn(arr[i]))
	}
	return arr2
}

// ListTo2 转换切片数据类型。
func ListTo2[T1, T2, T3 any](arr []T1, fn func(T1) (T2, T3)) ([]T2, []T3) {
	if len(arr) <= 0 {
		return []T2{}, []T3{}
	}
	arr2 := make([]T2, 0, len(arr))
	arr3 := make([]T3, 0, len(arr))
	for i := range arr {
		t1, t2 := fn(arr[i])
		arr2 = append(arr2, t1)
		arr3 = append(arr3, t2)
	}
	return arr2, arr3
}

// ListToUnique 转换切片数据类型，并且结果去重。
func ListToUnique[T1 any, T2 comparable](arr []T1, fn func(T1) T2) []T2 {
	if len(arr) <= 0 {
		return []T2{}
	}
	arr2 := make([]T2, 0, len(arr))
	set := make(map[T2]struct{}, len(arr))
	for i := range arr {
		t := fn(arr[i])
		_, ok := set[t]
		if !ok {
			arr2 = append(arr2, t)
			set[t] = struct{}{}
		}
	}
	return arr2
}

// ListToUnique2 转换切片数据类型，并且结果去重。
func ListToUnique2[T1 any, T2, T3 comparable](arr []T1, fn func(T1) ([]T2, T3)) ([]T2, []T3) {
	if len(arr) <= 0 {
		return []T2{}, []T3{}
	}
	arr2 := make([]T2, 0, len(arr))
	arr3 := make([]T3, 0, len(arr))
	set := make(map[T2]struct{}, len(arr))
	set2 := make(map[T3]struct{}, len(arr))
	for i := range arr {
		t1, t2 := fn(arr[i])
		if len(t1) > 0 {
			for _, v := range t1 {
				if _, ok := set[v]; !ok {
					arr2 = append(arr2, v)
					set[v] = struct{}{}
				}
			}
		}
		if _, ok := set2[t2]; !ok {
			arr3 = append(arr3, t2)
			set2[t2] = struct{}{}
		}
	}
	return arr2, arr3
}

// SeqListTo 转换切片数据类型。
func SeqListTo[T1, T2 any](seq iter.Seq[T1], fn func(T1) T2) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		for t1 := range seq {
			if !yield(fn(t1)) {
				break
			}
		}
	}
}

// ListToMap 切片转换成映射。
func ListToMap[T1 any, T2 comparable, T3 any](arr []T1, fn func(T1) (T2, T3)) map[T2]T3 {
	if len(arr) <= 0 {
		return map[T2]T3{}
	}
	m := make(map[T2]T3, len(arr))
	for i := range arr {
		k, v := fn(arr[i])
		m[k] = v
	}
	return m
}

// ListAssociateBy 切片转换成映射。
func ListAssociateBy[T1 any, T2 comparable](arr []T1, fn func(T1) T2) map[T2]T1 {
	if len(arr) <= 0 {
		return map[T2]T1{}
	}
	m := make(map[T2]T1, len(arr))
	for i := range arr {
		m[fn(arr[i])] = arr[i]
	}
	return m
}

// ListFilter 过滤切片。
func ListFilter[T any](arr []T, fn func(T) bool) []T {
	arr2 := make([]T, 0, len(arr))
	for i := range arr {
		if fn(arr[i]) {
			arr2 = append(arr2, arr[i])
		}
	}
	return arr2
}

// MapFilter 过滤映射。
func MapFilter[K comparable, V any](m map[K]V, fn func(K, V) bool) map[K]V {
	m2 := make(map[K]V, len(m))
	for k, v := range m {
		if fn(k, v) {
			m2[k] = v
		}
	}
	return m2
}

// MapToList 映射转成切片。
func MapToList[T any, K comparable, V any](m map[K]V, fn func(K, V) T) []T {
	arr := make([]T, 0, len(m))
	for k, v := range m {
		arr = append(arr, fn(k, v))
	}
	return arr
}

// Join 合并成字符串。
func Join[E any](arr []E, sep string) string {
	if len(arr) <= 0 {
		return ""
	}
	str := strings.Builder{}
	for i := 0; i < len(arr)-1; i++ {
		_, _ = fmt.Fprintf(&str, "%v%s", arr[i], sep)
	}
	_, _ = fmt.Fprintf(&str, "%v", arr[len(arr)-1])
	return str.String()
}

// IsZero 任何零值和空容器都返回 true
func IsZero(v any) bool {
	if v == nil {
		return true
	}
	switch value := v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return value == 0
	case float32, float64:
		return value == 0.0
	case bool:
		return value == false
	case string:
		return len(value) == 0
	case complex64, complex128:
		return value == 0i
	default:
		reflectValue := reflect.ValueOf(value)
		switch reflectValue.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return reflectValue.Len() == 0
		case reflect.Interface, reflect.Pointer, reflect.Func, reflect.Chan, reflect.UnsafePointer:
			return reflectValue.IsNil()
		case reflect.Struct:
			return reflectValue.IsZero()
		}
	}
	return false
}

// ListDropZero 去除切片中的零值元素。
func ListDropZero[E any](arr []E) []E {
	arr2 := make([]E, 0, len(arr))
	for i := range arr {
		if IsZero(arr[i]) {
			continue
		}
		arr2 = append(arr2, arr[i])
	}
	return arr2
}

// CleanStrings 清理重复和空串元素。
func CleanStrings(arr []string) []string {
	arr2 := make([]string, 0, len(arr))
	set := make(map[string]struct{}, len(arr))
	for i := range arr {
		if len(arr[i]) <= 0 {
			continue
		}
		if _, ok := set[arr[i]]; ok {
			continue
		}
		arr2 = append(arr2, arr[i])
		set[arr[i]] = struct{}{}
	}
	return arr2
}

// CleanNumbers 清理重复和零元素。
func CleanNumbers[T Number](arr []T) []T {
	arr2 := make([]T, 0, len(arr))
	set := make(map[T]struct{}, len(arr))
	for i := range arr {
		if arr[i] <= 0 {
			continue
		}
		if _, ok := set[arr[i]]; ok {
			continue
		}
		arr2 = append(arr2, arr[i])
		set[arr[i]] = struct{}{}
	}
	return arr2
}

// ContainsAny 切片是否包含任意一个元素。
func ContainsAny[T comparable](arr []T, elems ...T) bool {
	for i := range arr {
		if slices.Contains(elems, arr[i]) {
			return true
		}
	}
	return false
}

// ContainsAll 切片是否包含了所有元素。
func ContainsAll[T comparable](arr []T, elems ...T) bool {
	for i := range elems {
		if !slices.Contains(arr, elems[i]) {
			return false
		}
	}
	return true
}

// In 是否存在。
func In[T comparable](elem T, elems ...T) bool {
	return slices.Contains(elems, elem)
}
