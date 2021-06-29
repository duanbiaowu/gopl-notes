# 第12章 反射

#### 12.1 为什么使用反射

当我们无法透视一个未知类型的布局时，代码就无法继续，这时就需要使用反射了。

#### 12.2 refelct.Type 和 reflect.Value

反射功能由 `reflect` 包提供，它定义了两个重要的类型：`Type` 和 `Value`。

`Type` 表示Go语言的一个类型，它是一个有很多方法的接口，这些方法可以用来识别类型以及透视类型的组成部分，比如一个结构的各个字段或者一个函数的各个参数。

`reflect.Type` 接口只有一个实现，即类型描述符，接口值中的动态类型也是类型描述符号。

**`reflect.TypeOf ` 接受任意的 interface{} 类型，并以 reflect.Type 形式返回其动态类型：**

```go
// 把数值3赋值给interface{}参数
t := reflect.TypeOf(3)  // reflect.Type
fmt.Println(t.String()) // "int"
fmt.Println(t)          // "int"
```

**把一个具体值赋给一个接口类型时会发生一个隐式类型转换，转换会生成一个包含两部分内容的接口值：动态类型部分是操作数的类型(int)，动态值部分是操作数的值(3)**

因为 `reflect.TypeOf` 返回一个接口值对应的动态类型，所以它返回总是具体类型（而不是接口类型）。

```go
var w io.Writer = os.Stdout
fmt.Println(reflect.TypeOf(w)) // "*os.File"  not "io.Writer"

// 简写方式 %T
// 内部实现使用了 reflect.Typeof
fmt.Printf("%T\n", 3) // "int"
```



`reflect.Value` 可以包含一个任意类型的值。

`reflect.ValueOf` 函数接受任意的 interface{} 并将接口的动态值以 `reflect.Value` 的形式返回。与 `reflect.TypeOf` 类型，`reflect.ValueOf` 的返回值也都是具体值，不过 `reflect.Value` 也可以包含一个接口值。

```go
v := reflect.ValueOf(3) // reflect.Value
fmt.Println(v)          // "3"
// %v 对 reflect.Value 会进行特殊处理
fmt.Printf("%v\n", v)   // "3"
fmt.Println(v.String()) // NOTE: "<int Value>"
```

`reflect.ValueOf` 的逆操作是 `reflect.Value.Interface` 方法。它返回一个 interface{} 类型，装载着与 reflect.Value 相同的具体值：

```go
v := reflect.ValueOf(3) // reflect.Value
x := v.Interface()      // interface{}
i := x.(int)            // int
fmt.Printf("%d\n", i)   // "3"
```

