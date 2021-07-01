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

平常由编译器来检查的那些可赋值性条件，在这种情况下则是在运行时由 Set 方法来检查。上面的变量和值都是 int 类型，但如果变量类型是 int64， 这个程序就会崩溃，所以确保这个值对于变量类型是可赋值的。

d.Set(reflect.ValueOf(int64(5))) // panic: int64 is not assignable to int

当然，在一个不可寻址的 reflect.Value 上调用 Set 方法也会崩溃
x := 2
b := reflect.ValueOf(x)
b.Set(reflect.ValueOf(3)) // panic: Set using unaddressable value
```

还有很多用于基本类型的 Set 方法，SetInt、SetUint、SetString和SetFloat等，这些方法有一定程度的容错性，只要变量类型是某种带符号的整数，比如 SetInt， 甚至可以是底层类型为带符号整数的命名类型，都可以成功。如果值太大了还会无提示地截断，但需要注意的是，在指向 `interface{}` 变量的 `reflect.Value` 上调用 SetInt会崩溃，尽管使用 Set 就没有问题。

```go
x := 1
rx := reflect.ValueOf(&x).Elem()
rx.SetInt(2)                     // OK, x = 2
rx.Set(reflect.ValueOf(3))       // OK, x = 3
rx.SetString("hello")            // panic: string is not assignable to int
rx.Set(reflect.ValueOf("hello")) // panic: string is not assignable to int

var y interface{}
ry := reflect.ValueOf(&y).Elem()
ry.SetInt(2)                     // panic: SetInt called on interface Value
ry.Set(reflect.ValueOf(3))       // OK, y = int(3)
ry.SetString("hello")            // panic: SetString called on interface Value
ry.Set(reflect.ValueOf("hello")) // OK, y = "hello"
```



**一个可寻址的 `refletct.Value` 会记录它是否通过遍历一个未导出字段来获得的，如果是这样，则不允许修改。因此，CanAddr方法并不能正确反映一个变量是否是可以被修改的。另一个相关的方法CanSet是用于检查对应的 `reflect.Value` 是否是可取地址并可被修改的**

```go
fmt.Println(fd.CanAddr(), fd.CanSet()) // "true false"
```



#### 12.9 注意事项

反射是一个强大并富有表达力的工具，但是它应该被小心地使用，原因有三：

- 基于反射的代码是比较脆弱的。对于每一个会导致编译器报告类型错误的问题，在反射中都有与之相对应的误用问题，不同的是编译器会在构建时马上报告错误，而反射则是在真正运行到的时候才会抛出panic异常
  - 反射同样降低了程序的安全性，还影响了自动化重构和分析工具的准确性，因为它们无法识别运行时才能确认的类型信息
- 即使对应类型提供了相同文档，但是反射的操作不能做静态类型检查，而且大量反射的代码通常难以理解。总是需要小心翼翼地为每个导出的类型和其它接受interface{}或reflect.Value类型参数的函数维护说明文档
- 基于反射的代码通常比正常的代码运行速度慢一到两个数量级。对于一个典型的项目，大部分函数的性能和程序的整体性能关系不大，所以当反射能使程序更加清晰的时候可以考虑使用。测试是一个特别适合使用反射的场景，因为每个测试的数据集都很小。但是对于性能关键路径的函数，最好避免使用反射
