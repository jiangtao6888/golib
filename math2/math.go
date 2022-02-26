package math2

import (
	"math"
)

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
	Signed | Unsigned
}

type Float interface {
	~float32 | ~float64
}

type Number interface {
	Integer | Float
}

func Max[T Number](ns ...T) T {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func Min[T Number](ns ...T) T {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func Round[T Float, V Integer](x T) V {
	return V(math.Floor(float64(x) + 0.5))
}

func RoundN[T Float](f T, n int) T {
	n10 := math.Pow10(n)
	return T(math.Trunc(float64(f)*n10+0.5) / n10)
}

func Ceil[T Integer](n, m T) T {
	v := n / m

	if v*m < n {
		return v + 1
	}

	return v
}

func IIf[T Number](b bool, n, m T) T {
	if b {
		return n
	}

	return m
}

func Range[T Integer](start, end T) []T {
	nums := make([]T, end-start+1)

	for n := start; n <= end; n++ {
		nums[n-start] = n
	}

	return nums
}

func InList[T Number](value T, list []T) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func SameList[T Number](a1, a2 []T) bool {
	if len(a1) != len(a2) {
		return false
	}

	for k, v := range a1 {
		if a2[k] != v {
			return false
		}
	}

	return true
}

func TrimList[T Number](value, removeList []T) []T {
	removeMap := make(map[T]bool, len(removeList))

	for _, v := range removeList {
		removeMap[v] = true
	}

	result := make([]T, 0, len(value))

	for _, v := range value {
		if !removeMap[v] {
			result = append(result, v)
		}
	}

	return result
}

func UniqueList[T Number](list []T) []T {
	result := make([]T, 0, len(list))
	flags := make(map[T]bool, len(list))

	for _, v := range list {
		if !flags[v] {
			result = append(result, v)
		}

		flags[v] = true
	}

	return result
}

func SumList[T Number](list []T) (sum T) {
	for _, v := range list {
		sum += v
	}

	return
}

func AvgList[T Number](list []T) (avg T) {
	if n := len(list); n > 0 {
		return SumList(list) / T(n)
	}

	return
}

func Percent[T Number, V Float](num, denom T, decimal int) V {
	if denom <= 0 {
		return 0
	}

	return V(RoundN(float64(num*100)/float64(denom), decimal))
}
