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

**`refleact.Value` 和 interface{} 都可以包含任意的值。二者的区别是 interface{} 隐藏了值的布局信息、内置操作和相关方法。`reflect.Value` 有很多方法可以用来分析所包含的值，而不用知道它的类型。**

`reflect.Value` 的 `KInd` 方法来区分不同的类型。尽管有无限种类型，但类型的分类 (kind) 只有少数几种:

- 基础类型 Bool, String 以及各种数字类型
- 聚合类型 Array 和 Struct
- 引用类型 Chan, Func, Ptr, Slice, Map
- 接口类型 Interface
- Invalid类型 （空值，`reflect.Value` 的零值就属于 Invalid 类型）

#### 12.5 使用 reflect.Value 设置值

`reflect.Value` 中所有通过 `reflect.ValueOf(x)` 返回的 `reflect.Value` 都是不可取地址的。可以通过调用 `reflect.ValueOf(&x).Elem()` 来获得任意变量x可寻址的Value值。

```go
x := 2                   // value   type    variable?
// 不可取地址, 仅仅是整数2的拷贝副本
a := reflect.ValueOf(2)  // 2       int     no
// 不可取地址, 仅仅是x的拷贝副本
b := reflect.ValueOf(x)  // 2       int     no
// 不可取地址, 只是一个指针&x的拷贝副本
c := reflect.ValueOf(&x) // &x      *int    no
// 可取地址, c的解引用方式生成的，指向另一个变量
d := c.Elem()            // 2       int     yes (x)

// 可以通过调用 reflect.Value 的 CanAddr 方法来判断是否可寻址
fmt.Println(a.CanAddr()) // "false"
fmt.Println(b.CanAddr()) // "false"
fmt.Println(c.CanAddr()) // "false"
fmt.Println(d.CanAddr()) // "true"
```

可以通过一个指针来间接获取一个可寻址的 `reflect.Value` ，即时这个指针是不可寻址的。可寻址的常见规则都在反射包里边有对应项:

- slice 的索引表达式 e[i] 隐式地包含一个指针，即使 e 是不可寻址的，表达式仍然可寻址

- reflect.ValueOf(e).Index(i) 代表一个变量，即时 reflect.ValueOf(e) 是不可寻址的，表达式仍然可寻址

  

**从一个可寻址的 `reflect.Value` 获取变量需要三步：**

- 调用 Addr() ，返回一个 Value， 里面保存了指向变量的指针
- 在 Value 上调用 Interface()，返回一个 interaface{} ，里面包含指向变量的指针
- 可以使用类型断言把接口内容转换为一个普通指针，之后就可以通过普通指针更新变量了

```go
x := 2
d := reflect.ValueOf(&x).Elem()   // d refers to the variable x
px := d.Addr().Interface().(*int) // px := &x
*px = 3                           // x = 3
fmt.Println(x)                    // "3"
```

