package main

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

type Movies struct {
	ID    int    `json:"id"`
	Name  string `json:"name" validate:"required,max=100"`
	Year  int    `json:"year" validate:"required"`
	Point int    `json:"point" validate:"required,gte=0,lte=100"`
}

var movies []Movies
var currentID = 1

func init() {
	movies = append(movies, Movies{ID: 1, Name: "The Shawshank Redemption", Year: 1994})
	currentID++
	movies = append(movies, Movies{ID: 2, Name: "The Godfather", Year: 1972})
	currentID++
	movies = append(movies, Movies{ID: 3, Name: "The Dark Knight", Year: 2008})
	currentID++
	movies = append(movies, Movies{ID: 4, Name: "Pulp Fiction", Year: 1994})
	currentID++
	movies = append(movies, Movies{ID: 5, Name: "Inception", Year: 2010})
	currentID++
	movies = append(movies, Movies{ID: 6, Name: "Fight Club", Year: 1999})
	currentID++
	movies = append(movies, Movies{ID: 7, Name: "Forrest Gump", Year: 1994})
	currentID++
	movies = append(movies, Movies{ID: 8, Name: "The Matrix", Year: 1999})
	currentID++
	movies = append(movies, Movies{ID: 9, Name: "Gladiator", Year: 2000})
	currentID++
	movies = append(movies, Movies{ID: 10, Name: "Interstellar", Year: 2014})
	currentID++
}

func main() {
	app := fiber.New()

	app.Route("/", func(router fiber.Router) {
		app.Get("/movies", getMovies)
		app.Get("/movie/:id", getMovieByID)
		app.Post("/movie", createMovie)
		app.Delete("/movie/:id", deleteMovie)
		app.Patch("/movie/:id", updateMovie)
	})
	app.Listen(":8080")
}

func getMovies(c *fiber.Ctx) error {
	return c.JSON(movies)
}

func getMovieByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	for _, movie := range movies {
		if movie.ID == id {
			return c.JSON(movie)
		}
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{})
}

func createMovie(c *fiber.Ctx) error {
	var newMovie Movies
	if err := c.BodyParser(&newMovie); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	if err := validate.Struct(newMovie); err != nil {
		errors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	newMovie.ID = currentID
	currentID++

	movies = append(movies, newMovie)

	return c.JSON(newMovie)
}

func deleteMovie(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}
	for i, movie := range movies {
		if movie.ID == id {
			movies = append(movies[:i], movies[i+1:]...)
			return c.JSON(fiber.Map{"message": "Kullanıcı başarıyla silindi"})
		}
	}
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{})
}

func updateMovie(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}

	var updatedMovie Movies
	if err := c.BodyParser(&updatedMovie); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}

	if err := validate.Struct(updatedMovie); err != nil {
		errors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			errors[err.Field()] = err.Tag()
		}
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	for i, movie := range movies {
		if movie.ID == id {
			movies[i] = updatedMovie
			return c.JSON(updatedMovie)
		}
	}
	return c.JSON(updatedMovie)

}
