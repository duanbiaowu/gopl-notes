# 第11章 测试

#### 11.1 go test

**`go test` 子命令是Go语言包的测试驱动程序， 这些包根据某些约定组织在一起。在一个包目录中，以 `_test.go` 结尾的文件不是 `go build` 命令编译的目标，而是 `go test` 编译的目标。**  

- 功能测试函数： 以 Test 前缀命名的函数，检测程序逻辑正确性，`go test` 运行这些测试函数
- 基准测试函数：以 Benchmark 前缀命名的函数，测试某些操作的性能，`go test` 汇报操作的平均执行时间
- 示例函数：以 Example 前缀命名的函数，提供一个由编译器保证正确性的示例文档

**`go test` 工具扫描 *_test.go 文件来寻找特殊函数，并生成一个临时的 `main` 包来调用它们。然后编译和运行，并汇报结果，最后清空临时文件。**  



#### 11.2 Test 函数

**功能测试函数必须以 `Test` 开头，可选的后缀名必须以大写字母开头**

- 参数 `-v` 可用于输出每个测试函数的名字和运行时间
- 参数 `-run` 对应一个正则表达式，只有测试函数名被它正确匹配的测试函数才会被`go test`测试命令运行

```go
// t参数用于报告测试失败和附加的日志信息

func TestSin(t *testing.T) { /* ... */ }
func TestCos(t *testing.T) { /* ... */ }
func TestLog(t *testing.T) { /* ... */ }
```

**比较好的实践是先写测试用例然后发现它触发的错误和用户 bug 报告里面的一致，只有这样，我们才能确信我们要修复的内容是针对这个出现的问题。**

**一旦我们已经修复了失败的测试用例，在我们提交代码更新之前，我们应该以不带参数的`go test`命令运行全部的测试用例，以确保修复失败测试的同时没有引入新的问题。**

```go
// 表格驱动测试
func TestIsPalindrome(t *testing.T) {
    var tests = []struct {
        input string
        want  bool
    }{
      	// 根据需要添加新的测试数据
        {"", true},
        {"a", true},
        {"aa", true},
        {"ab", false},
        {"kayak", true},
        {"detartrated", true},
        {"A man, a plan, a canal: Panama", true},
        {"Evil I did dwell; lewd did I live.", true},
        {"Able was I ere I saw Elba", true},
        {"été", true},
        {"Et se resservir, ivresse reste.", true},
        {"palindrome", false}, // non-palindrome
        {"desserts", false},   // semi-palindrome
    }
    for _, test := range tests {
        if got := IsPalindrome(test.input); got != test.want {
          	// t.Errof 输出的失败的测试用例信息没有包含整个跟踪栈信息
          	// 也不会导致程序宕机或者终止执行
          	// 测试用例彼此是独立的，如果测试表中的一个测试用例失败，其他的用例仍然继续测试
          	// 这样就可以在一次测试中发现多个失败的情况
            t.Errorf("IsPalindrome(%q) = %v", test.input, got)
          
          	// 如果真的需要终止一个测试函数，使用 t.Fatal 或者 t.Fatalf
          	// 这些函数的调用必须和测试函数在同一个 goroutine 内调用
          	// t.Fatalf("IsPalindrome(%q) = %v", test.input, got)
        }
    }
}
```

##### 11.2.1  随机测试

**通过构建更广泛的随机输入来扩展测试的覆盖范围**

如果输入是随机的，怎么知道函数输出什么内容呢？

- 额外写一个函数，使用低效但是清晰的算法（也可以是语言内置方法），然后检测两种实现输出是否一致
- 构建符合某种模式的输入，就可以知道对应的输出是什么

```go
// 随机生成回文字符串
func randomPalindrome(rng *rand.Rand) string {
    n := rng.Intn(25) // random length up to 24
    runes := make([]rune, n)
    for i := 0; i < (n+1)/2; i++ {
        r := rune(rng.Intn(0x1000)) // random rune up to '\u0999'
        runes[i] = r
        runes[n-1-i] = r
    }
    return string(runes)
}
```



##### 11.2.2 测试命令

**在测试的代码里面不要掉一哦那个 `log.Fatal ` 或者 `os.Exit`， 因为这两个调用会阻止跟踪的过程，这两个函数的调用可以认为是 `main 函数的特权` **

##### 11.2.3 白盒测试

