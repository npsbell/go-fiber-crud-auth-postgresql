package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Password string
}

func createUser(db *gorm.DB, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost) //แปลง password ธรรมดา → hash

	if err != nil { //ถ้า hash ไม่ได้ → ส่ง error กลับ
		return err
	}

	user.Password = string(hashedPassword) //เอา hash ไปแทน password เดิม

	result := db.Create(user) //INSERT ลง database
	if result.Error != nil {  //ถ้า DB error → ส่งกลับ
		return result.Error
	}
	return nil
}

func loginUser(db *gorm.DB, user *User) (string, error) {
	selectedUser := new(User)                                       //สร้าง struct ว่างไว้รับข้อมูลจาก DB
	result := db.Where("email = ?", user.Email).First(selectedUser) //SELECT user จาก email

	if result.Error != nil { //ถ้าไม่เจอ → login ไม่ผ่าน
		return "", result.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(user.Password)) //เทียบ:hash ใน DB password ที่ user กรอก

	if err != nil { //ถ้าไม่ตรง → login fail
		return "", err
	}

	//สร้าง JWT
	jwtsecretkey := os.Getenv("JWT_SECRET")               //key ลับสำหรับเซ็น token
	token := jwt.New(jwt.SigningMethodHS256)              //สร้าง token ใหม่
	claims := token.Claims.(jwt.MapClaims)                //ดึง payload ภายใน token
	claims["user_id"] = selectedUser.ID                   //ใส่ user id
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() //หมดอายุใน 72 ชม.

	// เซ็น token
	t, err := token.SignedString([]byte(jwtsecretkey)) //เอา token + secret → กลายเป็น string
	if err != nil {                                    //ถ้าเซ็นพัง → fail
		return "", err
	}

	return t, nil
}
