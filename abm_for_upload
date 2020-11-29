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
	wg sync.WaitGroup
	currentTime time.Time // the time that all checkout can work(but the customer may not arrive the checkout)
	idMutex   sync.Mutex // to generate unique id for each customer, different goroutines generates cusId, so we have to lock it
	sharedMutex sync.Mutex //protect some code for different goroutines
	cusId int

	//record customer numbers for different checkouts,
	//the value of cusNumEachCheckout[index] will reduce once the checkout finishes a customer,
	//so when cusNumEachCheckout[index]==0 we can know that there is no customers in the checkout, and record the time
	cusNumEachCheckout [8]int

	// in a checkout, the waiting time for customer3 = the processing time for customer2 + the processing time for customer1
	accumulatedWaitingTime [8]float32

	//checkoutRunningTime[index] = time(checkout index finishes the last customer) - time(checkout index begins with the first customer)
	checkoutRunningTime [8] float64

	//checkoutUtilization[index] =  checkoutRunningTime[index]/time(the time when all the checkout finish)
	//for the recording: total utilization for each checkout
	//for the recording: average checkout utilization = sum(checkoutUtilization)/checkoutNum
	checkoutUtilization [8] float64


	arriveTimeToCheckout int //depend on the weather agent

	//for the recording: total products processed
	productNumForEachCheckout [8] int

	//cusNumEachCheckout will decrease but we need the total customers in the end, so we have to have a copy
	cusNumEachCheckoutCopy [8]int
	totalWaitingTime float32


	numOfLostCustomer int

	//every checkout(as consumer) has a channel, every producer will generate customer to a checkout through the corresponding channel
	channels  [8] chan customer

	//********** the time for each product to be entered at the checkout to be generated randomly within a user specified range (.5 to 6)
	timeForEachProduct = [...]float32{0.5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6}
	//if isCheckoutRestricted[index]==true, the numOfItem will less than 5
	isCheckoutRestricted = [...]bool{false,false,false,false,false,false,false,false}

	cusNumFromTimeOfTheDay int // depend on the part of the day
	cusNumForEachCheckoutInRandom [8]int // 20 5checkouts[ 4 4 4 5 3 0 0 0]
)


type customer struct {
	customerId int
	trolleyItem int

}

func main(){

	var checkoutNum int = 4
	var restrictedCheckoutNum int = 0
	var tryTime int =5

	for i:=0;i<tryTime;i++{
		fmt.Println("Please input the checkoutNum(1 to 8, default 4) and the restricted checkoutNum(default 0):")
		stdin := bufio.NewReader(os.Stdin)
		_, err := fmt.Fscan(stdin, &checkoutNum,&restrictedCheckoutNum)
		stdin.ReadString('\n')
		if err != nil {
			fmt.Println("wrong input!!!!!")
			continue
		}else if checkoutNum<restrictedCheckoutNum {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The restrictedCheckoutNum must be less than the checkoutNum")
			continue
		}else if checkoutNum>8 {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The checkoutNum should be 1 to 8")
			continue
		}else if checkoutNum<1 {
			fmt.Println("wrong input!!!!!")
			fmt.Println("The checkoutNum should be 1 to 8")
			continue
		}else{
			//fmt.Printf("checkout Num and restrictedCheckoutNum:%d %d\n", checkoutNum,restrictedCheckoutNum)
			break
		}
	}


	initData()
	handleCusNumForEachCheckout(checkoutNum)
	handleRestriction(0,checkoutNum,restrictedCheckoutNum) // generate restrictedCheckout randomly
	go timer(&wg)// the service time for the supermarket

	fmt.Println("all the checkouts are waiting customers to the checkout...")
	time.Sleep(time.Duration(arriveTimeToCheckout)*time.Second)


	wg.Add(1)

	for i:=0;i<checkoutNum;i++ {
		channel:=make(chan customer,6)
		channels[i] = channel
		//go checkout(i,channels[i],1.5)// 1.5second is for testing, the value should be 0.5-6 generated in random
		go checkout(i,channels[i],GenerateRangeTimeForEachProduct(0,12))// 1.5second is for testing, the value should be 0.5-6 generated in random
	}



	for i:=0;i<checkoutNum;i++ {
		go producer(i,channels[i])
	}


	wg.Wait()

}


