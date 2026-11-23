package handler

import (
	"fmt"
	"sync"
)

func printABC() {
	// 定义3个channel
	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)
	ch3 := make(chan struct{}, 1)

	var wg sync.WaitGroup
	wg.Add(3)

	// 打印a
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch1
			fmt.Println("a")
			ch2 <- struct{}{}
		}
	}()

	// 打印b
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch2
			fmt.Println("b")
			ch3 <- struct{}{}
		}
	}()

	// 打印c
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			<-ch3
			fmt.Println("b")
			// 为了防止阻塞，第10次不写入ch1
			if i < 9 {
				ch1 <- struct{}{}
			}
		}
	}()

	// 启动
	ch1 <- struct{}{}
	wg.Wait()
	close(ch1)
	close(ch2)
	close(ch3)
}
