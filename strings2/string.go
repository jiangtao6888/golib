package strings2

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	regHtmlTag   = regexp.MustCompile("\\<[\\S\\s]+?\\>")
	regHtmlSpace = regexp.MustCompile("\\s{2,}")
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

func InArray(val string, list []string) (exists bool) {
	exists = false

	for _, v := range list {
		if val == v {
			exists = true
			break
		}
	}

	return
}

func ToNsArray(s, separator string) []string {
	items := strings.Split(s, separator)
	numbers := make([]string, 0, len(items))

	for _, v := range items {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			numbers = append(numbers, strconv.FormatInt(n, 10))
		}
	}

	return numbers
}

func ToIntArray(s, separator string) []int {
	items := strings.Split(s, separator)
	numbers := make([]int, 0, len(items))

	for _, v := range items {
		if n, e := strconv.Atoi(v); e == nil {
			numbers = append(numbers, n)
		}
	}

	return numbers
}

func FromIntArray(nums []int, separator string) string {
	items := make([]string, 0, len(nums))

	for _, v := range nums {
		items = append(items, strconv.Itoa(v))
	}

	return strings.Join(items, separator)
}

func Unique(list []string) []string {
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

	for v := range set {
		list = append(list, v)
	}

	return list
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func ShortNumStr(n int64) string {
	m, s := int64(0), make([]rune, 0, 8)

	for n > 0 {
		n, m = n/int64(len(letterRunes)), n%int64(len(letterRunes))
		s = append(s, letterRunes[m])
	}

	return string(s)
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

func TrimHtmlTags(s string) string {
	s = regHtmlTag.ReplaceAllString(s, "\n")
	s = regHtmlSpace.ReplaceAllString(s, "\n")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}
