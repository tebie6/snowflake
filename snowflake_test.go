package snowflake

import (
	"fmt"
	"testing"
)

// go test -v snowflake_test.go snowflake.go
// 测试脚本
func TestSnowFlake(t *testing.T) {

	//for true {
	//	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	//	time.Sleep(1 * time.Second)
	//}

	// 生成节点实例
	worker, err := NewWorker(1)

	if err != nil {
		fmt.Println(err)
		return
	}

	ch := make(chan int64)
	count := 100000
	// 并发 count 个 goroutine 进行 snowflake ID 生成
	for i := 0; i < count; i++ {
		go func() {
			id := worker.GetId()
			ch <- id
		}()
	}

	defer close(ch)

	m := make(map[int64]int)
	for i := 0; i < count; i++ {
		id := <-ch
		// 如果 map 中存在为 id 的 key, 说明生成的 snowflake ID 有重复
		_, ok := m[id]
		if ok {
			t.Error("ID is not unique!\n")
			return
		}
		// 将 id 作为 key 存入 map
		m[id] = i
	}

	// 成功生成 snowflake ID
	fmt.Println("All", count, "snowflake ID Get successed!")
}
