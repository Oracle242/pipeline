package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	m     sync.Mutex
}

// Функция конструктор
func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

// Добовление эллемента
func (r *RingIntBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
	} else {
		r.pos++
		r.array[r.pos] = el
	}
}

// Поиск эллемента
func (r *RingIntBuffer) Get() []int {
	if r.size <= 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.array[:r.pos+1]
	r.pos = -1
	log.Println("Передача данных для поиска эл-та")
	return output
}

func read(nextStage chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			log.Println("Введен неверный символ")
			fmt.Println("Программа обрабатывает толлько целые числа")
			continue
		}
		nextStage <- i
	}
}

func negativeFilterStageInt(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {

	for {
		select {
		case data := <-previosStageChannel:
			log.Println("Сортировка на положительные числа")
			if data > 0 {
				nextStageChannel <- data
			}
		case <-done:
			return
		}
	}
}

func notDividedThreeFunc(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool) {

	for {
		select {
		case data := <-previosStageChannel:
			log.Println("Сортировка на числа кратные 3м")
			if data%3 == 0 {
				nextStageChannel <- data
			}
		case <-done:
			return
		}
	}
}

var bufferSize int = 10
var bufferDrainInterval time.Duration = 10 * time.Second

func bufferStageFunc(previosStageChannel <-chan int, nextStageChannel chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingIntBuffer(size)
	for {
		select {
		case data := <-previosStageChannel:
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			if bufferData != nil {
				for _, data := range bufferData {
					nextStageChannel <- data
				}
			}
		case <-done:
			return
		}
	}
}

func main() {

	input := make(chan int)
	done := make(chan bool)
	go read(input, done)

	negativeChanFilter := make(chan int)
	go negativeFilterStageInt(input, negativeChanFilter, done)

	notDividedThreeCannel := make(chan int)
	go notDividedThreeFunc(negativeChanFilter, notDividedThreeCannel, done)

	bufferedInthannel := make(chan int)
	go bufferStageFunc(notDividedThreeCannel, bufferedInthannel, done, bufferSize, bufferDrainInterval)

	for {
		select {
		case data := <-bufferedInthannel:
			log.Printf("Вывод данных пользователю %d", data)
			fmt.Println("Обработанные данные", data)
		case <-done:
			return
		}
	}
}
