// 轻量级单对象并行运行池
// 会产生一定的数据COPY
// version 1.0 beta
// by koangel
// email: jackliu100@gmail.com
// 2017/12/18
package Utils

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type jobhandler interface{}
type jobElement struct {
	handler    jobhandler
	needResult bool
	resultcall jobhandler
	in         []reflect.Value
}

type SyncJob struct {
	wait sync.WaitGroup

	inter []*jobElement
}

func (job *SyncJob) Append(fn jobhandler, args ...interface{}) error {
	t := reflect.TypeOf(fn) // 获得对象类型
	argArr := []interface{}(args)

	if t.Kind() != reflect.Func {
		return errors.New("Handler must be function...")
	}

	// 数量和参数对不上
	if len(argArr) < t.NumIn() {
		return errors.New("Not enough arguments")
	}

	// 逐步压入参数
	in := make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		if argType != reflect.TypeOf(argArr[i]) {
			return errors.New(fmt.Sprintf("Value not found for type %v", argType))
		}
		in[i] = reflect.ValueOf(argArr[i]) // 完成一个基本的CALL
	}

	job.inter = append(job.inter, &jobElement{
		handler:    fn,
		needResult: false,
		in:         in,
	})

	job.wait.Add(1)

	return nil
}

func (job *SyncJob) AppendR(fn, rcall jobhandler, args ...interface{}) error {
	t := reflect.TypeOf(fn) // 获得对象类型
	rt := reflect.TypeOf(rcall)

	if t.Kind() != reflect.Func || rt.Kind() != reflect.Func {
		return errors.New("Handler or Result Caller must be function...")
	}

	argArr := []interface{}(args)

	// 数量和参数对不上
	if len(argArr) < t.NumIn() {
		return errors.New("Not enough arguments")
	}

	// 返回Call的参数对不上
	if t.NumOut() != rt.NumIn() {
		return errors.New("Result Not enough arguments")
	}

	// 逐步压入参数
	in := make([]reflect.Value, t.NumIn()) //Panic if t is not kind of Func
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		if argType != reflect.TypeOf(argArr[i]) {
			return errors.New(fmt.Sprintf("Value not found for type %v", argType))
		}
		in[i] = reflect.ValueOf(argArr[i]) // 完成一个基本的CALL
	}

	job.inter = append(job.inter, &jobElement{
		handler:    fn,
		resultcall: rcall,
		needResult: true,
		in:         in,
	})

	job.wait.Add(1)

	return nil
}

func (job *SyncJob) proc(e *jobElement) {

	rv := reflect.ValueOf(e.handler).Call(e.in)
	if e.needResult {
		reflect.ValueOf(e.resultcall).Call(rv) // 调用callback传值
	}

	job.wait.Done() // 完成一次
}

func (job *SyncJob) StartWait() {

	for _, e := range job.inter {
		go job.proc(e) // 开始
	}

	job.wait.Wait()
}
