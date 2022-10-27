package main

import (
	"math/rand"
	"time"

	"net/http"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	backends := []string{"10.0.0.1:9000", "10.0.0.2:9000"}
	client := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	app.All("/", func(c *fiber.Ctx) error {
		backendIdx := rand.Intn(len(backends))
		backend := backends[backendIdx]
		c.Set("X-Backend", backend)

		req, _ := http.NewRequest(http.MethodGet, "http://"+backend, nil)

		res, err := client.Do(req)
		if err != nil {
			return err
		}

		c.Status(res.StatusCode)
		return c.SendStream(res.Body)
	})

	app.Listen(":80")
}
