package utils

import (
	"fmt"
	"strings"

	"github.com/kataras/iris/v12"
)

func AppendError(existErr, newErr error) error {
	if existErr == nil {
		return newErr
	}

	return fmt.Errorf("%v, %w", existErr, newErr)
}

/********* IRIS func start region *********/
func PrintRequestHead(c iris.Context) {
	// 1. 打印请求头
	headers := c.Request().Header
	fmt.Println("Request Headers:")
	for key, values := range headers {
		fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
	}

	// 2. 打印请求体（注意会消耗读取流）
	c.RecordRequestBody(true) // 这样似乎不会消耗读取流

	body, err := c.GetBody()
	if err != nil {
		fmt.Println("Error reading request body:", err)
	} else {
		fmt.Printf("Request Body:\n%s\n", string(body))
	}
}

/********* IRIS func end region *********/
