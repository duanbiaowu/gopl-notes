package bank1_test

import (
	"../bank1"
	"fmt"
	"testing"
)

func TestBank(t *testing.T) {
	done := make(chan struct{})

	go func() {
		bank1.Deposit(200)
		fmt.Println("=", bank1.Balance())
		done <- struct{}{}
	}()

	go func() {
		bank1.Deposit(100)
		done <- struct{}{}
	}()

	<-done
	<-done

	if got, want := bank1.Balance(), 300; got != want {
		t.Errorf("Balance = %d, want %d", got, want)
	}
}
