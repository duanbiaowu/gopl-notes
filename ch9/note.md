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

## 9.2 互斥锁: Sync.Mutex

**一个只能为1和0的信号量叫做二元信号量（binary semaphore）。**

```go
var (
    sema    = make(chan struct{}, 1) // a binary semaphore guarding balance
    balance int
)

func Deposit(amount int) {
    sema <- struct{}{} // acquire token
    balance = balance + amount
    <-sema // release token
}

func Balance() int {
    sema <- struct{}{} // acquire token
    b := balance
    <-sema // release token
    return b
}

// 互斥锁模式应用非常广泛，而且被sync包里的Mutex类型直接支持
// Lock方法能够获取到token(这里叫锁)，并且Unlock方法会释放这个token
import "sync"

// 函数、互斥锁、变量的组合方式称为 “监控模式”
var (
    mu      sync.Mutex // guards balance
    balance int
)

func Deposit(amount int) {
    mu.Lock()
    balance = balance + amount
    mu.Unlock()
}

func Balance() int {
    mu.Lock()
    b := balance
    mu.Unlock()
    return b
}
```

**一个 goroutine 访问bank变量时（这里只有balance余额变量），它都会调用mutex的Lock方法来获取一个互斥锁。如果其它的 goroutine 已经获得了这个锁的话，这个操作会被阻塞直到其它 goroutine 调用了 Unlock 使该锁变回可用状态。mutex会保护共享变量。**

按照惯例，被mutex所保护的变量是在mutex变量声明之后立刻声明的。如果你的做法和惯例不符，确保在文档里对你的做法进行说明。



**在 Lock 和 Unlock 之间的代码，可以自由地读取和修改共享变量，这一部分称为 `临界区域`。锁的持有者调用 Unlock 之前，其他 goroutine 不能获取锁。goroutine 在完成后应该立即释放锁，这包括函数的所有分支特别是错误分支。**

```go
// 很难判断对Lock和Unlock的调用是成对执行的
// defer 通过延迟执行 Unlock 就可以把临界区隐式扩展到当前函数的结尾
// 避免了必须在一个或者多个远离 Lock 的位置插入一条 Unlock 语句
// 带来的另一点好处是，我们再也不需要一个局部变量b了
func Balance() int {
    mu.Lock()
    defer mu.Unlock()
    return balance
}
```

**此外，在临界区域发生 panic 时，延迟执行的 Unlock 也会正确执行，这对于用 recover 来恢复的程序来说是很重要的。defer调用只会比显式地调用Unlock成本高那么一点点，不过却在很大程度上保证了代码的整洁性。大多数情况下对于并发程序来说，代码的整洁性比过度的优化更重要。如果可能的话尽量使用defer来将临界区扩展到函数的结束。**

```go
// NOTE: not atomic!
// 函数最终能给出正确的结果，单有一个不良的副作用
// 当过多的取款操作同时执行时，balance可能会瞬时被减到0以下
// 可能会引起一个并发的取款被不合逻辑地拒绝
// 如果Bob尝试买一辆sports car时，导致Alice无法支付早上的咖啡

// 问题在于取款不是一个原子操作，它包含了三个串行的操作
// 每个操作都申请并释放了互斥锁，但对于整个序列没有上锁
func Withdraw(amount int) bool {
    Deposit(-amount)
    if Balance() < 0 {
        Deposit(amount)
        return false // insufficient funds
    }
    return true
}


// NOTE: incorrect!
// Deposit 会调用mu.Lock()第二次去获取互斥锁
// 由于互斥锁是不能再入的（无法对一个已经上锁的互斥量再上锁）
// 这会导致程序死锁，Withdraw会永远阻塞下去
func Withdraw(amount int) bool {
    mu.Lock()
    defer mu.Unlock()
    Deposit(-amount)
    if Balance() < 0 {
        Deposit(amount)
        return false // insufficient funds
    }
    return true
}

// 一个通用的解决方案是将一个函数分离为多个函数，将Deposit分离成两个：
// 一个不可导出的函数deposit，假定已经获得互斥锁，并完成实际业务逻辑
// 一个可导出的函数Deposit，用来获取锁并调用 deposit
func Withdraw(amount int) bool {
    mu.Lock()
    defer mu.Unlock()
  	// deposit 函数代码很少，这里可以不用调用，直接修改 balance 变量即可
    // 主要通过这个例子很好地演示了规则
    deposit(-amount)
    if balance < 0 {
        deposit(amount)
        return false // insufficient funds
    }
    return true
}

func Deposit(amount int) {
    mu.Lock()
    defer mu.Unlock()
    deposit(amount)
}

func Balance() int {
    mu.Lock()
    defer mu.Unlock()
    return balance
}

// This function requires that the lock be held.
func deposit(amount int) { balance += amount }
```



**mutex的目的是确保共享变量在程序执行时的关键点上能够保证不变性。**

**不变性的一层含义是 “没有goroutine访问共享变量”，但实际上这里对于mutex保护的变量来说，不变性还包含更深层含义：当一个goroutine获得了一个互斥锁时，它能断定被互斥锁保护的变量正处于不变状态（译注：即没有其他代码块正在读写共享变量），在其获取并保持锁期间，可能会去更新共享变量，这样不变性只是短暂地被破坏，然而当其释放锁之后，锁必须保证共享变量重获不变性并且多个goroutine按顺序访问共享变量。**

**尽管一个可以重入的mutex也可以保证没有其它的goroutine在访问共享变量，但它不具备不变性更深层含义。**



**封装：通过在程序中减少对数据结构的非预期交互，来帮助我们保证数据结构中的不变量。因为类似的原因，封装也可以用来保持并发中的不变性。当你使用mutex时，确保mutex和其保护的变量没有被导出，无论这些变量是包级的变量还是一个struct的字段。**

