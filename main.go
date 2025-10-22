package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "swagger-database-movie/docs"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var validate = validator.New()
var db *gorm.DB
var err error

type Movies struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" validate:"required,max=100"`
	Year  int    `json:"year" validate:"required"`
	Point int    `json:"point" validate:"required,gte=0,lte=100"`
}

func main() {
	db, err = gorm.Open(sqlite.Open("movies.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	if err := db.AutoMigrate(&Movies{}); err != nil {
		log.Fatal("Tablo oluşturulamadı:", err)
	}
	seedMovies()

	app := fiber.New()

	app.Route("/", func(router fiber.Router) {
		router.Get("movies", getMovies)
		router.Get("movie/:id", getMovieByID)
		router.Post("movie", createMovie)
		router.Delete("movie/:id", deleteMovie)
		router.Patch("movie/:id", updateMovie)
	})

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Println("Server hata:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Sunucu kapatılıyor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Sunucu kapatılamadı: %v", err)
	}

	log.Println("Sunucu kapatıldı")
}

func seedMovies() {
	var count int64
	db.Model(&Movies{}).Count(&count)
	if count > 0 {
		return
	}

	initial := []Movies{
		{Name: "Inception", Year: 2010, Point: 87},
		{Name: "The Matrix", Year: 1999, Point: 88},
		{Name: "Interstellar", Year: 2014, Point: 86},
		{Name: "The Godfather", Year: 1972, Point: 98},
		{Name: "The Dark Knight", Year: 2008, Point: 94},
		{Name: "Pulp Fiction", Year: 1994, Point: 92},
		{Name: "Fight Club", Year: 1999, Point: 88},
		{Name: "Forrest Gump", Year: 1994, Point: 89},
		{Name: "The Shawshank Redemption", Year: 1994, Point: 99},
		{Name: "Gladiator", Year: 2000, Point: 85},
	}

	for _, m := range initial {
		if err := db.Create(&m).Error; err != nil {
			log.Println("Seed eklenirken hata:", err)
		}
	}
}

// getMovies godoc
// @Summary      Tüm filmleri getir
// @Description  Veritabanındaki tüm filmleri listeler
// @Tags         movies
// @Produce      json
// @Success      200  {array}   Movies
// @Failure      500  {object}  map[string]string
// @Router       /movies [get]
func getMovies(c *fiber.Ctx) error {
	var movies []Movies
	if err := db.Find(&movies).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Veritabanı hatası"})
	}
	return c.JSON(movies)
}

// getMovieByID godoc
// @Summary      ID'ye göre film getir
// @Description  Verilen ID'li filmi döner
// @Tags         movies
// @Produce      json
// @Param        id   path      int  true  "Film ID"
// @Success      200  {object}  Movies
// @Failure      404  {object}  map[string]string
// @Router       /movie/{id} [get]
func getMovieByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var movie Movies
	if err := db.First(&movie, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Film bulunamadı"})
	}
	return c.JSON(movie)
}

// createMovie godoc
// @Summary      Yeni film ekle
// @Description  Gönderilen veriyle yeni film oluşturur
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        movie  body      Movies  true  "Film nesnesi"
// @Success      201  {object}  Movies
// @Failure      400  {object}  map[string]string
// @Router       /movie [post]
func createMovie(c *fiber.Ctx) error {
	var newMovie Movies
	if err := c.BodyParser(&newMovie); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz istek"})
	}

	if err := validate.Struct(newMovie); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"validation_error": err.Error()})
	}

	if err := db.Create(&newMovie).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Veritabanına eklenemedi"})
	}
	return c.Status(fiber.StatusCreated).JSON(newMovie)
}

// deleteMovie godoc
// @Summary      Film sil
// @Description  Verilen ID'li filmi siler
// @Tags         movies
// @Param        id   path  int  true  "Film ID"
// @Success      204  "No Content"
// @Failure      500  {object}  map[string]string
// @Router       /movie/{id} [delete]
func deleteMovie(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&Movies{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Silinemedi"})
	}
	return c.JSON(fiber.Map{"message": "Film silindi"})
}

// updateMovie godoc
// @Summary      Film güncelle
// @Description  Verilen ID'li filmi günceller
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id     path      int     true  "Film ID"
// @Param        movie  body      Movies  true  "Güncellenmiş film"
// @Success      200  {object}  Movies
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /movie/{id} [patch]
func updateMovie(c *fiber.Ctx) error {
	id := c.Params("id")
	var movie Movies
	if err := db.First(&movie, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Film bulunamadı"})
	}

	var updated Movies
	if err := c.BodyParser(&updated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Geçersiz veri"})
	}

	if err := validate.Struct(updated); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"validation_error": err.Error()})
	}

	movie.Name = updated.Name
	movie.Year = updated.Year
	movie.Point = updated.Point

	db.Save(&movie)
	return c.JSON(movie)
}
