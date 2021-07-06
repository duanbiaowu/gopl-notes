# 第10章 包和工具

## 10.1 引言

```go
// 查看标准包的具体数目
go list std | wc -l
```

每个包一般都定义了一个不同的名字空间用于它内部的每个标识符的访问。每个名字空间关联到一个特定的包，让我们给类型、函数等选择简短明了的名字，这样可以在使用它们的时候减少和其它部分名字的冲突。

每个包还通过控制包内名字的可见性和是否导出来实现封装特性。通过限制包成员的可见性并隐藏包API的具体实现，将允许包的维护者在不影响外部包用户的前提下调整包的内部实现。通过限制包内变量的可见性，还可以强制用户通过某些特定函数来访问和更新内部变量，这样可以保证内部变量的一致性和并发时的互斥约束。

**当我们修改了一个源文件，我们必须重新编译该源文件对应的包和所有依赖该包的其他包。**

Go语言编译速度明显快于其它编译语言：

- 所有导入的包必须在每个文件的开头显式声明
  - 编译器就没有必要读取和分析整个源文件来判断包的依赖关系
- 禁止包的环状依赖
  - 因为没有循环依赖，包的依赖关系形成一个有向无环图，每个包可以被独立编译，而且很可能是被并发编译
- 编译后包的目标文件不仅仅记录包本身的导出信息，目标文件同时还记录了包的依赖关系
  - 因此，在编译一个包的时候，编译器只需要读取每个直接导入包的目标文件，而不需要遍历所有依赖的的文件（译注：很多都是重复的间接依赖）

## 10.2 导入路径

**每个包是由一个全局唯一的字符串进行标识，称为导入路径。出现在import语句中的导入路径也是字符串。**

```go
// 为了避免冲突，所有非标准库包的导入路径建议以所在组织的互联网域名为前缀
// 而且这样也有利于包的检索
import (
    "fmt"
    "math/rand"
    "encoding/json"

    "golang.org/x/net/html"

    "github.com/go-sql-driver/mysql"
)
```

## 10.3 包声明

**在每个Go语言源文件的开头都必须有包声明语句。包声明语句的主要目的是确定当前包被其它包导入时默认的标识符（也称为包名）。**

```go
// 默认的包名就是包导入路径名的最后一段
// 即使两个包的导入路径不同，它们依然可能有一个相同的包名
// 两个包名都是 rand
import (
    // "crypto/rand"
    // "math/rand"
)
```

默认包名一般采用导入路径名的最后一段作为约定， 但是有三种例外情况：

- 包对应一个可执行程序，也就是main包

  - main包本身的导入路径是无关紧要的，名字为`main` 的包是给 `go build` 构建命令一个信息，这个包编译完之后必须调用连接器生成一个可执行程序

- 包所在的目录中可能有一些文件名是以`_test.go`为后缀的Go源文件，并且这些源文件声明的包名也是以`_test`为后缀名的

  - 文件名称前面必须有其它的字符，因为以`_`或`.`开头的源文件会被构建工具忽略
  - 目录可以包含两种包：一种是普通包，另一种则是测试的外部扩展包
  - `_test` 后缀告诉 `go test` 两个包都需要构建，并且指明文件属于哪个包

- 一些依赖版本号的管理工具会在导入路径后追加版本号信息

  ```go
  // 包的名字并不包含版本号后缀，而是yaml
  gopkg.in/yaml.v2
  ```

## 10.4 导入声明

导入的包之间可以通过添加空行来分组；通常将来自不同组织的包独自分组。包的导入顺序无关紧要，但是在每个分组中一般会根据字符串顺序排列。

```go
// gofmt和goimports工具都可以自动分组并排序
import (
    "fmt"
    "html/template"
    "os"

    "golang.org/x/net/html"
    "golang.org/x/net/ipv4"
)

// 重命名导入
// 除了命名冲突外，如果导入的包名很笨重，重命名一个简短名称会更方便
import (
    "crypto/rand"
    mrand "math/rand" // alternative name mrand avoids conflict
)
```

**每个导入声明从当前包向导入的包建立一个依赖，如果这些依赖形成一个循环， `go build` 会报错。** 

