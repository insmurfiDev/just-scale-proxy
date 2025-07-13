package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/insmurfiDev/just-scale-proxy/tests/generated/my_package"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()

	cc, err := grpc.NewClient("0.0.0.0:9505", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		panic(err)
	}

	client := my_package.NewMyServiceClient(cc)

	var wg sync.WaitGroup

	sendReq := func(i int) {
		wg.Add(1)

		go func() {
			defer wg.Done()
			slog.Info("Отправляю запрос на proxy", slog.Int("id", i))

			_, err := client.SomeMethod(ctx, &my_package.SomeReq{
				Id: int64(i),
			})

			if err != nil {
				slog.Error("ошибка", err.Error())
			} else {
				slog.Info("запрос отправлен")
			}
		}()
	}

	for i := range 10 {
		sendReq(i)
	}

	wg.Wait()
}
