package main

import (
	"context"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

/*
一般的RPC框架中都会用到反射机制，比如服务提供端使用反射来发布服务，服务消费方调用的时候内部也会用到反射

既然函数可以像普通的类型变量一样可以的话，那么在反射机制中就和不同的变量一样的，
在反射中函数和方法的类型（Type）都是reflect.Func，如果要调用函数的话，可以通过Value的Call方法，例如：

*/

type CmdIn struct {
	Param string
}

func (cmdIn CmdIn) String() string {
	return fmt.Sprintf("[CmdIn Param is %s]", cmdIn.Param)
}

type CmdOut struct {
	Result int
	Info   string
}

func (cmdOut CmdOut) String() string {
	return fmt.Sprintf("Result [%d] Info[%s]", cmdOut.Result, cmdOut.Info)
}

// 首先定义服务类Hello
type Hello struct {
	Param string
}

func (this *Hello) Test1(ctx context.Context, in *CmdIn, out *CmdOut) error {
	fmt.Println("接收到服务消费方输入：", in.Param)
	out.Result = 12345678
	out.Info = "我已经吃饭了!"
	return nil
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

type MethodType struct {
	method   reflect.Method
	ArgType  reflect.Type
	ReplyTpe reflect.Type
}

type Reflect struct {
	Name   string
	Rcvr   reflect.Value
	Typ    reflect.Type
	Method map[string]*MethodType
}

func (this *Reflect) Setup(rcvr interface{}, name string) {

	this.Typ = reflect.TypeOf(rcvr)
	this.Rcvr = reflect.ValueOf(rcvr)
	this.Name = reflect.Indirect(this.Rcvr).Type().Name()
	this.Method = suitableMethods(this.Typ, true)
}

func suitableMethods(typ reflect.Type, reportErr bool) map[string]*MethodType {
	methods := make(map[string]*MethodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if method.PkgPath != "" {
			continue
		}

		if mtype.NumIn() != 4 {
			if reportErr {
				fmt.Println("method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}

		ctxtType := mtype.In(1)
		if !ctxtType.Implements(typeOfContext) {
			if reportErr {

				fmt.Println("method", mname, "must use context.Contet as the first parameter")
			}
			continue
		}

		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				fmt.Println(mname, "parameter type not exported:", argType)
			}
			continue
		}
		// Third arg must be a pointer.
		replyType := mtype.In(3)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				fmt.Println("method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				fmt.Println("method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				fmt.Println("method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				fmt.Println("method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &MethodType{method: method, ArgType: argType, ReplyTpe: replyType}
	}

	return methods
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fmt.Println(t.PkgPath())
	return isExported(t.Name()) || t.PkgPath() == ""
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)

}

func (this *Reflect) Call(ctx context.Context, service string, method string, in *CmdIn, out *CmdOut) error {
	mtype := this.Method[method]
	returnValues := mtype.method.Func.Call([]reflect.Value{this.Rcvr, reflect.ValueOf(ctx), reflect.ValueOf(in), reflect.ValueOf(out)})
	errInter := returnValues[0].Interface()
	if errInter != nil {
		return errInter.(error)
	} else {
		return nil
	}
}

func main() {

	hello := &Hello{}

	reflect := &Reflect{}

	reflect.Setup(hello, "Hello")

	fmt.Println("发起消费方调用")

	fmt.Println("发起消费方调用")
	//定义输入参数
	in := &CmdIn{
		Param: "吃饭了吗?",
	}
	//定义输出参数
	out := &CmdOut{}
	//消费方访问服务提供方hello函数Test1
	err := reflect.Call(context.Background(), "Hello", "Test1", in, out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("接收到服务提供方返回:", out)

	}
	// 服务消费端并不是直接调用hello.Test1函数，而是通过reflect.Call方式调用，传入服务名Hello和需要访问的函数Test1

	/*
		reflect.ValueOf() 返回值类型：reflect.Value 反射值对象
		reflect.TypeOf() 返回值类型：reflect.Type 反射类型对象
	*/

	/*	var ss string

		t := reflect.TypeOf(ss)
		fmt.Println(t)
		sptr := reflect.New(t)
		fmt.Printf("%s\n",sptr)
		// reflect.New(）返回的是一个指针,可以通过 reflect.Value.Elem()来取得其实际的值
		sval := sptr.Elem() // 返回值类型 reflect.value

		ss2 := sval.Interface().(string)

		fmt.Println(ss2)

		// reflect.Indirect(reflect.ValueOf(someX)) === reflect.ValueOf(someX).Elem()
	*/

	/*
		反射与间接访问Elem()
		reflect.Value的Elem方法可以解决上面的问题,如果reflect.Value内部存储值为指针或接口,则Elem方法返回指针或接口指向的数据。
	*/

}
