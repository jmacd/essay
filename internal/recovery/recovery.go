package recovery

import (
	"fmt"
	"runtime"
)

func Here() func() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println("ENTER", file, ":", line)
	return func() {
		fmt.Println("EXIT", file, ":", line)
		if ret := recover(); ret != nil {
			fmt.Println("RECOVERED", ret)
		}
	}
}
