package main

import (
	"log"

	"gorm.io/gorm"
)

type Book struct {
	gorm.Model         //มี field: ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Price       uint   `json:"price"`
}

func getBook(db *gorm.DB, id int) (*Book, error) {
	var book Book
	result := db.First(&book, id) //SELECT * FROM books WHERE id = ?

	if result.Error != nil {
		log.Fatalf("Error get book: %v", result.Error)
		return nil, result.Error
	}

	return &book, nil
}

func createBook(db *gorm.DB, book *Book) error {
	result := db.Create(book) //INSERT

	if result.Error != nil {
		return result.Error
	}

	// fmt.Println("Successfully created book")
	return nil
}

func updateBook(db *gorm.DB, book *Book) error {
	result := db.Save(book) //UPDATE ทั้งแถว

	if result.Error != nil {
		// log.Fatalf("Error updating book: %v", result.Error)
		return result.Error
	}

	// fmt.Println("Successfully updated book")
	return nil
}

// update เฉพาะ fields ที่ต้องการ
func updateFieldBook(db *gorm.DB, book *Book) error {
	result := db.Model(&book).Updates(book)

	if result.Error != nil {
		// log.Fatalf("Error updating book: %v", result.Error)
		return result.Error
	}

	// fmt.Println("Successfully updated book")
	return nil
}

func deleteBook(db *gorm.DB, id uint) error {
	var book Book
	result := db.Delete(&book, id) //DELETE FROM books WHERE id = ?

	if result.Error != nil {
		// log.Fatalf("Error deleting book: %v", result.Error)
		return result.Error
	}

	// fmt.Println("Successfully deleted book")
	return nil

}

func getBooks(db *gorm.DB) []Book {
	var books []Book

	result := db.Find(&books) //SELECT * FROM books

	if result.Error != nil {
		log.Fatalf("Error get book: %v", result.Error)
	}

	return books
}

// onebook
func searchBook(db *gorm.DB, bookName string) (*Book, error) {
	var book Book

	result := db.Where("name = ?", bookName).First(&book)
	//WHERE name = 'ค่าที่อยู่ในตัวแปร bookName' เครื่องหมาย ? = placeholder GORM จะเอา bookName ไปใส่แทนให้อัตโนมัติ (กัน SQL injection)
	//First(&book) เอา เล่มแรกเล่มเดียว ที่ตรงเงื่อนไข SQL จริง ๆ SELECT * FROM books WHERE name = 'Go 101' ORDER BY id LIMIT 1;

	if result.Error != nil {
		log.Fatalf("Error searching book: %v", result.Error)
	}

	return &book, nil
}

// morebooks
func searchBooks(db *gorm.DB, bookName string) ([]Book, error) {
	var books []Book

	result := db.Where("name = ?", bookName).Find(&books)
	//Find(&books) เอา ทุกเล่มที่ตรงเงื่อนไข SQL จริง ๆ SELECT * FROM books WHERE name = 'Go 101';

	if result.Error != nil {
		log.Fatalf("Error searching book: %v", result.Error)
	}

	return books, nil
}

// sorting
func sortBooks(db *gorm.DB, bookName string) ([]Book, error) {
	var books []Book

	result := db.Where("name = ?", bookName).Order("price desc").Find(&books)
	//เอาทุกเล่มที่ตรงเงื่อนไข แล้ว เรียงราคาจากมาก → น้อย SQL จริง ๆ SELECT * FROM books WHERE name = 'Go 101' ORDER BY price DESC;

	if result.Error != nil {
		log.Fatalf("Error searching book: %v", result.Error)
	}

	return books, nil
}