## 10.5 包的匿名导入

**如果导入的包的名字没有在文件中引用，就会产生一个编译错误。但是有时候导入一个包仅仅是为了利用其副作用：对包级别的变量执行初始化表示式求值，并执行它的init函数。这时需要抑制 `“unused import”` 编译错误，可以用下划线`_`来重命名导入的包。下划线`_`为空白标识符，并不能被访问，这称为 `空白导入`。**

## 10.6 包及其命名

包名一般采用单数的形式。标准库的bytes、errors和strings使用了复数形式，这是为了避免和预定义的类型冲突，同样还有go/types是为了避免和type关键字冲突。

```go
// 当设计一个包的时候，需要考虑包名和成员名两个部分如何很好地配合
bytes.Equal    flag.Int    http.Get    json.Marshal

package strings

func Index(needle, haystack string) int

type Replacer struct{ /* ... */ }
func NewReplacer(oldnew ...string) *Replacer

type Reader struct{ /* ... */ }
func NewReader(s string) *Reader
```

## 10.7 go工具

```go
// go或go help命令查看内置的帮助文档
go
...
    build            compile packages and dependencies
    clean            remove object files
    doc              show documentation for package or symbol
    env              print Go environment information
    fmt              run gofmt on package sources
...
```

### 10.7.1 工作空间的组织

```go
// 对于大多数的Go语言用户，只需要配置一个名叫GOPATH的环境变量，用来指定当前工作目录即可
// 当需要切换到不同工作区的时候，只要更新GOPATH
export GOPATH=$HOME/gobook
go get gopl.io/...
```

**GOPATH对应的工作区目录有三个子目录：**

- src子目录用于存储源代码
  - 每个包被保存在与$GOPATH/src的相对路径为包导入路径的子目录中
- pkg子目录用于保存编译后的包的目标文件
- bin子目录用于保存编译后的可执行程序

**GOROOT用来指定Go的安装目录，还有它自带的标准库包的位置。**

GOROOT的目录结构和GOPATH类似，因此存放fmt包的源代码对应目录应该为$GOROOT/src/fmt。用户一般不需要设置GOROOT，默认情况下Go语言安装工具会将其设置为安装的目录路径。

**`go env` 命令输出与工具链相关的已经设置有效值的环境变量及值，还会输出未设置有效值的环境变量及其默认值。**

```go
go env 
// GOOS环境变量用于指定目标操作系统
// GOARCH环境变量用于指定处理器的类型
...
```

### 10.7.2 包的下载

```go
// 获取 golint 工具 
go get github.com/golang/lint/golint
// 用golint命令对 gopl.io/ch2/popcount包代码进行编码风格检查
$GOPATH/bin/golint gopl.io/ch2/popcount
// go get命令支持当前流行的托管网站GitHub、Bitbucket和Launchpad
// go help importpath获取相关的信息
```

`go get -u`命令只是简单地保证每个包是最新版本，如果是第一次下载包则是比较方便的；但是对于发布程序则可能是不合适的，因为本地程序可能需要对依赖的包做精确的版本依赖管理。通常的解决方案是使用vendor的目录用于存储依赖包的固定版本的源代码，对本地依赖的包的版本更新也是谨慎和持续可控的。

### 10.7.3 包的构建

**`go build`命令编译命令行参数指定的每个包。如果包是一个库，则忽略输出结果；这可以用于检测包是可以正确编译的。如果包的名字是main，`go build`将调用链接器在当前目录创建一个可执行程序；以导入路径的最后一段作为可执行程序的名字。**

```go
// 每个包可以由它们的导入路径指定
// 或者用一个相对目录的路径名指定，相对路径必须以.或..开头

1. 
cd $GOPATH/src/gopl.io/ch1/helloworld
go build

2. 
cd anywhere
go build gopl.io/ch1/helloworld

3.
cd $GOPATH
go build ./src/gopl.io/ch1/helloworld

// 也可以指定包的源文件列表，这一般只用于构建一些小程序或做一些临时性的实验
// 如果是main包，将会以第一个Go源文件的基础文件名作为最终的可执行程序的名字
cat quoteargs.go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Printf("%q\n", os.Args[1:])
}
$ go build quoteargs.go
$ ./quoteargs one "two three" four\ five
["one" "two three" "four five"]
```

