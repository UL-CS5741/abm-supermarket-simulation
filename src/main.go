package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//Customer struct defination
type Customer struct {
	cID      int
	numItems int
}

//Checkout struct defination
type Checkout struct {
}

var msgs = make(chan Customer)
var done = make(chan bool)

func (c *Customer) setNumItems(val int) {
	c.numItems = val
}
func createCustomer1() *Customer {
	id := rand.Intn(100)
	num := Random(1, 200)
	c := Customer{cID: id, numItems: num}
	return &c
}

func produce() {
	min, max := weatherEffect()
	n := Random(min, max)
	// customers := make([]*Customer, n)
	// fmt.Println("Today, We have total customers: ", n)
	// fmt.Println(customers)
	for i := 0; i < n; i++ {
		newCustomer := createCustomer1()
		msgs <- *newCustomer
	}
}
func consume(consumeNo int) {
	for {
		time.Sleep(1 * time.Second)
		msg := <-msgs
		fmt.Println("Checkout: # ", consumeNo, msg)
	}
}

//weather effects
func weatherEffect() (int, int) {
	var weather = [5]string{"Sunny", "Raining", "Cloudy", "Windy", "Chrismas"}
	var min, max int
	i := Random(0, len(weather))
	fmt.Println("Today the weather is ", weather[i])
	switch weather[i] {
	case "Sunny":
		min, max = 35, 55
	case "Raining":
		min, max = 0, 5
	case "Cloudy":
		min, max = 5, 25
	case "Windy":
		min, max = 15, 35
	case "Chrismas":
		min, max = 45, 60
	default:
		min, max = 15, 40
	}
	return min, max
}

var (
	mutex sync.Mutex
)

func init() {
	// Random seed
	rand.Seed(time.Now().UTC().UnixNano())
}

//Random number generator within a range of min amd max (closed)
func Random(min, max int) int {
	return rand.Intn(max - min + 1 + min)
}

func main() {
	numCheckout := 4
	numExpress := 0
	fmt.Println("*****************WELCOME TO GOLANGER'S SUPERMARKET*****************")
	fmt.Println("Please Enter Number of Checkout(1-8): ")
	_, err := fmt.Scanf("%d", &numCheckout)
	// checkouts := make(chan *Customer, numCheckout)
	if err != nil {
		fmt.Println("Wrong Input!")
	}
	fmt.Printf("Please Enter Number of Express Checkout(0 - %v): ", numCheckout)
	_, err = fmt.Scanf("%d", &numExpress)
	if err != nil {
		fmt.Println("Wrong Input!")
	}
	fmt.Println("Total Checkout: ", numCheckout)
	fmt.Println("Total Express Checkout: ", numExpress)
	go produce()
	for i := 1; i <= numCheckout; i++ {
		go consume(i)
		time.Sleep(2 * time.Second)
	}
	<-done
}
