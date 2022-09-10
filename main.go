package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"
)

// функция для записи строк в файл
func write(file *os.File, someString ...string) {
	resultString := ""

	//конкатенируем элементы слайса в одну строку
	for i := range someString {
		resultString += someString[i]
	}

	//производим попытку записи результирующей строки в файл
	_, writeToFileFailed := file.WriteString(resultString + "\n")

	//если произошла ошибка при записи - логируем это и выполнение программы останавливается
	if writeToFileFailed != nil {
		log.Fatal(writeToFileFailed)
	}
}

// функция для генерации псевдослучайных чисел
func random(channel *chan bool, channelNumber int) {
	//запоминаем время начала
	startTime := time.Now()

	//обращаемся к файлу
	//если файла с соответствующим названием нет - он создается, если есть - перезаписывается
	file, openFileFailed := os.Create("data" + strconv.Itoa(channelNumber) + ".txt")

	//если произошла ошибка при обращении к файлу - логируем это и выполнение программы останавливается
	if openFileFailed != nil {
		log.Fatal(openFileFailed)
	}

	//в конце выполнения функции закрываем файл, к которому обращались (на всякий случай)
	defer file.Close()

	//пишем в файл время и дату начала
	write(file, "Начало ", startTime.Format("01.02.2006 15:04:05"), ".", strconv.Itoa(startTime.Nanosecond()))

	//генерируем псевдослучайные числа и пишем их в файл
	for i := 0; i < 1000000; i++ {
		write(file, strconv.Itoa(rand.Int()))
	}

	//записываем в файл, сколько времени заняла генерация
	write(file, "Заняло ", time.Since(startTime).String())
	endTime := time.Now()

	fmt.Println("Горутина номер", channelNumber, "успешно завершила работу", endTime.Format("01.02.2006 15:04:05")+"."+strconv.Itoa(endTime.Nanosecond()))

	//отправляем в канал сообщение, что все числа сгенерировались
	*channel <- true
}

// функция получения пользовательского ввода строки
func userInput() string {
	//создаем новый сканер
	scanner := bufio.NewScanner(os.Stdin)

	//сканируем
	scanner.Scan()

	//введенный текст возвращаем на выход функции
	return scanner.Text()
}

func setGoroutinesAmount() int {
	//получаем строку от пользователя
	inputString := userInput()

	//производим попытку перевести полученную строку в int
	inputNumber, inputError := strconv.Atoi(inputString)

	//инициализируем переменную для результирующего числа
	var resultingAmount int

	//если ввели q - программу пора завершать
	if inputString == "q" {
		fmt.Println("До свидания")
	} else if inputString == "" { //если ввели пустую строку - выбираем число горутин автоматически (по количеству потоков процессора)
		resultingAmount = runtime.NumCPU() * 2
	} else if inputError != nil || inputNumber < 1 || inputNumber > 50 { //если ввод некорректен, сообщаем об этом пользователю и рекурсивно вызываем функцию определения числа горутин
		fmt.Println("Некорректный ввод. Нужно ввести целое число, находящееся в диапазоне от 1 до 50 включительно")
		return setGoroutinesAmount()
	} else {
		resultingAmount = inputNumber
	}
	return resultingAmount
}

func main() {
	//восстанавливаем работу, если что-то пошло не так
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("восстановление после", r, "прошло успешно")
		}
	}()

	//выводим сообщение для пользователя
	fmt.Println("Введите число горутин (от 1 до 50 включительно), либо пустую строку, для автоматического определения\nВвод (для выхода из программы введите q): ")

	//устанавливаем число горутин
	goroutinesAmount := setGoroutinesAmount()

	//если установленное число горутин равно 0, значит был запрошен выход из программы
	if goroutinesAmount == 0 {
		return
	}

	//рандомизируем генератор с помощью текущего значения времени
	rand.Seed(time.Now().Unix())
	//создаем слайс каналов, для коммуникации между горутинами
	channelsDoneSlice := make([]chan bool, goroutinesAmount)
	for i := 0; i < goroutinesAmount; i++ {
		channelsDoneSlice[i] = make(chan bool)
	}

	//создаем канал, на который придет сообщение, что все горутины закончили работу
	allDoneChannel := make(chan bool)
	//запускаем горутины для генерации случайных чисел
	for i := range channelsDoneSlice {
		go random(&channelsDoneSlice[i], i+1)
	}

	//анонимная функция, запускаемая в горутине, проверяющая, отработали ли все горутины, генерирующие случайные числа
	go func() {
		channelsDoneBoolSlice := make([]bool, len(channelsDoneSlice))
		isDone := false
		for i := range channelsDoneBoolSlice {
			channelsDoneBoolSlice[i] = <-channelsDoneSlice[i]
			if channelsDoneBoolSlice[i] {
				if i == len(channelsDoneBoolSlice)-1 {
					isDone = true
				}
			} else {
				break
			}
		}
		if isDone {
			allDoneChannel <- true
		}
	}()

	//ожидание сообщения о завершении работы остальных горутин (main - тоже горутина, по идее)
	<-allDoneChannel

	//вывод сообщения
	fmt.Println("Все потоки отработали успешно")
}
