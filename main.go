package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Middleware: authRequired
func authRequired(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")              //ดึง cookie ชื่อ jwt จาก request
	jwtSecretKey := os.Getenv("JWT_SECRET") //secret key สำหรับถอด JWT
	fmt.Println("JWT:", os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) { //พยายามแปลง string → JWT object
		return []byte(jwtSecretKey), nil //บอก library ว่า key คืออะไร
	})

	if err != nil || !token.Valid { //ถ้า token พัง → 401
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	claim := token.Claims.(jwt.MapClaims) //ดึง payload ข้างใน token ออกมา
	fmt.Println(claim)
	// fmt.Println(claim["user_id"])

	return c.Next() //อนุญาตให้ request วิ่งต่อไป
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Connection string, สร้าง DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	newLogger := logger.New( //สร้าง logger
		log.New(os.Stdout, "\r\n", log.LstdFlags), // log ลง console
		logger.Config{ //ตั้งค่า log
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)

	//connect DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{ //เปิด connection
		Logger: newLogger, //ใช้ logger
	})

	if err != nil { //ถ้าต่อไม่ได้ → โปรแกรมตาย
		panic("failed to connect database")
	}

	//migrate table
	db.AutoMigrate(&Book{}, &User{}) //สร้าง table books และ users อัตโนมัติ

	app := fiber.New()              //สร้าง web server
	app.Use("/books", authRequired) //ทุก route ที่ขึ้นต้น /books ต้องผ่าน auth

	app.Get("/books", func(c *fiber.Ctx) error { //สร้าง endpoint
		return c.JSON(getBooks(db)) //เรียก DB แล้วแปลงเป็น JSON
	})

	app.Get("/books/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id") //ดึง id จาก URL
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book, err := getBook(db, id) //query DB
		return c.JSON(book)          //ส่ง JSON กลับ
	})

	app.Post("/books", func(c *fiber.Ctx) error {
		book := new(Book) //สร้าง struct ว่าง

		if err := c.BodyParser(book); err != nil { //แปลง JSON → struct
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err := createBook(db, book) //insert DB

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{
			"message": "Book created successfully",
		})
	})

	app.Put("/books/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book := new(Book)

		if err := c.BodyParser(book); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		book.ID = uint(id) //บอกว่าแก้ record ไหน

		err = updateBook(db, book) //save ลง DB

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{
			"message": "Book updated successfully",
		})
	})

	app.Delete("/books/:id", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err = deleteBook(db, uint(id)) //ลบ record

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{
			"message": "Book deleted successfully",
		})
	})

	app.Post("/register", func(c *fiber.Ctx) error {
		user := new(User)

		if err := c.BodyParser(user); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		err := createUser(db, user) //สร้าง user ใหม่ + hash password

		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.JSON(fiber.Map{
			"message": "User registered successfully",
		})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		user := new(User)

		if err := c.BodyParser(user); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		token, err := loginUser(db, user) //เช็ค user + สร้าง JWT

		if err != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		c.Cookie(&fiber.Cookie{ //เก็บ token ลง cookie
			Name:     "jwt",
			Value:    token,
			Expires:  time.Now().Add(time.Hour * 72), //cookie อยู่ 3 วัน และ JS อ่านไม่ได้
			HTTPOnly: true,
		})

		return c.JSON(fiber.Map{
			// "token": token,
			"message": "User logged in successfully",
		})
	})

	app.Listen(":8080")
}