**`go run`命令实际上是结合了构建和运行的两个步骤：**

```go
go run quoteargs.go one "two three" four\ five
["one" "two three" "four five"]
```

**默认情况下，`go build`命令构建指定的包和它依赖的包，然后丢弃除了最后的可执行文件之外所有的中间编译结果。随着项目包数量和代码行数的增加，编译时间将变得可观，即使依赖项没有改变。**

**`go install`命令和`go build`命令很相似，但是它会保存每个包的编译成果，而不是将它们都丢弃。被编译的包会被保存到$GOPATH/pkg目录下，目录路径和 src目录路径对应，可执行程序被保存到$GOPATH/bin目录。（很多用户会将$GOPATH/bin添加到可执行程序的搜索列表中。）**

**`go install`命令和`go build`命令都不会重新编译没有发生变化的包，这可以使后续构建更快捷。为了方便编译依赖的包，`go build -i`命令将安装每个目标所依赖的包。因为编译对应不同的操作系统平台和CPU架构，`go install`命令会将编译结果安装到 `GOOS 和 GOARCH 对应的目录`。**

```go
// 针对不同操作系统或CPU的交叉构建
func main() {
    fmt.Println(runtime.GOOS, runtime.GOARCH)
}

go build gopl.io/ch10/cross
./cross
darwin amd64

GOARCH=386 go build gopl.io/ch10/cross
./cross
darwin 386

// 构建标签的特殊注释，提供更细粒度的控制
// 只有编译程序对应的目标操作系统是Linux或Mac OS X时才编译这个文件
// +build linux darwin
// 不编译这个文件
// +build ignore

// 更多细节参考：
go doc go/build
```

### 10.7.4 包的文档化

Go语言中的文档注释一般是完整的句子，第一行通常是摘要说明，以被注释者的名字开头。注释中函数的参数或其它的标识符并不需要额外的引号或其它标记注明。

**`go doc` 命令打印其后所指定的实体的声明与文档注释**

```go
// 一个包
go doc time
package time // import "time"
...

// 一个包成员
go doc time.Since
func Since(t Time) Duration
...

// 一个方法
go doc time.Duration.Seconds
func (d Duration) Seconds() float64
...

// go doc 并不需要输入完整的包导入路径或正确的大小写
go doc json.decode
func (dec *Decoder) Decode(v interface{}) error
...
```

**`godoc` 提供相互链接的HTML页面服务，包含和`go doc`命令相同以及更多的信息。**

```go
// 在自己的工作区目录运行godoc服务
// -analysis=type			// 打开文档中关于静态分析的结果
// -analysis=pointer	// 打开代码中关于静态分析的结果
godoc -http :8000		//  http://localhost:8000/pkg
```

### 10.7.5 内部包

**`go build` 工具对导入路径中包含路径片段 `internal` 做了特殊处理。这种包叫internal包，一个internal包只能被和internal目录有同一个父目录的包所导入。**

```go
// net/http/internal/chunked 内部包只能被 net/http/httputil 或 net/http包导入
// 不能被net/url包导入
net/http
net/http/internal/chunked
net/http/httputil
net/url
```

### 10.7.6 包的查询

**`go list`命令可以查询可用包的信息。**

```go
// 测试包是否在工作区并打印它的导入路径：
go list github.com/go-sql-driver/mysql
github.com/go-sql-driver/mysql

// "..."表示匹配任意的包的导入路径
go list ...
archive/tar
archive/zip
...

// 或者是特定子目录下的所有包：
go list gopl.io/ch3/...
gopl.io/ch3/basename1
gopl.io/ch3/basename2
...

// 或者是和某个主题相关的所有包:
go list ...xml...
encoding/xml
gopl.io/ch7/xmlselect
...

// -json命令行参数表示用JSON格式打印每个包的元信息
go list -json hash
{
    "Dir": "/home/gopher/go/src/hash",
    "ImportPath": "hash",
    "Name": "hash",
  ......
```

