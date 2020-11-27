package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type customer struct {
	customerId       int
	items            []Item
	totalWaitingTime float32
}

type checkout struct {
	customers []customer
}

type Store struct {
	checkouts []checkout
}

type Item struct {
	itype    string
	scanTime float32
}

var store Store
var waitGroup sync.WaitGroup
var sharedMutex sync.Mutex

func main() {

	checkoutChannel := make(chan int, 6)
	for i := 0; i < 8; i++ {
		var c checkout
		store.checkouts = append(store.checkouts, c)
		go store.checkouts[0].checkoutCart(checkoutChannel)
	}
	waitGroup.Add(1)
	for i := 0; i < 30; i++ {
		checkoutChannel <- i
	}
	close(checkoutChannel)
	waitGroup.Wait()
}

func (c *checkout) checkoutCart(checkoutChannel chan int) {
	for index := range checkoutChannel {
		var cust customer
		cust.customerId = index
		noOfItems := GenerateRangeNum(1, 10)
		for i := 0; i < noOfItems; i++ {
			var items []Item
			itype := "item" + strconv.Itoa(i)
			scanTime := GenerateRangeTimeForEachProduct(1, 6)
			items = append(items, Item{itype, scanTime})
			cust.items = items
		}

		var customerWaitingTime float32
		for _, t := range cust.items {
			customerWaitingTime = customerWaitingTime + t.scanTime
		}

		// previouscutomerindex := index - 1
		// if previouscutomerindex > 1 {
		// 	cust.totalWaitingTime = customerWaitingTime + c.customers[previouscutomerindex].totalWaitingTime
		// } else {
		cust.totalWaitingTime = customerWaitingTime
		//}
		sharedMutex.Lock()
		c.customers = append(c.customers, cust)
		sharedMutex.Unlock()
		time.Sleep(time.Duration(customerWaitingTime*1000) * time.Millisecond)

	}
	fmt.Printf("Total number of customers processed : %v \n", len(c.customers))
	waitGroup.Done()
}

func GenerateRangeNum(min, max int) int {
	randNum := rand.Intn(max-min) + min
	return randNum
}

func GenerateRangeTimeForEachProduct(min, max int) float32 {
	timeForEachProduct := [...]float32{0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6}
	randNum := rand.Intn(max-min) + min
	return timeForEachProduct[randNum]
}
