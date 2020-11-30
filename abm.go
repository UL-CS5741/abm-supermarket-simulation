package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	wg          sync.WaitGroup
	currentTime time.Time  // the time that all checkout can work(but the customer may not arrive the checkout)
	idMutex     sync.Mutex // to generate unique id for each customer, different goroutines generates cusId, so we have to lock it
	sharedMutex sync.Mutex //protect some code for different goroutines
	cusId       int

	//record customer numbers for different checkouts,
	//the value of cusNumEachCheckout[index] will reduce once the checkout finishes a customer,
	//so when cusNumEachCheckout[index]==0 we can know that there is no customers in the checkout, and record the time
	cusNumEachCheckout [8]int

	// in a checkout, (the waiting time for customer3 = (waiting time for customer2 + the waiting time for customer3)
	//f(x) = f(x-1) + f(x-2)
	accumulatedWaitingTime [8]float32

	//checkoutRunningTime[index] = time(checkout index finishes the last customer) - time(checkout index begins with the first customer)
	checkoutRunningTime [8]float64

	//checkoutUtilization[index] =  checkoutRunningTime[index]/time(the time when all the checkout finish)
	//for the recording: total utilization for each checkout
	//for the recording: average checkout utilization = sum(checkoutUtilization)/checkoutNum
	checkoutUtilization [8]float64

	arriveTimeToCheckout int //depend on the weather agent

	//for the recording: total products processed
	productNumForEachCheckout [8]int //

	//cusNumEachCheckout will decrease but we need the total customers in the end
	cusNumEachCheckoutCopy [8]int
	totalWaitingTime       float32

	numOfLostCustomer int

	//every checkout(as consumer) has a channel, every producer will generate customer to a checkout through the corresponding channel
	channels [8]chan customer

	//•	the time for each product to be entered at the checkout to be generated randomly within a user specified range (.5 to 6)
	//here i simplify the condition, any other ideas?
	timeForEachProduct = [...]float32{0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6}
	//if isCheckoutRestricted[index]==true, the numOfItem will less than 5
	isCheckoutRestricted = [...]bool{false, false, false, false, false, false, false, false}
)

//make it simple, modify the struct(like we discuss in the meeting, add some attributes or methods like purchase to change the trolleyItem)
type customer struct {
	customerId       int
	trolleyItem      int
	totalWaitingTime int //is this the tolerance time for each customer?
}
checkouts := make(chan *Customer, numCheckout)

func main() {

	var checkoutNum int = 4
	var restrictedCheckoutNum int = 0
	var tryTime int = 5

	for i := 0; i < tryTime; i++ {
		fmt.Println("Please input the checkoutNum(1 to 8, default 4) and the restricted checkoutNum(default 0):")
		stdin := bufio.NewReader(os.Stdin)
		_, err := fmt.Fscan(stdin, &checkoutNum, &restrictedCheckoutNum)
		stdin.ReadString('\n')
		if err != nil {
			//fmt.Println(err)
			//fmt.Printf("count:%d\n", count)
			//fmt.Printf("people:%d %d\n", checkoutNum,restrictedCheckoutNum)
			fmt.Println("wrong input!!!!!")
			continue
		} else if checkoutNum < restrictedCheckoutNum {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The restrictedCheckoutNum must be less than the checkoutNum")
			continue
		} else if checkoutNum > 8 {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The checkoutNum should be 1 to 8")
			continue
		} else if checkoutNum < 1 {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The checkoutNum should be 1 to 8")
			continue
		} else {
			//fmt.Printf("count:%d\n", count)
			fmt.Printf("checkout Num and restrictedCheckoutNum:%d %d\n", checkoutNum, restrictedCheckoutNum)
			break
		}
	}

	initData()
	go timer(&wg)                                            // the service time for the supermarket
	handleRestriction(0, checkoutNum, restrictedCheckoutNum) // generate restrictedCheckout randomly
	fmt.Println("restricted checkout condition", isCheckoutRestricted)

	fmt.Println("all the checkouts are waiting customers to the checkout...")
	time.Sleep(time.Duration(arriveTimeToCheckout) * time.Second)

	wg.Add(1)

	for i := 0; i < checkoutNum; i++ {
		channel := make(chan customer, 1)  // 6
		channels[i] = channel
		go checkout(i, channels[i], 1.5) // 1.5second for testing, the value should be 0.5-6
	}
	//go checkout(2,channels[2],GenerateRangeTimeForEachProduct(0,12))

	for i := 0; i < checkoutNum; i++ {
		go producer(i, channels[i])
	}

	wg.Wait()

}

func handleRestriction(min int, max int, n int) {
	restrictionCheckoutId := randomArray(min, max, n)
	for _, n := range restrictionCheckoutId {
		isCheckoutRestricted[n] = true
	}
}

func createNewCustomer(customerId int) *customer {
	p := customer{customerId: customerId}
	return &p
}

func producer(checkoutNo int, ch chan customer) {

	// the document don't mention how many customers a checkout should have? 6-8 is for testing
	cusNum := GenerateRangeNum(6, 8) //it will just give 6 or 7
	fmt.Printf("checkout %v has %v customers\n", checkoutNo, cusNum)
	if cusNum > 6 { //does this mean number of lost customer can't be more than 3
		numOfLostCustomer += cusNum - 6
	}
	cusNumEachCheckout[checkoutNo] = cusNum
	cusNumEachCheckoutCopy[checkoutNo] = cusNum
	for j := 0; j < cusNum; j++ {
		newCustomer := createNewCustomer(generateId())
		if isCheckoutRestricted[checkoutNo] {
			newCustomer.trolleyItem = GenerateRangeNum(1, 6) // need to be less than 5
		} else {
			// •	the number of products for each trolley to be generated randomly within a user specified range (1 to 200).
			//1 to 10 for testing
			newCustomer.trolleyItem = GenerateRangeNum(1, 201) //changed it for getting max limit 200
		}
		productNumForEachCheckout[checkoutNo] += newCustomer.trolleyItem
		ch <- *newCustomer

	}

}

