package main

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	userRegex := regexp.MustCompile("user(/.*)?$")
	settingsRegex := regexp.MustCompile("settings(/.*)?$")

	mapToBackend := map[*regexp.Regexp]string{
		userRegex:     "10.0.0.1",
		settingsRegex: "10.0.0.2",
	}

	catchAllBackend := "10.0.0.3"

	client := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	app.All("/*", func(c *fiber.Ctx) error {
		route := c.Params("*")

		var backend string

		for regex, target := range mapToBackend {
			if regex.MatchString(route) {
				backend = target
				break
			}
		}

		if backend == "" {
			backend = catchAllBackend
		}

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
