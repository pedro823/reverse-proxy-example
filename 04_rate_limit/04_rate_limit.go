package main

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

func main() {
	app := fiber.New()
	backend := "10.0.0.1:9000"

	// 1 request per second limiter
	limiter := rate.NewLimiter(1, 1)

	app.All("/", func(c *fiber.Ctx) error {
		c.Set("X-Backend", backend)

		reservation := limiter.Reserve()
		if !reservation.OK() || reservation.Delay() > 100*time.Millisecond {
			reservation.Cancel()
			c.Status(429)
			return c.SendString("Your request was rate limited. Please try again later.")
		}

		time.Sleep(reservation.Delay())

		return c.SendStatus(http.StatusOK)
	})

	app.Listen(":80")
}