func initData() {
	rand.Seed(time.Now().Unix())
	currentTime = time.Now()
	cusId = 0
	arriveTimeToCheckout = initArrivingTimeToCheckout() //this only return number of customers //rateofCustomer
	fmt.Println("according to the weather condition, waiting time for all checkout could be: ", arriveTimeToCheckout)
}

//modify the weather agent, this is based on my individual understanding
type Weather string

const (
	Cloudy  Weather = "Cloudy"
	Sunny   Weather = "Sunny"
	Raining Weather = "Raining"
)

var weatherArray = [3]Weather{Cloudy, Sunny, Raining}

func initArrivingTimeToCheckout() int { //this function is just returning number of customers arrival depending on the weather
	weather := weatherArray[GenerateRangeNum(0, len(weatherArray))]
	fmt.Println("The weather is:", weather)
	switch Cloudy { //here we should have switch weather
	case Cloudy:
		return GenerateRangeNum(0, 5)
	case Sunny:
		return GenerateRangeNum(20, 41)
	case Raining:
		return GenerateRangeNum(40, 61) //why would raining day is going to have such values??? Isn't this too high?
	default:
		return GenerateRangeNum(20, 41)
	}
}

func generateId() int {
	idMutex.Lock()
	cusId += 1
	defer idMutex.Unlock()
	return cusId
}

func checkout(checkoutNo int, ch chan customer, timeForProduct float32) {
	for {
		customer := <-ch
		waitingTime := (float32)(customer.trolleyItem) * timeForProduct  //ProcessingTime

		sharedMutex.Lock()
		accumulatedWaitingTime[checkoutNo] = accumulatedWaitingTime[checkoutNo] + waitingTime
		totalWaitingTime += accumulatedWaitingTime[checkoutNo]
		sharedMutex.Unlock()
		time.Sleep(time.Duration(waitingTime*1000) * time.Millisecond)

		sharedMutex.Lock()
		cusNumEachCheckout[checkoutNo] = cusNumEachCheckout[checkoutNo] - 1

		fmt.Printf("--------------------------------checkout %v finished for customer %v, the waiting time is %v second(s)\n", checkoutNo, customer.customerId, accumulatedWaitingTime[checkoutNo])
		if cusNumEachCheckout[checkoutNo] == 0 {
			fmt.Printf("###################################################checkout %v finished and running time is %v\n", checkoutNo, time.Since(currentTime).Seconds())
			checkoutRunningTime[checkoutNo] = time.Since(currentTime).Seconds()
			checkAndPrint()
		}
		sharedMutex.Unlock()
	}
}
func checkAndPrint() {
	printOrNot := true
	for _, n := range cusNumEachCheckout {
		if n > 0 {
			printOrNot = false
		}
	}
	if printOrNot == true {
		fmt.Println("All checkouts have finished ")

		totalRunningTime := time.Since(currentTime).Seconds()
		var checkoutCount float64
		var avgUtilization float64
		var totalUtilization float64
		for i, n := range checkoutRunningTime {
			if n > 0 {
				checkoutUtilization[i] = math.Floor(n / totalRunningTime * 100)
				totalUtilization += checkoutUtilization[i]
				checkoutCount++
				fmt.Printf("********** utilization for each checkout: checkout %v - %v\n", i, checkoutUtilization[i])
			}
		}
		avgUtilization = totalUtilization / checkoutCount
		fmt.Printf("********** average checkout utilization: %v\n", avgUtilization)

		var totalProduct int
		for _, n := range productNumForEachCheckout {
			totalProduct += n

		}
		fmt.Printf("********** total products processed %v\n", totalProduct)

		var totalCustomer int
		for _, n := range cusNumEachCheckoutCopy {
			totalCustomer += n
		}
		fmt.Printf("********** average products per trolley: %v\n", float64(totalProduct)/float64(totalCustomer))
		fmt.Printf("********** average customer wait time: %v\n", float64(totalWaitingTime)/float64(totalCustomer))
		fmt.Printf("********** the number of lost customers: %v\n", numOfLostCustomer)

		wg.Done()

	}
}

// service time for the supermarket
func timer(wg *sync.WaitGroup) {
	time.Sleep(600 * time.Second)
	wg.Done()
}

func GenerateRangeNum(min, max int) int {
	randNum := rand.Intn(max-min) + min
	//fmt.Printf("rand is %v\n", randNum)
	return randNum
}

func GenerateRangeTimeForEachProduct(min, max int) float32 {
	randNum := rand.Intn(max-min) + min

	return timeForEachProduct[randNum]
}

func randomArray(min int, max int, n int) (result []int) {
	length := max - min

	source := make([]int, length)
	for i := min; i < min+length; i++ {
		source[i-min] = i
	}

	index := 0
	for i := 0; i < n; i++ {
		index = int(math.Abs(float64(rand.Int() % (length))))
		length = length - 1
		result = append(result, source[index])
		source[index] = source[length]

	}
	fmt.Println("the restriction checkout: ", result)
	return result

}
