package main

import (
	//"strconv"
	"sort"
	"strings"
	"sync"
	"fmt"
)

// SingleHash: crc32(data)+"~"+crc32(md5(data))
func SingleHash(in, out chan interface{}) {
	mutex := &sync.Mutex{}
	wgSH := &sync.WaitGroup{}

	for inputRaw := range in {								// получаем дату из канала in 

		data := fmt.Sprintf("%v", inputRaw)
		wgSH.Add(1)

		go func(data string, mutex *sync.Mutex, wgSH *sync.WaitGroup) {
			defer wgSH.Done()	

			mutex.Lock()
			md5Result := DataSignerMd5(data)				// считаем...
			mutex.Unlock()

			crcChan1 := make(chan string)
			go func(data string, outChan chan string) {
				crc32FirstResult := DataSignerCrc32(data)
				outChan <- crc32FirstResult
			}(data, crcChan1)

			crcChan2 := make(chan string)
			go func(data string, outChan chan string) {
				crc32SecondResult := DataSignerCrc32(data)
				outChan <- crc32SecondResult
			}(md5Result, crcChan2)

			crc32FirstResult := <-crcChan1
			crc32SecondResult := <-crcChan2

			out <- crc32FirstResult+"~"+crc32SecondResult	// пишем результат в канал out
		}(data, mutex, wgSH)
	}
	wgSH.Wait()
}

// OneMultiHash: res_th = crc32(th+data)), где th=0..5, -> конкатенация всех res_th в порядке от 0 до 5
func OneMultiHash(data string, wgMH *sync.WaitGroup, out chan interface{}) {
	defer wgMH.Done()
	wgOMH := &sync.WaitGroup{}
	result := [6]string{}	// слайс для результатов
	for th := 0; th < 6; th++ {
		wgOMH.Add(1)

		go func(data string,  th int){
			defer wgOMH.Done()
			res := DataSignerCrc32(fmt.Sprintf("%d%s", th, data))
			result[th] = res
		}(data, th)

	}
	wgOMH.Wait()
	out <- strings.Join(result[:], "")
}

// MultiHash: вызывает функцию, считающую MultiHash, для каждого полученного результата SingleHash
func MultiHash(in, out chan interface{}) {
	wgMH := &sync.WaitGroup{}

	for dataRaw := range in {	// получаем дату из канала in 
		data := fmt.Sprintf("%v", dataRaw)
		wgMH.Add(1)
		go OneMultiHash(data, wgMH, out)
	}

	wgMH.Wait()
}

// CombineResults - функция, собирающая результаты 
func CombineResults(in, out chan interface{}) {
	results := make([]string, 0)		// создаем слайс для результатов
	for inputRaw := range in {			// читаем результаты, пока они есть 
		results = append(results, fmt.Sprintf("%s", inputRaw))
	}
	sort.Strings(results) 				// сортируем результаты
    out <- strings.Join(results, "_") 	// соединяем результаты в строку и пишем их в канал out 
}

// JobWrapper - обертка для job для закрытия канала и работы с WaitGroup
func JobWrapper(job_ job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	job_(in, out)
	close(out)
}

// ExecutePipeline - функция, реализующая конвейер из поданных на вход функций
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}

	var chanIn chan interface{} = make(chan interface{})
	var chanOut chan interface{} = make(chan interface{})
	for _, job_ := range jobs {
		wg.Add(1)
		go JobWrapper(job_, chanIn, chanOut, wg)
		chanIn = chanOut
		chanOut = make(chan interface{})
	}
	wg.Wait()
}
