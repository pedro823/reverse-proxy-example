package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	routeMapChan := make(chan map[*regexp.Regexp]string)
	var routeMap map[*regexp.Regexp]string

	go func() {
		for {
			newRouteMap := <-routeMapChan
			routeMap = newRouteMap
			fmt.Println("Configured new routeMap")
		}
	}()

	setupControlPlane(routeMapChan)

	catchAllBackend := "10.0.0.3"

	client := &http.Client{
		Timeout: 50 * time.Millisecond,
	}

	app.All("/*", func(c *fiber.Ctx) error {
		if routeMap == nil {
			return errors.New("route map is unset")
		}

		route := c.Params("*")

		var backend string

		for regex, target := range routeMap {
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

func setupControlPlane(routeMapChan chan<- map[*regexp.Regexp]string) {
	controlPlane := fiber.New()

	controlPlane.Put("/routemap", func(c *fiber.Ctx) error {
		body := c.Body()

		var newRouteMapRaw map[string]string
		err := json.Unmarshal(body, &newRouteMapRaw)
		if err != nil {
			return err
		}

		newRouteMap := make(map[*regexp.Regexp]string)

		for route, backend := range newRouteMapRaw {
			regex, err := regexp.Compile(route)
			if err != nil {
				return err
			}
			newRouteMap[regex] = backend
		}

		routeMapChan <- newRouteMap

		return c.SendStatus(200)
	})

	go controlPlane.Listen(":10000")
}