```go
// 全局变量模拟Mock
var notifiedUser, notifiedMsg string
notifyUser = func(user, msg string) {
  notifiedUser, notifiedMsg = user, msg
}

// ...simulate a 980MB-used condition...
const user = "joe@example.org"
CheckQuota(user)

// 当测试函数返回后， CheckQuota 因为仍然使用该测试的伪通知实现 notifyUser
// 所以再次在其他测试中调用它时就不能正常工作了
// 必须修改测试代码恢复 notifyUser 原来的值，这样后面的测试才不会受到影响

// Save and restore original notifyUser.
saved := notifyUser
defer func() { notifyUser = saved }()

// 这种方式也可以用来临时保存并恢复各种全局变量，包括命令行选项、调试参数，以及性能参数；
// 也可以用来安装和移除钩子程序来让产品代码调用测试代码；
// 或者将产品代码设置为少见却很重要的状态，比如超时、错误，甚至是刻意制造的交叉并行执行
// 以这种方式使用全局变量是安全的，go test命令一般不会同时并发地执行多个测试
```

##### 11.2.4 外部测试包

**解决循环引用： 将测试函数定义在外部测试包中。**

将文件的包声明为 ```xxx_test```，这个额外的 `后缀 _test` 告诉 `go test` 工具，它应该单独地编译一个包，这个包仅包含这些文件，然后运行它的测试。

由于外部测试在一个单独的包里面，因此它们也可以引用一些依赖于被测试包的帮助包，这个是包内是无法做到的，从设计层次来看，外部测试包逻辑上在它所依赖的两个包之上。

