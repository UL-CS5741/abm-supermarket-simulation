//CS5741 Supermarket checkout simulation assignment
//Team members : Golangers
//YUKTI PATIL - 20097786
//LEI HAO  - 19167024
//DIVYA GANGAMAGALA - 20002971
//AMAN NIYAZ - 20135262

package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	currentTime              time.Time
	totalCheckouts           int = 4
	checkoutLessThan5        int = 0
	totalCustomers           int
	r                        = rand.New(rand.NewSource(time.Now().UnixNano()))
	customers                []*Customer
	randomWeather            Weather
	totalWaitingTime         float64
	checkoutLock             sync.Mutex
	checkoutQueueLock        sync.Mutex
	customerLock             sync.Mutex
	totalCustomersCheckedOut int = 0
	totalItemsProcessed      int = 0
	customersLost            int = 0
)

//declare structure definition for Weather,Cashier,Consumer

type Checkout struct {
	checkoutId          int
	customerQueue       chan int
	itemsLimit          int
	BeginTime           int64
	totalRunningTime    float64
	accumulatedWaitTime float64
	customersProcessed  int
}
type Customer struct {
	customerId        int
	noOfProducts      int
	totalCheckoutTime float64
	WaitTime          int
	BeginTime         int64
	items             []Item
}

type Item struct {
	itemId       int
	checkoutTime float64
}

type Weather struct {
	BeginTime     int64
	weatherEffect int
}

//changeWeather function  helps in setting the weather based on time
func (w *Weather) changeWeather() {
	for {
		time.Sleep(time.Millisecond * 100)

		time := time.Now().UnixNano()
		gap := time - w.BeginTime
		// change the weatherEffect
		if gap%(3*1e10) > 2*1e10 {
			w.setWeather(3)
		} else if gap%(3*1e10) > 1e10 {
			w.setWeather(2)
		} else {
			w.setWeather(1)
		}
	}
}

// Set Weather Effect
func (w *Weather) setWeather(value int) {
	w.weatherEffect = value
}

//Set product limit on a checkout
func (c *Checkout) setLimit(inVal int) {
	c.itemsLimit = inVal
}

//Add customer to channel
func (c *Checkout) addCustomer(inVal int) {
	c.customerQueue <- inVal
}

//Assign Current Time
func (c *Checkout) setBeginTime() {
	time := time.Now().UnixNano()
	c.BeginTime = time
}

//Process Checkouts
func (c *Checkout) doCheckout(wg *sync.WaitGroup) {
	for {

		if len(c.customerQueue) > 0 {

			customerId := <-c.customerQueue

			checkoutLock.Lock()
			waitingTime := customers[customerId].totalCheckoutTime
			c.accumulatedWaitTime = c.accumulatedWaitTime + waitingTime
			totalWaitingTime += c.accumulatedWaitTime
			checkoutLock.Unlock()

			time.Sleep(time.Duration(waitingTime*1000) * time.Millisecond)

			checkoutLock.Lock()
			totalCustomersCheckedOut++
			c.customersProcessed++
			fmt.Printf("--------------------Checkout %v : Customer %2v checked out\tItems Processed : %2v \tWait time: %5.2f second(s)\n", c.checkoutId, customerId, customers[customerId].noOfProducts, c.accumulatedWaitTime)
			c.totalRunningTime = time.Since(currentTime).Seconds()

			checkoutLock.Unlock()
			wg.Done()
		}
	}
}

//Generate a random no between 0 and 20 for random arrival rate of customers at checkout
func (c *Customer) setWaitTime() {
	num := rand.Intn(20)
	c.WaitTime = num
}

//setBeginTime function sets the current time as the customers BeginTime
func (c *Customer) setBeginTime() {
	time := time.Now().UnixNano()
	c.BeginTime = time
}

//Generate a random number - for the numbers of products each customer will shop between 1 to 200
func (c *Customer) generateRandomNumberOfProducts() {
	num := rand.Float64()
	if num >= 0.2 {
		c.noOfProducts = rand.Intn(195) + 6
		test := rand.Intn(100)
		if test >= 99 {
			c.noOfProducts = rand.Intn(195) + 6
		} else if num >= 98 {
			c.noOfProducts = rand.Intn(175) + 6
		} else if num >= 97 {
			c.noOfProducts = rand.Intn(155) + 6
		} else if num >= 95 {
			c.noOfProducts = rand.Intn(135) + 6
		} else if num >= 90 {
			c.noOfProducts = rand.Intn(115) + 6
		} else if num >= 50 {
			c.noOfProducts = rand.Intn(71) + 6
		} else {
			c.noOfProducts = rand.Intn(21) + 6
		}
	} else {
		c.noOfProducts = rand.Intn(5) + 1
	}
}

//Generate a random time - equivalent to time taken by each customer for checkout
func (c *Customer) setCheckoutTime() {
	for i := 0; i < c.noOfProducts; i++ {
		c.items = append(c.items, Item{
			itemId: i})
		num := rand.Intn(100)
		if num >= 99 {
			c.items[i].checkoutTime = 5.0 + rand.Float64()
		} else if num >= 98 {
			c.items[i].checkoutTime = 4.0 + rand.Float64()
		} else if num >= 97 {
			c.items[i].checkoutTime = 3.0 + rand.Float64()
		} else if num >= 95 {
			c.items[i].checkoutTime = 2.0 + rand.Float64()
		} else if num >= 90 {
			c.items[i].checkoutTime = 1.0 + rand.Float64()
		} else {
			c.items[i].checkoutTime = float64(rand.Intn(50)+50) / 100
		}
	}
}

