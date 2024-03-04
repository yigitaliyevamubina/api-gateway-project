package models

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-ozzo/ozzo-validation/v3/is"
)

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type GetAllUsersRespModel struct {
	Users []*User
	Count int64
}

// Login model
type LoginReqModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserModel struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	BirthDate   string `json:"birth_date"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	AccessToken string `json:"access_token"`
}

type LoginRespModel struct {
	Result bool      `json:"result"`
	User   UserModel `json:"user"`
}

func (r *User) Validate() error {
	return validation.ValidateStruct(
		r,
		validation.Field(&r.FirstName, validation.Required, validation.Length(3, 50), validation.Match(regexp.MustCompile("^[A-Z][a-z]*$")).Error("should start with a capital letter and should only contain letters")),
		validation.Field(&r.LastName, validation.Required, validation.Length(3, 50), validation.Match(regexp.MustCompile("^[A-Z][a-z]*$")).Error("should start with a capital letter and should only contain letters")),
		validation.Field(&r.Email, validation.Required, validation.Length(5, 100), is.Email),
		validation.Field(&r.Password,
			validation.Required,
			validation.Length(5, 30),
			validation.Match(regexp.MustCompile("\\d")).Error("should contain at least one digit"),
			validation.Match(regexp.MustCompile("^[a-zA-Z\\d]+$")).Error("should only contain letters (either lowercase or uppercase) and digits"),
		),
		validation.Field(&r.BirthDate, validation.Required, validation.Match(regexp.MustCompile("^\\d{4}-\\d{2}-\\d{2}$")).Error("should be in the format 'yyyy-mm-dd'")),
	)
}

type RegisterUserModel struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Code      string `json:"code"`
}

type VerifyRespModel struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	BirthDate   string `json:"birth_date"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	AccessToken string `json:"access_token"`
}

type RegisterRespModel struct {
	Message string `json:"message"`
}

type Status struct {
	Message string `json:"message"`
}

type UserResp struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListUsersResp struct {
	Count int64
	Users []*UserResp
}

type ChangePasswordReq struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
}

func (c *ChangePasswordReq) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.NewPassword, validation.Required, validation.Length(5, 30),
			validation.Match(regexp.MustCompile("\\d")).Error("should contain at least one digit"),
			validation.Match(regexp.MustCompile("^[a-zA-Z\\d]+$")).Error("should only contain letters (either lowercase or uppercase) and digits"),
		),
	)
}
