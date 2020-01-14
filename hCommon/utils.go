package hCommon

import (
	"bytes"
	"crypto/md5"
	"custom/happy/hLog"
	"encoding/hex"
	"errors"
	"github.com/bwmarrin/snowflake"
	"math/rand"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	_        = iota
	KB int64 = 1 << (iota * 10)
	MB
	GB
)

func CheckError() {
	if r := recover(); r != nil {
		var str string
		switch r.(type) {
		case error:
			str = r.(error).Error()
		case string:
			str = r.(string)
		}
		err := errors.New("\n" + str + "\n" + string(debug.Stack()))
		hLog.Error(err)
	}
}

func IsExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return IsExported(t.Name()) || t.PkgPath() == ""
}

//条件等待
func When(interval time.Duration, conditions ...func() bool) {
	var pass bool
	for {
		pass = true
		for _, cond := range conditions {
			if !cond() {
				pass = false
				break
			}
		}
		if pass {
			break
		}
		time.Sleep(interval)
	}
}

type Interface interface {
	DeepCopy() interface{}
}

func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	original := reflect.ValueOf(src)
	cpy := reflect.New(original.Type()).Elem()
	copyRecursive(original, cpy)

	return cpy.Interface()
}

func copyRecursive(src, dst reflect.Value) {
	if src.CanInterface() {
		if copier, ok := src.Interface().(Interface); ok {
			dst.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	switch src.Kind() {
	case reflect.Ptr:
		originalValue := src.Elem()

		if !originalValue.IsValid() {
			return
		}
		dst.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		originalValue := src.Elem()
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(src.Field(i), dst.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyRecursive(src.Index(i), dst.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := Copy(key.Interface())
			dst.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		dst.Set(src)
	}
}

func MD5(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func GenRandom(start int, end int, count int) []int {
	end += 1
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

/*  流程执行  */

type Procedure struct {
	Task      func()
	Condition func() bool
}

func StartProcedure(checkInterval time.Duration, tasks ...*Procedure) {
	for _, task := range tasks {
		When(checkInterval, task.Condition)
		Try(task.Task)
	}
}

var pool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func StrToBytes(strData string) []byte {
	buffer := pool.Get().(*bytes.Buffer)

	defer pool.Put(buffer)

	buffer.Reset()
	buffer.WriteString(strData)

	return buffer.Bytes()
}

func BytesToStr(b []byte) string {
	buffer := pool.Get().(*bytes.Buffer)

	defer pool.Put(buffer)

	buffer.Reset()
	buffer.Write(b)

	return buffer.String()
}

func FormatMem(mem float64) string {
	d := int64(mem)
	if d < KB {
		return strconv.FormatInt(d, 10) + "KB"
	} else if d >= KB && d < MB {
		return strconv.FormatInt(d/KB, 10) + "KB"
	} else if d >= MB && d < GB {
		return strconv.FormatInt(d/MB, 10) + "MB"
	}
	return strconv.FormatInt(d/GB, 10) + "GB"
}

func Contains(arr interface{}, ele interface{}) bool {
	tmp := reflect.ValueOf(arr)
	switch reflect.TypeOf(arr).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < tmp.Len(); i++ {
			if tmp.Index(i).Interface() == ele {
				return true
			}
		}
	case reflect.Map:
		if tmp.MapIndex(reflect.ValueOf(ele)).IsValid() {
			return true
		}
	}
	return false
}

func GenId(id int64) int64 {
	node, err := snowflake.NewNode(id)
	if err != nil {
		return -1
	}
	return node.Generate().Int64()
}