// Weather agent,the rate the customer arrive at the checkouts to be generated randomly within a user specified range (0 to 60).
// (depending on the weather agent)
type Weather string
const (
	Cloudy Weather = "Cloudy"
	Sunny Weather = "Sunny"
	Raining Weather = "Raining"
)
var weatherArray = [3]Weather{Cloudy,Sunny,Raining}
func initArrivingTimeToCheckout() int{
	weather:=weatherArray[GenerateRangeNum(0, len(weatherArray))]
	fmt.Println("---------- The weather is:",weather)
	switch weather {
	case Cloudy:return GenerateRangeNum(0,21)
	case Sunny:return GenerateRangeNum(20,41)
	case Raining:return GenerateRangeNum(40,61)
	default: return GenerateRangeNum(20,41)
	}
}



// Time agent, decide number of customer(10-50)
type TimeOfTheDay string
const (
	Morning TimeOfTheDay = "Morning"
	Afternoon TimeOfTheDay = "Afternoon"
	Evening TimeOfTheDay = "Evening"
)
var timeArray = [3]TimeOfTheDay{Morning,Afternoon,Evening}
func initCusNumFromTimeOfTheDay() int{
	// customerNum: from 10 to 40
	time:=timeArray[GenerateRangeNum(0, len(timeArray))]
	fmt.Println("---------- The part of the day is:",time)
	switch time {
	case Morning:return GenerateRangeNum(10,21)
	case Afternoon:return GenerateRangeNum(20,31)
	case Evening:return GenerateRangeNum(30,41)
	default: return GenerateRangeNum(20,31)
	}
}


//Producer
func producer(checkoutNo int,ch (chan customer)){

	cusNum := cusNumForEachCheckoutInRandom[checkoutNo]
	cusNumEachCheckout[checkoutNo] = cusNum
	cusNumEachCheckoutCopy[checkoutNo] = cusNum
	lostCusNum:=0
	if cusNum > 6 {
		lostCusNum = cusNum-6
		cusNumEachCheckout[checkoutNo] = 6
		cusNumEachCheckoutCopy[checkoutNo] = 6
		numOfLostCustomer += cusNum-6
	}
	fmt.Printf("checkout %v has %v customers and the lost number of customers can be %v\n",checkoutNo,cusNum,lostCusNum)
	for j:=0;j<cusNumEachCheckoutCopy[checkoutNo];j++ {
		newCustomer := createNewCustomer(generateId())
		if isCheckoutRestricted[checkoutNo] {
			newCustomer.trolleyItem = GenerateRangeNum(1,6)// need to be less than 5
		}else{
			// •	the number of products for each trolley to be generated randomly within a user specified range (1 to 200).
			newCustomer.trolleyItem = GenerateRangeNum(1,201)
		}
		productNumForEachCheckout[checkoutNo]+=newCustomer.trolleyItem
		ch <- *newCustomer

	}
}

//Consumer
func checkout(checkoutNo int,ch chan customer,timeForProduct float32) {
	for {
		customer := <-ch
		processingTime := (float32)(customer.trolleyItem)*timeForProduct

		sharedMutex.Lock()
		accumulatedWaitingTime[checkoutNo]=accumulatedWaitingTime[checkoutNo]+processingTime
		totalWaitingTime +=accumulatedWaitingTime[checkoutNo]
		sharedMutex.Unlock()
		time.Sleep(time.Duration(processingTime*1000)*time.Millisecond)


		sharedMutex.Lock()
		cusNumEachCheckout[checkoutNo] = cusNumEachCheckout[checkoutNo]-1


		fmt.Printf("--------------------------------checkout %v finished for customer %v, the waiting time is %v second(s)\n", checkoutNo,customer.customerId,accumulatedWaitingTime[checkoutNo])
		if cusNumEachCheckout[checkoutNo]==0 {
			fmt.Printf("###################################################checkout %v finished and running time is %v second(s)\n",checkoutNo, time.Since(currentTime).Seconds())
			checkoutRunningTime[checkoutNo] = time.Since(currentTime).Seconds()
			checkAndPrint()
		}
		sharedMutex.Unlock()
	}
}

