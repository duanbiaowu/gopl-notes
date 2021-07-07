# 第9章 使用共享变量实现并发

## 9.1 竞态

**如果一个函数在并发调用是仍然能正确运行，那么这个函数是兵法安全的。**

**并发安全是特例而不是普遍存在的，所以只有在文档指出类型是安全的情况下，才可以并发地访问一个变量。对于绝大多数变量，如果要避免并发访问，要么限制变量只存在一个 `goroutine `内，要么维护一个更高层的`互斥不变量`。**

**包级别的导出函数一般情况下都是并发安全的。由于package级的变量没法被限制在单一的 `gorouine`，所以修改这些变量“必须”使用互斥条件。**

**并发调用无法工作的原因：**

- 死锁（deadlock）
- 活锁（livelock）：比如多个线程在尝试绕开死锁，却由于过分同步导致反复冲突
- 饿死（resource starvation）

**竞争条件指的是程序在多个 goroutine 交叉执行操作时，没有给出正确的结果。**

**数据竞态发生于两个 goroutine 并发读写同一个变量并且至少其中一个是写入操作。**

```go
// 最后一个语句中的x的值是未定义的
// 可能是nil，也可能是一个长度为10的slice，也可能是一个长度为1,000,000的slice

// 如果指针是从第一个make调用来，而长度从第二个make来，x就变成了一个混合体
// 一个长度为1,000,000 但实际上内部只有10个元素的slice
// 存储 999999 元素的位置会伤及一个遥远的内存位置，后果无法预测，问题难以调试和定位
// 这种语义雷区被称为 "未定义行为"
var x []int
go func() { x = make([]int, 10) }()
go func() { x = make([]int, 1000000) }()
x[999999] = 1 // NOTE: undefined behavior; memory corruption possible!
```

**一个好的经验法则是根本就没有什么所谓的良性数据竞争，三种方法来避免：**

- 第一种方法：不要修改变量

  ```go
  // 并发调用时，map就会存在数据竞争
  var icons = make(map[string]image.Image)
  func loadIcon(name string) image.Image
  
  // NOTE: not concurrency-safe!
  func Icon(name string) image.Image {
      icon, ok := icons[name]
      if !ok {
          icon = loadIcon(name)
          icons[name] = icon
      }
      return icon
  }
  
  // 在创建goroutine之前的初始化阶段
  // 就初始化了map中的所有条目并且再也不去修改它们
  // 那么任意数量的goroutine并发访问Icon都是安全的，因为每一个goroutine都只是去读取而已
  var icons = map[string]image.Image{
      "spades.png":   loadIcon("spades.png"),
      "hearts.png":   loadIcon("hearts.png"),
      "diamonds.png": loadIcon("diamonds.png"),
      "clubs.png":    loadIcon("clubs.png"),
  }
  
  // Concurrency-safe.
  func Icon(name string) image.Image { return icons[name] }
  ```

- 第二种方法：避免从多个goroutine访问变量

  **不要通过共享内存来通信，而应该通过通信来共享内存。**

  使用通道请求来代理一个受限变量的所有访问的 goroutine 称为该变量的 `监控 goroutine`。

  ```go
  var deposits = make(chan int) // send amount to deposit
  var balances = make(chan int) // receive balance
  
  func Deposit(amount int) { deposits <- amount }
  func Balance() int       { return <-balances }
  
  // 使用一个 teller 作为监控goroutine 限制 balance 变量
  func teller() {
      var balance int // balance is confined to teller goroutine
      for {
          select {
          case amount := <-deposits:
              balance += amount
          case balances <- balance:
          }
      }
  }
  
  func init() {
      go teller() // start the monitor goroutine
  }
  
  
  // 如果一个变量无法在整个生命周期受限于单个 goroutine 
  // 可以通过借助通道来把共享变量的地址从上一步传到下一步
  // 从而在 “流水线” 上的多个 goroutine 之间共享该变量
  // 在流水线的每一步，再把变量地址传给下一步之后就不再访问该变量
  // 所有对这个变量的访问都是串行的，这种规则有时称为 “串行绑定”
  
  // Cakes 是串行绑定，先是baker gorouine，然后是icer gorouine
  // 思考：与设计模式中[状态模式]的实现关系
  type Cake struct{ state string }
  
  func baker(cooked chan<- *Cake) {
      for {
          cake := new(Cake)
          cake.state = "cooked"
          cooked <- cake // baker never touches this cake again
      }
  }
  
  func icer(iced chan<- *Cake, cooked <-chan *Cake) {
      for cake := range cooked {
          cake.state = "iced"
          iced <- cake // icer never touches this cake again
      }
  }
  ```

- 第三种方法：

  **允许多个goroutine去访问变量，但是在同一个时刻最多只有一个goroutine在访问。这种方式被称为 `互斥`。** 

