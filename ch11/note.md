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

##### 12.2.1  随机测试

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



##### 12.2.2 测试命令

**在测试的代码里面不要掉一哦那个 `log.Fatal ` 或者 `os.Exit`， 因为这两个调用会阻止跟踪的过程，这两个函数的调用可以认为是 `main 函数的特权` **

