# 第13章 低级编程

**unsafe包是由编译器实现的，提供了对语言内置特性的访问功能。这些特性一般不可见，因为它们暴露Go语言的内存布局。**

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

