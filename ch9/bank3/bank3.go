package bank3

import "sync"

var (
	mu      sync.Mutex
	balance int
)

func Deposit(amount int) {
	mu.Lock()
	balance += amount
	mu.Unlock()
}

func Balance() int {
	mu.Lock()
	b := balance
	mu.Unlock()
	return b
}

//func Balance2() int {
//	mu.Lock()
//	defer mu.Lock()
//	return balance
//}

// NOTE: not atomic!
//func Withdraw(amount int) bool {
//	Deposit(-amount)
//	if Balance() < 0 {
//		Deposit(amount)
//		return false	// insufficient funds
//	}
//	return true
//}

// NOTE: incorrect!
//func Withdraw(amount int) bool {
//	mu.Lock()
//	defer mu.Unlock()
//	Deposit(-amount)
//	if Balance() < 0 {
//		Deposit(amount)
//		return false // insufficient funds
//	}
//	return true
//}

//func Withdraw(amount int) bool {
//	mu.Lock()
//	defer mu.Unlock()
//	deposit(-amount)
//	if balance < 0 {
//		deposit(amount)
//		return false // insufficient funds
//	}
//	return true
//}

//func Deposit(amount int) {
//	mu.Lock()
//	defer mu.Unlock()
//	deposit(amount)
//}

//func Balance() int {
//	mu.Lock()
//	defer mu.Unlock()
//	return balance
//}

// This function requires that the lock be held.
//func deposit(amount int) { balance += amount }