func checkAndPrint(){
	printOrNot := true
	for _,n := range cusNumEachCheckout {
		if n > 0 {
			printOrNot = false
		}
	}
	if printOrNot==true {
		fmt.Println("All checkouts have finished ")

		totalRunningTime :=time.Since(currentTime).Seconds()
		var checkoutCount float64
		var avgUtilization float64
		var totalUtilization float64
		for i,n := range checkoutRunningTime {
			if n > 0 {
				checkoutUtilization[i] = math.Floor(n/totalRunningTime*100)
				totalUtilization+=checkoutUtilization[i]
				checkoutCount++
				fmt.Printf("********** utilization for each checkout: checkout %v - %v %%\n",i,checkoutUtilization[i])
			}
		}
		avgUtilization = totalUtilization/checkoutCount
		fmt.Printf("********** average checkout utilization: %v %%\n",avgUtilization)

		var totalProduct int
		for _,n := range productNumForEachCheckout {
			totalProduct+=n

		}
		fmt.Printf("********** total products processed %v\n",totalProduct)

		var totalCustomer int
		for _,n :=range cusNumEachCheckoutCopy {
			totalCustomer+=n;
		}
		fmt.Printf("********** average products per trolley: %v\n",float64(totalProduct)/float64(totalCustomer))
		fmt.Printf("********** average customer wait time: %v second(s)\n",float64(totalWaitingTime)/float64(totalCustomer))
		fmt.Printf("********** the number of lost customers: %v\n",numOfLostCustomer)

		wg.Done()


	}
}
















func handleCusNumForEachCheckout(checkoutNum int){
	cusNumFromTimeOfTheDay = initCusNumFromTimeOfTheDay();
	fmt.Println("according to the part of the day, customer number could be: ",cusNumFromTimeOfTheDay)
	arr:=distributeCustomerInRandom(cusNumFromTimeOfTheDay,checkoutNum)
	for i,n := range arr {
		cusNumForEachCheckoutInRandom[i] = n
	}

}


func initData() {
	rand.Seed(time.Now().Unix())
	currentTime = time.Now()
	cusId = 0
	arriveTimeToCheckout = initArrivingTimeToCheckout()
	fmt.Printf("according to the weather condition, waiting time for all checkout could be: %v second(s)\n",arriveTimeToCheckout)
}


// for example: distribute 30 customers to 5 checkouts in random
func distributeCustomerInRandom(customerNum int, checkoutNum int) []int {
	var arr []int
	totalCustomer:= (customerNum);

	leftCustomer := totalCustomer;
	leftCount := checkoutNum;
	for i := 0; i < checkoutNum - 1; i++ {
		money_ := 0;
		if leftCustomer > 0 {
			if (leftCustomer / leftCount * 2) < 1 {
				money_ = leftCustomer;
			} else {
				money_ = 1 + rand.Intn(leftCustomer / leftCount * 2);
			}

		} else {
			money_ = 0;
		}
		arr = append(arr, money_);
		if money_ > 0 {
			leftCustomer -= money_;
			leftCount--;
		}

	}
	arr = append(arr,leftCustomer);
	return arr;
}


// service time for the supermarket
func timer(wg *sync.WaitGroup){
	// 100 minutes
	time.Sleep(6000*time.Second)
	wg.Done()
}


func generateId() int{
	idMutex.Lock()
	cusId+=1
	defer idMutex.Unlock()
	return cusId
}


func GenerateRangeNum(min, max int) int {
	randNum := rand.Intn(max - min) + min
	//fmt.Printf("rand is %v\n", randNum)
	return randNum
}


func GenerateRangeTimeForEachProduct(min, max int) float32 {
	randNum := rand.Intn(max - min) + min

	return timeForEachProduct[randNum]
}

func randomArray(min int,max int, n int) (result []int) {
	length := max-min;

	source :=make([]int,length)
	for i := min; i < min+length; i++{
		source[i-min] = i;
	}

	index := 0;
	for i := 0; i < n; i++ {
		index = int(math.Abs(float64(rand.Int() % (length))))
		length = length-1
		result = append(result,source[index])
		source[index] = source[length];

	}
	return result;

}

func handleRestriction(min int,max int, n int){
	restrictionCheckoutId:=randomArray(min,max,n)
	for _,n :=range restrictionCheckoutId{
		isCheckoutRestricted[n] = true
	}
}


func createNewCustomer(customerId int) *customer{
	p := customer{customerId:customerId}
	return &p
}
