package serverwaiter

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	fmt.Println(Wait(ctx, "https://google.com"))
}
