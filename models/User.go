package models

import (
	"errors"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/jameslahm/conduit-server-gin/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// JwtClaims jwt claims
type JwtClaims struct {
	UserID string `json:"id"`
	jwt.StandardClaims
}

// User User struct
type User struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"-"`
	Email     string               `bson:"email,omitempty" json:"email"`
	Username  string               `bson:"username,omitempty" json:"username"`
	Password  string               `bson:"password,omitempty" json:"-"`
	Bio       string               `bson:"bio,omitempty" json:"bio"`
	Image     string               `bson:"image,omitempty" json:"image"`
	Token     string               `bson:"-" json:"token"`
	Following []primitive.ObjectID `bson:"following,omitempty" json:"following"`
	Favorites []primitive.ObjectID `bson:"favorites,omitempty" json:"favorites"`
}

// Profile Profile struct
type Profile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

// GenerateHashPassword generate passsword hash using bcrypt
func GenerateHashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Fatal("Error: Generate Password Hash Failed!")
	}
	return string(hash)
}

// VerifyPassword verify password
func VerifyPassword(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// GenerateJwtToken generate token
func GenerateJwtToken(ID primitive.ObjectID) (string, error) {
	claims := JwtClaims{
		ID.Hex(),
		jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "conduit",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	ss, err := token.SignedString([]byte(os.Getenv("SECRET")))
	return ss, err
}

// VerifyToken verify token
func VerifyToken(ss string) (*JwtClaims, error) {
	var claims *JwtClaims
	token, err := jwt.ParseWithClaims(ss, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid && (err == nil) {
		return claims, nil
	}
	return nil, errors.New("error: token invalid")
}

// ToProfile to profile
func (user *User) ToProfile(loginUser *User) Profile {
	var profile Profile
	profile.Username = user.Username
	profile.Bio = user.Bio
	profile.Image = user.Image
	profile.Following = false
	if loginUser != nil && !loginUser.ID.IsZero() {
		if utils.IndexOf(loginUser.Following, user.ID) != -1 {
			profile.Following = true
		}
	}
	return profile
}

// Follow follow
func (user *User) Follow(userToFollow *User) {
	user.Following = append(user.Following, userToFollow.ID)
}

// UnFollow unfollow
func (user *User) UnFollow(userToUnFollow *User) error {
	index := utils.IndexOf(user.Following, userToUnFollow.ID)
	if index == -1 {
		return errors.New("error: bad request")
	}
	user.Following = append(user.Following[0:index], user.Following[index+1:]...)
	return nil
}

// Favorite favorite
func (user *User) Favorite(article *Article) error {
	index := utils.IndexOf(user.Favorites, article.ID)
	if index == -1 {
		user.Favorites = append(user.Favorites, article.ID)
		return nil
	}
	return errors.New("error: already favorite")
}

// UnFavorite unfavorite
func (user *User) UnFavorite(article *Article) error {
	index := utils.IndexOf(user.Favorites, article.ID)
	if index == -1 {
		return errors.New("error: no favorite")
	}
	user.Favorites = append(user.Favorites[:index], user.Favorites[index+1:]...)
	return nil
}