![img](https://books.studygolang.com/gopl-zh/images/ch11-02.png)

为了避免包循环导入， 外部测试包允许测试用例，尤其是集成测试用例（用来测试多个组件的交互），自由地导入其他的包，就像一个应用程序那样。

`go list` 工具来汇总一个包目录中哪些是产品代码，哪些是包内测试以及哪些是外部测试，

```go
// GoFiles表示产品代码对应的Go源文件列表；也就是go build命令要编译的部分
go list -f={{.GoFiles}} fmt

// TestGoFiles表示的是包内部测试代码，以_test.go为后缀文件名，不过只在测试时被构建
go list -f={{.TestGoFiles}} fmt

// XTestGoFiles表示的是属于外部测试包的测试代码，因此它们必须先导入fmt包
// 同样，这些文件只在测试时被构建运行
go list -f={{.XTestGoFiles}} fmt

// 有时候外部测试包也需要访问被测试包内部的代码
// 例如在一个为了避免循环导入而被独立到外部测试包的白盒测试
// 在包内的一个_test.go文件中导出一个内部的实现给外部测试包
// 因为这些代码只有在测试时才需要，因此一般会放在export_test.go文件中

// fmt包的fmt.Scanf函数需要unicode.IsSpace函数提供的功能
// 为了避免创建不合理的依赖，fmt没有导入unicode包及其巨大的数据表
// 而是包含了一个更加简单的实现 isSpace

// /fmt/scan.go
package fmt

func isSpace(r rune) bool {
	if r >= 1<<16 {
		return false
	}
	rx := uint16(r)
	for _, rng := range space {
		if rx < rng[0] {
			return false
		}
		if rx <= rng[1] {
			return true
		}
	}
	return false
}

// /fmt/export_test.go
// 这个文件没有定义测试；它仅定义了一个导出符号 fmt.IsSpace 用来给外部测试使用
// 任何外部测试需要使用白盒测试技术时都可以使用这个技巧
package fmt

var IsSpace = isSpace
var Parsenum = parsenum

// /fmt/fmt_test.go
package fmt_test

func TestIsSpace(t *testing.T) {
	// This tests the internal isSpace function.
	// IsSpace = isSpace is defined in export_test.go.
	for i := rune(0); i <= unicode.MaxRune; i++ {
		if IsSpace(i) != unicode.IsSpace(i) {
			t.Errorf("isSpace(%U) = %v, want %v", i, IsSpace(i), unicode.IsSpace(i))
		}
	}
}
```

##### 11.2.5 编写有效的测试

一个好的测试不会在发生错误时崩溃，而是输出该问题的一个简洁、清晰的现象描述，以及其他与上下文相关的信息。理想情况下，维护者不需要再通过阅读源代码来探究测试失败的原因。一个好的测试不应该在发生失败后就终止，而是要在一次运行中尝试报告多个错误，因为错误发生的方式本身会揭露错误的原因。

一个好测试的关键是首先实现你所期望的具体行为，并且仅在这个时候再使用工具函数来使代码简洁并且避免重复。好的结果很少是从抽象、通用的测试函数开始的。

##### 11.2.6 避免脆弱的测试

避免脆弱测试代码的方法是只检测你真正关心的属性。保持测试代码的简洁和内部结构的稳定。特别是对断言部分要有所选择。不要对字符串进行全字匹配，而是针对那些在项目的发展中是比较稳定不变的子串。很多时候值得花力气来编写一个从复杂输出中提取用于断言的必要信息的函数，虽然这可能会带来很多前期的工作，但是它可以帮助迅速及时修复因为项目演化而导致的不合逻辑的失败测试。

#### 11.3 覆盖率

**测试旨在发现bug，而不是证明其不存在。**

```go
// 显示测试覆盖率工具的使用用法
// got tool 运行 Go 工具链里面的一个可执行文件
go tool cover

// -coverprofile 标志参数运行测试，通过检测产品代码，启用了覆盖数据收集
// 它修改了源代码的副本，这样在每个语句块执行之前，设置一个布尔变量，每个语句快都对应一个变量
// 在修改的程序退出之前，它将每个变量的值都写入到执行的 c.out 文件并且输出被执行语句的汇总信息
// 如果只需要汇总信息，那么使用 test -cover
go test -run=Coverage -coverprofile=c.out gopl.io/ch7/eval

// -covermode=count 每个语句块的检测会递增一个计算器而不是设置布尔量
// 统计结果中记录了每个块的执行次数，可以识别出被频繁执行的 “热点代码”

// 生成数据之后，运行 cover 工具，来处理生成的日志，生成一个 HTML 报告
// 界面中，绿色标记的语句块表示它被覆盖了，红色的则表示它没有被覆盖
go tool cover -html=c.out
```

![img](https://books.studygolang.com/gopl-zh/images/ch11-03.png)



**100% 的测试覆盖率听起来很美，但是在具体实践中通常是不可行的，也不是值得推荐的做法。因为那只能说明代码被执行过而已，并不意味着代码就是没有BUG的；因为对于逻辑复杂的语句需要针对不同的输入执行多次。**

测试本质上是实用主义行为，在编写测试的代价和可以通过测试避免的错误造成的代价之间进行平衡。测试覆盖率工具可以帮助我们快速识别测试薄弱的地方，但是设计好的测试用例和编写应用代码一样需要严密的思考。

#### 11.4 Benchmark函数

```go
// 准测试函数和普通测试函数写法类似，但是以Benchmark为前缀名
// *testing.B参数除了提供和*testing.T类似的方法，还有额外一些和性能测量相关的方法 
// 它还提供了一个整数N，用于指定操作执行的循环次数
func BenchmarkIsPalindrome(b *testing.B) {
    for i := 0; i < b.N; i++ {
        IsPalindrome("A man, a plan, a canal: Panama")
    }
}

// 普通测试不同的是，默认情况下不运行任何基准测试
// 通过 -bench 命令行标志参数手工指定要运行的基准测试函数
// 结果中基准测试名的数字后缀部分，这里是8
// 表示运行时对应的GOMAXPROCS的值，这对于一些与并发相关的基准测试是重要的信息
go test -bench=.
PASS
BenchmarkIsPalindrome-8 1000000                1035 ns/op
ok      gopl.io/ch11/word2      2.179s
```

**基准测试运行器开始的时候并不清楚这个操作的耗时长短，它会尝试在真正运行基准测试前先尝试用较小的N运行测试来估算基准测试函数所需要的时间，然后推断一个较大的时间保证稳定的测量结果。**

**使用基准测试函数来实现循环而不是在基准测试框架内实现的原因是，在基准测试函数中可以在循环外面可以执行一些必要的初始化代码，并且这段时间不加到每次迭代的时间中。如果初始化代码干扰了结果，参数 `testing.B` 提供了方法用来停止、恢复和重置计时器，但是这些方法很少用到。**

```go
// -benchmem 命令行标志参数将在报告中包含内存的分配数据统计
go test -bench=. -benchmem
PASS
BenchmarkIsPalindrome    1000000   1026 ns/op    304 B/op  4 allocs/op

// 优化之后
// 在一次 make 调用中分配全部所需的内存减少了 75% 的分配次数并且减少了一半的内存使用
go test -bench=. -benchmem
PASS
BenchmarkIsPalindrome    2000000    807 ns/op    128 B/op  1 allocs/op

// 这种基准测试告诉我们给定操作的绝对耗时
// 但是在大多数情况下，引起关注的性能问题是两个不同操作之间的相对耗时
// 基准测试可以帮助我们在性能达标情况下选择出所需的最小内存
// 基准测试可以评估两种不同算法对于相同的输入在不同的场景和负载下的优缺点

// 性能比较函数只是普通的代码，表现形式通常是带有一个参数的函数
func benchmark(b *testing.B, size int) { /* ... */ }
func Benchmark10(b *testing.B)         { benchmark(b, 10) }
func Benchmark100(b *testing.B)        { benchmark(b, 100) }
func Benchmark1000(b *testing.B)       { benchmark(b, 1000) }
```

#### 11.5 性能剖析

Go语言支持多种类型的剖析性能分析，每一种关注不同的方面，但它们都涉及到每个采样记录感兴趣的一系列事件消息，每个事件都包含函数调用时函数调用堆栈的信息。内建的`go test`工具对几种分析方式都提供了支持。

- CPU剖析数据标识了最耗CPU时间的函数
- 堆剖析标识了最耗内存的语句
- 阻塞剖析则记录阻塞goroutine最久的操作，例如系统调用、管道发送和接收，还有获取锁等

```go
// 当一次使用多个标记的时候需要注意
// 获取性能分析报告的机制是获取其中一个类别的包时会覆盖掉其他类别的报告
go test -cpuprofile=cpu.out
go test -blockprofile=block.out
go test -memprofile=mem.out
```

获取性能剖析结果后，需要通过 `pprof` 工具来分析。这是Go发布包的标准部分，但是因为不经常使用，所以通过 `go tool pprof` 间接使用。它有很多特性和选项，但是基本的用法只有两个参数，产生性能剖析结果的可执行文件和性能剖析日志。

```go
// 通常最好是对业务关键代码的部分设计专门的基准测试，因为简单的基准测试几乎没法代表业务场景
// 基准测试永远没有代表性，因此我们用-run=NONE参数禁止
go test -run=NONE -bench=ClientServerParallelTLS64 \
    -cpuprofile=cpu.log net/http

PASS
 BenchmarkClientServerParallelTLS64-8  1000
    3141325 ns/op  143010 B/op  1747 allocs/op
ok       net/http       3.395s

// -text 参数用于指定输出格式
// -nodecount=10 参数限制了只输出前10行的结果
go tool pprof -text -nodecount=10 ./http.test cpu.log

// crypto/elliptic.p256ReduceDegree函数占用了将近一半的CPU资源
// 相比之下，如果一个概要文件中主要是runtime包的内存分配的函数
// 那么减少内存消耗可能是一个值得尝试的优化策略
2570ms of 3590ms total (71.59%)
Dropped 129 nodes (cum <= 17.95ms)
Showing top 10 nodes out of 166 (cum >= 60ms)
    flat  flat%   sum%     cum   cum%
  1730ms 48.19% 48.19%  1750ms 48.75%  crypto/elliptic.p256ReduceDegree
   230ms  6.41% 54.60%   250ms  6.96%  crypto/elliptic.p256Diff
   120ms  3.34% 57.94%   120ms  3.34%  math/big.addMulVVW
   110ms  3.06% 61.00%   110ms  3.06%  syscall.Syscall
    90ms  2.51% 63.51%  1130ms 31.48%  crypto/elliptic.p256Square
    70ms  1.95% 65.46%   120ms  3.34%  runtime.scanobject
    60ms  1.67% 67.13%   830ms 23.12%  crypto/elliptic.p256Mul
    60ms  1.67% 68.80%   190ms  5.29%  math/big.nat.montgomery
    50ms  1.39% 70.19%    50ms  1.39%  crypto/elliptic.p256ReduceCarry
    50ms  1.39% 71.59%    60ms  1.67%  crypto/elliptic.p256Sum
```

对于一些更微妙的问题，可能需要使用pprof的图形显示功能。这个需要安装GraphViz工具，可以从 [http://www.graphviz.org](http://www.graphviz.org/) 下载。参数`-web`用于生成函数的有向图，标注有CPU的使用和最热点的函数等信息。

想了解更多，可以阅读Go官方博客的“ [Profiling Go Programs](https://blog.golang.org/pprof) ”

#### 11.6 Example函数

示例函数，以 Example为函数名开头 ，没有函数参数和返回值。

```go
func ExampleIsPalindrome() {
    fmt.Println(IsPalindrome("A man, a plan, a canal: Panama"))
    fmt.Println(IsPalindrome("palindrome"))
    // Output:
    // true
    // false
}
```

**示例函数有三个用处**

- 最主要的一个是作为文档
  - 一个包的例子可以更简洁直观的方式来演示函数的用法，比文字描述更直接易懂，特别是作为一个提醒或快速参考。
- 可以通过`go test` 运行可执行测试
  - 如果示例函数内含有类似上面例子中的`// Output:`格式的注释，那么测试工具会执行这个示例函数，然后检查示例函数的标准输出与注释是否匹配。
- 提供手动实验代码
  -  [http://golang.org](http://golang.org/) 就是由godoc提供的文档服务，它使用了Go Playground让用户可以在浏览器中在线编辑和运行每个示例函数。

