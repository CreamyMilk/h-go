package main

import (
	"fmt"
	"hosigo/notifiers"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog/log"
)

type Patient struct {
	Reference int    `json:"reference"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Served    bool   `json:"served"`
}

const defaultPort = "8081"

var patients []Patient

func main() {
	// utils.LoadEnvironmentVariables()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	app := fiber.New(fiber.Config{
		Prefork: false,
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Post("/registerUser", handleRegister)
	app.Post("/serveUser", handleServe)
	app.Get("/listUsers", handleList)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(418).JSON(&fiber.Map{
			"Message": "🍏 Route not found",
		}) // => 418 "I am a tepot"
	})

	log.Fatal().Err(app.Listen(":" + port))
}

type NewPatient struct {
	Email    string `json:"email" validate:"required"`
	Username string `json:"username" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
}

func handleRegister(c *fiber.Ctx) error {
	r := new(NewPatient)
	err := c.BodyParser(r)
	fmt.Println(r)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Wrong data format")
	}

	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var errorMsg string
		for _, e := range validationErrors {
			errorMsg = fmt.Sprintf("%s is required", e.Field())
		}
		return c.Status(fiber.StatusBadRequest).SendString(errorMsg)
	}

	currentNumber := len(patients)

	newPatient := Patient{
		Reference: currentNumber,
		Phone:     r.Phone,
		Username:  r.Username,
		Email:     r.Email,
		Served:    false,
	}

	patients = append(patients, newPatient)

	go sendQueueMessage(newPatient, false)

	for _, v := range patients {
		if !v.Served {
			return c.JSON(v)
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func handleServe(c *fiber.Ctx) error {
	number, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Patient Id not found")
	}

	var cui int
	found := false
	for i, v := range patients {
		if v.Reference == number {
			cui = i
			found = true
		}
	}

	if !found {
		return c.Status(fiber.StatusNotAcceptable).SendString("Patient not found")
	}
	if patients[cui].Served {
		return c.Status(fiber.StatusNotAcceptable).SendString("Patient already served")
	}

	patients[cui].Served = true

	// if len(patients) > cui {
	// 	return c.SendStatus(fiber.StatusAccepted)
	// }

	var patientNext Patient

	for _, v := range patients {
		if !v.Served {
			patientNext = v
		}
	}

	go sendQueueMessage(patientNext, true)

	return c.JSON(patientNext)
}

func handleList(c *fiber.Ctx) error {
	return c.JSON(patients)
}

func sendQueueMessage(user Patient, next bool) error {
	if user.Phone == "" {
		return nil
	}

	var message string
	if next {
		message = fmt.Sprintf("Hello %s, it's your turn to be served. Your reference number is %v", user.Username, user.Reference)
	} else {
		message = fmt.Sprintf("Hello %s, you have been assigned NO: \"%v\". \nWill text you when it's your turn to get served", user.Username, user.Reference)
	}
	fmt.Println("🚀 sending " + message)

	// Replace with your SMS sending code
	err := notifiers.SendSms(message, user.Phone, notifiers.RecipientAlert)
	if err != nil {
		log.Print("Errrr 🪽", err.Error())
		return fmt.Errorf("❌ issue with message sending: %s", err.Error())
	}
	log.Print("Sent 🪽")

	return nil
}