//Calculate checkout time which is sum of checkout time for each item
func (c *Customer) CalcCheckoutTime() {
	var totalTime float64
	for _, v := range c.items {
		totalTime = totalTime + v.checkoutTime
	}
	c.totalCheckoutTime = totalTime
}

// Add Customers to a checkout Queue
func (c *Customer) AddCustomersToQueue(checkouts []*Checkout, wg *sync.WaitGroup) {
	var num int = -1
	var cap int = 6
	time.Sleep(time.Second * time.Duration(c.WaitTime*randomWeather.weatherEffect)) // Weather would affect the customer rate arrival
	checkoutQueueLock.Lock()
	if c.noOfProducts <= 5 {
		for i := 0; i < checkoutLessThan5; i++ {
			if len(checkouts[i].customerQueue) < cap {
				num = i
				cap = len(checkouts[i].customerQueue)
			}
		}
	}

	if c.noOfProducts > 5 || num < 0 {
		for i := checkoutLessThan5; i < totalCheckouts; i++ {
			if len(checkouts[i].customerQueue) < cap {
				num = i
				cap = len(checkouts[i].customerQueue)
			}
		}
	}
	if num < 0 {
		customerLock.Lock()
		customersLost++
		fmt.Printf(" Customer %d left the store \n", c.customerId)
		fmt.Printf(" No. of customers left the store without purchasing : %d \n", customersLost)
		customerLock.Unlock()
		wg.Done()
	} else {
		c.setBeginTime()
		customerLock.Lock()
		totalItemsProcessed = totalItemsProcessed + c.noOfProducts
		checkouts[num].addCustomer(c.customerId)
		customerLock.Unlock()
	}
	checkoutQueueLock.Unlock()

}

//Print Results
func checkAndPrint(checkouts []*Checkout) {

	totalSimulationTime := time.Since(currentTime).Seconds()

	var avgUtilization float64
	var totalUtilization float64
	for _, n := range checkouts {

		checkoutUtilization := math.Floor(n.totalRunningTime / totalSimulationTime * 100)
		totalUtilization += checkoutUtilization
		fmt.Printf("********* Checkout %v Utilization: %3v%% \t\tCustomers through the checkout : %v\n", n.checkoutId, checkoutUtilization, n.customersProcessed)
	}
	avgUtilization = totalUtilization / float64(len(checkouts))
	fmt.Printf("********* Total products processed \t%10v\n", totalItemsProcessed)
	fmt.Printf("********* Average customer wait time \t%10.2f\n", float64(totalWaitingTime)/float64(totalCustomersCheckedOut))
	fmt.Printf("********* Average checkout utilization \t%10.2f\n", avgUtilization)
	fmt.Printf("********* Average products per trolley \t%10v\n", int(float64(totalItemsProcessed)/float64(totalCustomersCheckedOut)))
	fmt.Printf("********* The number of lost customers \t%10v\n", customersLost)
}

//Generate Customers
func GenerateCustomersAndCreateQueue(wg *sync.WaitGroup, checkouts []*Checkout) {
	for i := 0; i < totalCustomers; i++ {
		customers = append(customers, &Customer{
			customerId: i})
		customers[i].generateRandomNumberOfProducts()
		customers[i].setCheckoutTime()
		customers[i].CalcCheckoutTime()
		customers[i].setWaitTime()
		wg.Add(1)
		go customers[i].AddCustomersToQueue(checkouts, wg)
	}
}

//Read Inputs
func InputNumberOfCheckoutsAndConstumers() {
	fmt.Println("As a manager, enter the number of checkouts, Max 8 (Default: 4)")
	fmt.Scanln(&totalCheckouts)
	fmt.Println("Enter the number of Express Checkouts , Max 2 (Default: 0)")
	fmt.Scanln(&checkoutLessThan5)
	fmt.Println("Enter Number of Customers:")
	fmt.Scanln(&totalCustomers)
}

func main() {

	//Read No of checkouts, no of expressCheckouts, No of Customers
	InputNumberOfCheckoutsAndConstumers()

	currentTime := time.Now().UnixNano()
	var wg sync.WaitGroup

	//Weather agent - helps in deciding the customer arrival rate
	randomWeather = Weather{BeginTime: currentTime, weatherEffect: 1}
	go randomWeather.changeWeather()

	//Process Checkouts
	checkouts := make([]*Checkout, totalCheckouts)
	for i := 0; i < totalCheckouts; i++ {
		checkouts[i] = &Checkout{
			checkoutId: i, customerQueue: make(chan int, 6), totalRunningTime: 0.0}
		if i < checkoutLessThan5 {
			checkouts[i].setLimit(5)
		} else {
			checkouts[i].setLimit(200)
		}
		checkouts[i].setBeginTime()
		go checkouts[i].doCheckout(&wg) //start cashier agent
	}

	//Generate Customers
	GenerateCustomersAndCreateQueue(&wg, checkouts)
	wg.Wait()

	//Print Results
	fmt.Println("\n*****************Checkout Simulation Complete******************************")
	checkAndPrint(checkouts)
}
