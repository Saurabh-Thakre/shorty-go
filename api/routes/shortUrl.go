package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Saurabh-Thakre/shorty-go/database"
	"github.com/Saurabh-Thakre/shorty-go/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	// check for the incoming request body
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	// check if the input is an actual URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid URL",
		})
	}

	// check for the domain error
	// users may abuse the shortener by shorting the domain `localhost:3000` itself
	// leading to a inifite loop, so don't accept the domain for shortening
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "haha... nice try",
		})
	}

	// enforce https
	// all url will be converted to https before storing in database
	body.URL = helpers.EnforceHTTP(body.URL)

	r := database.CreateClient(2)
	defer r.Close()

	// check if the input URL already exists in the database
	val, err := r.Get(database.Ctx, body.URL).Result()
	if err != nil {
		fmt.Printf("Error getting value from Redis: %v\n", err)
	} else {
		fmt.Printf("Value for key %v: %v\n", body.URL, val)
	}

	if err == nil && val != "" {
		// input URL already exists in database, return the existing shortened URL
		resp := response{
			URL:             body.URL,
			CustomShort:     os.Getenv("DOMAIN") + "/" + val,
			Expiry:          24, // default expiry of 24 hours
			XRateRemaining:  10,
			XRateLimitReset: 30,
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}

	// check if the user has provided any custom short urls
	// if yes, proceed,
	// else, create a new short using the first 6 digits of uuid
	// haven't performed any collision checks on this
	// you can create one for your own
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// check if the user provided short is already in use
	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL short already in use",
		})
	}

	// add the input URL and corresponding shortened URL to the database
	err = r.Set(database.Ctx, id, body.URL, 24*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connect to server",
		})
	}
	// add the input URL and shortened URL to the database with the URL as key
	r.Set(database.Ctx, body.URL, id, 24*3600*time.Second)

	// respond with the url, short, expiry in hours, calls remaining and time to reset
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          24, // default expiry of 24 hours
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2 := database.CreateClient(1)
	defer r2.Close()
	r2.Decr(database.Ctx, c.IP())
	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)
}
