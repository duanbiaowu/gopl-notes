# 第13章 低级编程

**unsafe包是由编译器实现的，提供了对语言内置特性的访问功能。这些特性一般不可见，因为它们暴露Go语言的内存布局。**

**任何类型的指针都可以通过强制转换为unsafe.Pointer指针类型去掉原有的类型信息，然后再重新赋予新的指针类型而达到指针间的转换的目的。**

#### 13.1 unsafe.Sizeof, Alignof 和 Offsetof

```go
import "unsafe"
fmt.Println(unsafe.Sizeof(float64(0))) // "8"
```

**Sizeof函数返回的大小只包括数据结构中固定的部分，例如字符串对应结构体中的指针和字符串长度部分，但是并不包含指针指向的字符串的内容。考虑到可移植性，引用类型或包含引用类型的大小在32位平台上是4个字节，在64位平台上是8个字节。**

**由于地址对齐因素，聚合类型（结构体或数组）的值的长度至少是它的成员或元素长度之和，并且由于 “内存间隙” 的存在，或许比这个更大一些。<code>内存空位</code> 是由编译器添加的未使用的内存地址，用来确保连续的成员或者元素相对于结构体或数组的起始地址是对齐的。**

| 类型                                       | 大小                              |
| ------------------------------------------ | --------------------------------- |
| <code>bool</code>                          | 1 个字节                          |
| <code>intN, uintN, floatN, complexN</code> | N / 8个字（例如float64是8个字节） |
| <code>int, uint, uintptr</code>            | 1 个字                            |
| <code>*T</code>                            | 1 个字                            |
| <code>string</code>                        | 2 个字 （数据、长度）             |
| <code>[]T</code>                           | 3 个字 （数据、长度、容量）       |
| <code>map</code>                           | 1 个字                            |
| <code>func</code>                          | 1 个字                            |
| <code>chan</code>                          | 1 个字                            |
| <code>interface</code>                     | 2 个字 （类型、值）               |

语言规范并没有要求成员声明的顺序对应内存中的布局顺序，所以理论上，编译器可以重新排列每个成员的内存位置。虽然在写作本书的时候编译器还没有这么做。

下面的三个结构体虽然有着相同的字段，但是第一种写法比另外的两个需要多50%的内存：

```go
                               // 64-bit  32-bit
struct{ bool; float64; int16 } // 3 words 4words
struct{ float64; int16; bool } // 2 words 3words
struct{ bool; int16; float64 } // 2 words 3words
```

译注：未来的Go语言编译器应该会默认优化结构体的顺序，当然应该也能够指定具体的内存布局，相同讨论请参考 [Issue10014](https://github.com/golang/go/issues/10014) ），内存使用率和性能都可能会受益。

`unsafe.Alignof` 函数返回对应参数的类型需要对齐的倍数

`unsafe.Offsetof` 函数计算成员 `f` 相对于结构体 `x` 起始地址的偏移值，如果有内存空位，也计算在内，该函数的操作数必须是一个成员选择器 `x.f`

```go
var x struct {
    a bool
    b int16
    c []int
}
```

<img src="https://books.studygolang.com/gopl-zh/images/ch13-01.png" alt="img" style="zoom:75%;" />

**32位系统：**

```go
Sizeof(x)   = 16  Alignof(x)   = 4
Sizeof(x.a) = 1   Alignof(x.a) = 1 Offsetof(x.a) = 0
Sizeof(x.b) = 2   Alignof(x.b) = 2 Offsetof(x.b) = 2
Sizeof(x.c) = 12  Alignof(x.c) = 4 Offsetof(x.c) = 4
```



**64位系统：**

```go
Sizeof(x)   = 32  Alignof(x)   = 8
Sizeof(x.a) = 1   Alignof(x.a) = 1 Offsetof(x.a) = 0
Sizeof(x.b) = 2   Alignof(x.b) = 2 Offsetof(x.b) = 2
Sizeof(x.c) = 24  Alignof(x.c) = 8 Offsetof(x.c) = 8
```

#### 13.2 unsafe.Pointer

**`unsafe.Pointer` 类型是一种特殊类型的指针，它可以存储任何变量的地址。**

我们无法间接地通过一个 `unsafe.Pointer` 变量来使用 `*p` ，因为我们不知道这个表达式的具体类型。和普通指针一样，`unsafe.Pointer` 指针是可以比较的并且可以和 `nil` 做比较，` nil ` 是指针类型的零值。

一个普通的`*T`类型指针可以被转化为unsafe.Pointer类型指针，并且一个unsafe.Pointer类型指针也可以被转回普通的指针，被转回普通的指针类型并不需要和原始的`*T`类型相同。通过将`*float64`类型指针转化为`*uint64`类型指针，我们可以查看一个浮点数变量的位模式。

```go
package math

func Float64bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }

fmt.Printf("%#016x\n", Float64bits(1.0)) // "0x3ff0000000000000"
```

**`unsafe.Pointer`** 类型也可以转换为 `uintptr` 类型，`uintptr` 保存了指针所指向地址的数值，这就可以让我们对地址进行数值计算。（`uintptr` 类型是一个足够大的无符号整数，可以用来表示任何地址。）转换可以逆向进行，但是从 `uintptr` 到 `unsafe.Pointer` 的转换会破坏类型系统，因为并不是所有的数值都是合法的内存地址。



**移动的垃圾回收器**

一些垃圾回收器在内存中会把变量移来移去，以减少内存碎片或是为了进行薄记工作。当一个变量在内存中移动后，该变量指向旧地址的所有指针都需要更新以指向新地址。从垃圾回收器的角度看，`unsafe.Pointer` 是一个变量的指针，当变量移动的时候，它的值也需要改变，而 `uintptr` 仅仅是一个数值，所以它的值是不会变的。

```go
var x struct {
    a bool
    b int16
    c []int
}

// 和 pb := &x.b 等价
pb := (*int16)(unsafe.Pointer(
    uintptr(unsafe.Pointer(&x)) + unsafe.Offsetof(x.b)))
*pb = 42
fmt.Println(x.b) // "42"
```

**不要尝试引入 `uintptr` 类型的临时变量来破坏执行代码** 

```go
// NOTE: subtly incorrect!

// 垃圾回收器无法通过非指针变量 tmp 了解它背后的指针
tmp := uintptr(unsafe.Pointer(&x)) + unsafe.Offsetof(x.b)
// 变量 x 可能在内存中已经移动了，这个时候 tmp 中的值就不是变量 &x.b 的地址了
pb := (*int16)(unsafe.Pointer(tmp))
// 将向一个任意的内存地址写入值 42
*pb = 42

// 类似原因导致的错误
// 例如这条语句：
pT := uintptr(unsafe.Pointer(new(T))) // 注意: 错误!
// 垃圾回收器将会在语句执行结束后回收内存，之后，pT 存储的是变量的旧地址，不过旧地址对应的变量已经发生改变
```

**Go 确实会在内存中移动变量，例如 goroutine 栈会根据需要增长。这个时候，旧栈上面的所有的变量都会重新分配到新的、更大的栈上面，所以我们不能期望变量的地址值在整个生命周期都不变。**

**建议遵守最小可用原则，可认为所有的 `uintptr` 值都包含一个变量的旧地址，并且减少 unsafe.Pointer 到 uintptr 之间的转换到使用这个 uinptr 之间的操作次数 (简单来说就是：尽量避免转换，如果转换的话，争取在一条语句中实现)。**
