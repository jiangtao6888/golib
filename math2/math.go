package math2

import "math"

func Max(ns ...int) int {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func Min(ns ...int) int {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func MaxInt64(ns ...int64) int64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func MinInt64(ns ...int64) int64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func MaxFloat(ns ...float64) float64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] > m {
			m = ns[i]
		}
	}

	return m
}

func MinFloat(ns ...float64) float64 {
	m := ns[0]

	for i := 1; i < len(ns); i++ {
		if ns[i] < m {
			m = ns[i]
		}
	}

	return m
}

func Round(x float64) int {
	return int(math.Floor(x + 0.5))
}

func RoundFloat(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc(f*n10+0.5) / n10
}

func Ceil(n, m int) int {
	v := n / m

	if v*m < n {
		return v + 1
	}

	return v
}

func CeilInt64(n, m int64) int64 {
	v := n / m

	if v*m < n {
		return v + 1
	}

	return v
}

func IIf(b bool, n, m int) int {
	if b {
		return n
	}

	return m
}

func IIfInt64(b bool, n, m int64) int64 {
	if b {
		return n
	}

	return m
}

func IIfFloat(b bool, n, m float64) float64 {
	if b {
		return n
	}

	return m
}

func Range(start int, end int) []int {
	nums := make([]int, end-start+1)

	for n := start; n <= end; n++ {
		nums[n-start] = n
	}

	return nums
}

func InList(value int, list []int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func InInt64List(value int64, list []int64) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func InFloatList(value float64, list []float64) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func SameInt64List(a1 []int64, a2 []int64) bool {
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

func TrimInt64List(value []int64, removeList []int64) []int64 {
	result := make([]int64, 0, len(value))

	for _, v := range value {
		if InInt64List(v, removeList) {
			continue
		}
		result = append(result, v)
	}

	return result
}

func UniqueList(list []int) []int {
	result := make([]int, 0, len(list))
	flags := make(map[int]bool, len(list))

	for _, v := range list {
		if !flags[v] {
			result = append(result, v)
		}

		flags[v] = true
	}

	return result
}

func UniqueInt64List(list []int64) []int64 {
	result := make([]int64, 0, len(list))
	flags := make(map[int64]bool, len(list))

	for _, v := range list {
		if !flags[v] {
			result = append(result, v)
		}

		flags[v] = true
	}

	return result
}

func SumInt64List(list []int64) (sum int64) {
	for _, v := range list {
		sum += v
	}

	return
}

func AvgInt64List(list []int64) (avg int64) {
	if n := len(list); n > 0 {
		return SumInt64List(list) / int64(n)
	}

	return
}

func Percent(num, denom int64, decimal int) float64 {
	if denom <= 0 {
		return 0
	}

	return RoundFloat(float64(num*100)/float64(denom), decimal)
}
