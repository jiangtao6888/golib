package strings2

import (
	"fmt"
	"github.com/marsmay/golib/math2"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Params map[string]interface{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func IIf(b bool, n, m string) string {
	if b {
		return n
	}

	return m
}

func Select(n, m string) string {
	return IIf(n != "", n, m)
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func InList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}

func ToIntList[T math2.Integer](s, separator string) []T {
	items := strings.Split(s, separator)
	numbers := make([]T, 0, len(items))

	for _, v := range items {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			numbers = append(numbers, T(n))
		}
	}

	return numbers
}

func FromIntList[T math2.Integer](nums []T, separator string) string {
	items := make([]string, 0, len(nums))

	for _, v := range nums {
		items = append(items, strconv.FormatInt(int64(v), 10))
	}

	return strings.Join(items, separator)
}

func UniqueList(list []string) []string {
	result := make([]string, 0, len(list))
	flags := make(map[string]bool, len(list))

	for _, v := range list {
		if !flags[v] {
			result = append(result, v)
		}

		flags[v] = true
	}

	return result
}

func ListToSet(list []string) map[string]bool {
	set := make(map[string]bool, len(list))

	for _, v := range list {
		set[v] = true
	}

	return set
}

func SetToList(set map[string]bool) []string {
	list := make([]string, 0, len(set))

	for k, v := range set {
		if v {
			list = append(list, k)
		}
	}

	return list
}

func HasPrefixs(s string, prefixs []string) (hit bool, prefix string) {
	for _, v := range prefixs {
		if strings.HasPrefix(s, v) {
			return true, v
		}
	}

	return
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func Template(t string, params map[string]interface{}) string {
	pairs := make([]string, 0, 2*len(params))

	for k, v := range params {
		pairs = append(pairs, "{"+k+"}")

		if s, ok := v.(string); ok {
			pairs = append(pairs, s)
		} else {
			pairs = append(pairs, fmt.Sprintf("%v", v))
		}
	}

	return strings.NewReplacer(pairs...).Replace(t)
}

func Reverse(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}
