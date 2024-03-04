package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"myproject/api-gateway/api/handlers/tokens"
	"myproject/api-gateway/api/models"
	"myproject/api-gateway/email"
	pbu "myproject/api-gateway/genproto/user-service"
	"myproject/api-gateway/pkg/etc"
	"net/http"
	"strings"
	"time"
)

// Register User
// @Router /v1/register [post]
// @Summary register user
// @Tags User
// @Description Registration
// @Accept json
// @Produce json
// @Param UserData body models.User true "Register user"
// @Success 201 {object} models.RegisterRespModel
// @Failure 400 string error models.ResponseError
// @Failure 500 string error models.ResponseError
func (h *handlerV1) Register(c *gin.Context) {
	var (
		body       models.User
		code       string
		jspMarshal protojson.MarshalOptions
	)

	jspMarshal.UseProtoNames = true

	err := c.BindJSON(&body)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidJSON) {
		return
	}

	body.ID = uuid.New().String()
	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	err = body.Validate()
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorValidationError) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	exists, err := h.serviceManager.UserService().CheckField(ctx, &pbu.CheckFieldReq{
		Value: body.Email,
		Field: "email",
	})

	if handleInternalServerErrorWithMessage(c, h.log, err, "failed to check email uniqueness") {
		return
	}

	if exists.Status {
		if handleBadRequestErrWithMessage(c, h.log, fmt.Errorf("you've already registered before, try to log in"), ErrorBadRequest) {
			return
		}
	}

	code = etc.GenerateCode(5)
	registerUser := models.RegisterUserModel{
		ID:        body.ID,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		BirthDate: body.BirthDate,
		Email:     body.Email,
		Password:  body.Password,
		Code:      code,
	}

	userJson, err := json.Marshal(registerUser)
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while marshaling json") {
		return
	}

	timeOut := time.Second * 300

	err = h.inMemoryStorage.SetWithTTL(registerUser.Email, string(userJson), int(timeOut.Seconds()))
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while setting with ttl to redis") {
		return
	}

	message, err := email.SendVerificationCode(email.EmailPayload{
		From:     h.cfg.SendEmailFrom,
		To:       registerUser.Email,
		Password: h.cfg.EmailCode,
		Code:     registerUser.Code,
		Message:  fmt.Sprintf("Hi, %s", registerUser.FirstName),
	})
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while sending code to user's email") {
		return
	}

	c.JSON(http.StatusOK, models.RegisterRespModel{
		Message: message,
	})
}

// Verify User
// @Router /v1/verify/{email}/{code} [get]
// @Summary verify user
// @Tags User
// @Description Verify a user with code sent to their email
// @Accept json
// @Product json
// @Param email path string true "email"
// @Param code path string true "code"
// @Success 201 {object} models.VerifyRespModel
// @Failure 400 string error models.ResponseError
// @Failure 400 string error models.ResponseError
func (h *handlerV1) Verify(c *gin.Context) {
	var jspMarshal protojson.MarshalOptions
	jspMarshal.UseProtoNames = true

	userEmail := c.Param("email")
	userCode := c.Param("code")

	registeredUser, err := redis.Bytes(h.inMemoryStorage.Get(userEmail))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.StandardErrorModel{
			Status:  ErrorCodeNotFound,
			Message: "Code is expired, try again",
		})
		h.log.Error("Code is expired, TTL is over.")
		return
	}

	var user models.RegisterUserModel
	if err := json.Unmarshal(registeredUser, &user); err != nil {
		if handleInternalServerErrorWithMessage(c, h.log, err, "cannot unmarshal user from redis") {
			return
		}
	}

	if user.Code != userCode {
		if handleBadRequestErrWithMessage(c, h.log, fmt.Errorf("code is incorrect, verification is failed"), ErrorCodeInvalidCode) {
			return
		}
	}

	user.Password, err = etc.GenerateHashPassword(user.Password)
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while hashing the password") {
		return
	}

	h.jwtHandler = tokens.JWTHandler{
		Sub:       user.ID,
		Role:      "user",
		SignInKey: h.cfg.SignInKey,
		Log:       h.log,
		TimeOut:   h.cfg.AccessTokenTimeOut,
	}

	access, refresh, err := h.jwtHandler.GenerateAuthJWT()
	if handleInternalServerErrorWithMessage(c, h.log, err, "error generating access and refresh tokens") {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	respUser, err := h.serviceManager.UserService().CreateUser(ctx, &pbu.User{
		Id:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		BirthDate:    user.BirthDate,
		Email:        user.Email,
		Password:     user.Password,
		AccessToken:  access,
		RefreshToken: refresh,
	})
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while creating user") {
		return
	}
	userModel := models.VerifyRespModel{
		ID:          respUser.Id,
		FirstName:   respUser.FirstName,
		LastName:    respUser.LastName,
		BirthDate:   respUser.BirthDate,
		Email:       respUser.Email,
		Password:    respUser.Password,
		AccessToken: respUser.AccessToken,
	}

	c.JSON(http.StatusCreated, userModel)
}

// Login User
// @Summary login user
// @Tags User
// @Description Login
// @Accept json
// @Produce json
// @Param User body models.LoginReqModel true "Login"
// @Success 201 {object} models.LoginRespModel
// @Failure 400 string Error models.ResponseError
// @Failure 400 string Error models.ResponseError
// @Router /v1/login [post]
func (h *handlerV1) Login(c *gin.Context) {
	var (
		jspMarshal protojson.MarshalOptions
		body       models.LoginReqModel
	)

	jspMarshal.UseProtoNames = true
	err := c.ShouldBind(&body)
	if handleBadRequestErrWithMessage(c, h.log, err, "error while marshaling request body") {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	user, err := h.serviceManager.UserService().IfExists(ctx, &pbu.IfExistsReq{
		Email: body.Email,
	})

	if handleInternalServerErrorWithMessage(c, h.log, err, "error while checking if user exists") {
		return
	}

	if !etc.CompareHashPassword(user.User.Password, body.Password) {
		if handleBadRequestErrWithMessage(c, h.log, fmt.Errorf("wrong password"), ErrorInvalidCredentials) {
			return
		}
	}

	h.jwtHandler = tokens.JWTHandler{
		Sub:       user.User.Id,
		Role:      "user",
		SignInKey: h.cfg.SignInKey,
		Log:       h.log,
		TimeOut:   h.cfg.AccessTokenTimeOut,
	}

	access, _, err := h.jwtHandler.GenerateAuthJWT()
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while generating access and refresh token") {
		return
	}

	loginResp := models.UserModel{
		ID:          user.User.Id,
		FirstName:   user.User.FirstName,
		LastName:    user.User.LastName,
		BirthDate:   user.User.BirthDate,
		Email:       user.User.Email,
		Password:    user.User.Password,
		AccessToken: access,
	}

	c.JSON(http.StatusOK, models.LoginRespModel{
		Result: true,
		User:   loginResp,
	})
}

// CreateUser
// @Router /v1/user/create [post]
// @Security BearerAuth
// @Summary create user
// @Tags User
// @Description Create a new user with the provided details
// @Accept json
// @Produce json
// @Param UserInfo body models.User true "Create user"
// @Success 201 {object} models.UserModel
// @Failure 400 string Error models.ResponseError
// @Failure 500 string Error models.ResponseError
func (h *handlerV1) CreateUser(c *gin.Context) {
	var (
		body       models.User
		jspMarshal protojson.MarshalOptions
	)

	jspMarshal.UseProtoNames = true
	err := c.ShouldBindJSON(&body)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidJSON) {
		return
	}

	body.Password, err = etc.GenerateHashPassword(body.Password)
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while hashing password") {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	body.ID = uuid.New().String()
	access, refresh, err := h.jwtHandler.GenerateAuthJWT()
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while generating access and refresh token") {
		return
	}

	createReq := &pbu.User{
		Id:           body.ID,
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		BirthDate:    body.BirthDate,
		Email:        body.Email,
		Password:     body.Password,
		AccessToken:  access,
		RefreshToken: refresh,
	}

	respUser, err := h.serviceManager.UserService().CreateUser(ctx, createReq)
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while creating a user") {
		return
	}

	response := models.UserModel{
		ID:          respUser.Id,
		FirstName:   respUser.FirstName,
		LastName:    respUser.LastName,
		BirthDate:   respUser.BirthDate,
		Email:       respUser.Email,
		Password:    respUser.Password,
		AccessToken: respUser.AccessToken,
	}

	c.JSON(http.StatusCreated, response)
}

// Get User By Id
// @Router /v1/user/{id} [get]
// @Security BearerAuth
// @Summary get user by id
// @Tags User
// @Description Get user
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 201 {object} models.User
// @Failure 400 string Error models.ResponseError
// @Failure 500 string Error models.ResponseError
func (h *handlerV1) GetUserById(c *gin.Context) {
	var jspMarshal protojson.MarshalOptions
	jspMarshal.UseProtoNames = true

	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration((h.cfg.CtxTimeout)))
	defer cancel()

	respUser, err := h.serviceManager.UserService().GetUserById(ctx, &pbu.GetUserReqById{
		UserId: id,
	})
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while getting user by id") {
		return
	}

	response := models.User{
		ID:        respUser.Id,
		FirstName: respUser.FirstName,
		LastName:  respUser.LastName,
		BirthDate: respUser.BirthDate,
		Email:     respUser.Email,
		Password:  respUser.Password,
	}

	c.JSON(http.StatusOK, response)
}

// Update User
// @Router /v1/user/update/{id} [put]
// @Security BearerAuth
// @Summary update user
// @Tags User
// @Description Update user
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param UserInfo body models.User true "Update User"
// @Success 201 {object} models.User
// @Failure 400 string Error models.ResponseError
// @Failure 500 string Error models.ResponseError
func (h *handlerV1) UpdateUser(c *gin.Context) {
	var (
		body        models.User
		jspbMarshal protojson.MarshalOptions
	)

	jspbMarshal.UseProtoNames = true

	err := c.ShouldBindJSON(&body)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidJSON) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	id := c.Param("id")

	respUser, err := h.serviceManager.UserService().UpdateUser(ctx, &pbu.User{
		Id:        id,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		BirthDate: body.BirthDate,
		Email:     body.Email,
		Password:  body.Password,
	})

	response := models.User{
		ID:        respUser.Id,
		FirstName: respUser.FirstName,
		LastName:  respUser.LastName,
		BirthDate: respUser.BirthDate,
		Email:     respUser.Email,
		Password:  respUser.Password,
	}

	c.JSON(http.StatusOK, response)
}

// Delete User
// @Router /v1/user/delete/{id} [delete]
// @Security BearerAuth
// @Summary delete user
// @Tags User
// @Description Delete user
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 201 {object} models.Status
// @Failure 400 string Error models.ResponseError
// @Failure 500 string Error models.ResponseError
func (h *handlerV1) DeleteUser(c *gin.Context) {
	var jspbMarshal protojson.MarshalOptions
	jspbMarshal.UseProtoNames = true

	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	_, err := h.serviceManager.UserService().DeleteUser(ctx, &pbu.DeleteUserReq{
		UserId: id,
	})

	if handleInternalServerErrorWithMessage(c, h.log, err, "error while deleting user") {
		return
	}

	c.JSON(http.StatusOK, models.Status{Message: "user was successfully deleted"})
}

// List users
// @Router /v1/users/{page}/{limit}/{filter} [get]
// @Security BearerAuth
// @Summary get users' list
// @Tags User
// @Description get all users
// @Accept json
// @Produce json
// @Param page path string false "page"
// @Param limit path string false "limit"
// @Param filter path string false "filter"
// @Success 201 {object} models.ListUsersResp
// @Failure 400 string Error models.ResponseError
// @Failure 500 string Error models.ResponseError
func (h *handlerV1) ListUsers(c *gin.Context) {
	var jspbMarshal protojson.MarshalOptions
	jspbMarshal.UseProtoNames = true

	page, err := ParsePageQueryParam(c)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidParams) {
		return
	}
	limit, err := ParseLimitQueryParam(c)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidParams) {
		return
	}
	filter := c.Param("filter")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()
	response, err := h.serviceManager.UserService().GetAllUsers(ctx, &pbu.ListUsersReq{
		Limit:  int64(limit),
		Page:   int64(page),
		Filter: filter,
	})
	if handleInternalServerErrorWithMessage(c, h.log, err, ErrorCodeInternalServerError) {
		return
	}

	c.JSON(http.StatusOK, response)
}

// Change password
// @Router /v1/user/password/change [post]
// @Security BearerAuth
// @Summary change password
// @Tags User
// @Description Change password
// @Accept json
// @Produce json
// @Param Change-password body models.ChangePasswordReq true "Change password"
// @Success 201 {object} models.Status
// @Failure 400 string Error models.ResponseError
// @Failure 400 string Error models.ResponseError
func (h *handlerV1) ChangePassword(c *gin.Context) {
	var (
		body        models.ChangePasswordReq
		jspbMarshal protojson.MarshalOptions
	)

	jspbMarshal.UseProtoNames = true

	err := c.ShouldBindJSON(&body)
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorCodeInvalidJSON) {
		return
	}

	err = body.Validate()
	if handleBadRequestErrWithMessage(c, h.log, err, ErrorValidationError) {
		return
	}

	body.NewPassword, err = etc.GenerateHashPassword(body.NewPassword)
	if handleInternalServerErrorWithMessage(c, h.log, err, "error while hashing the password") {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.cfg.CtxTimeout))
	defer cancel()

	result, err := h.serviceManager.UserService().ChangePassword(ctx, &pbu.ChangeUserPasswordReq{
		Email:    body.Email,
		Password: body.NewPassword,
	})

	if handleInternalServerErrorWithMessage(c, h.log, err, "error while changing password") {
		return
	}

	if result.Status {
		c.JSON(http.StatusOK, models.Status{Message: "password was successfully changed."})
	} else {
		c.JSON(http.StatusOK, models.Status{Message: "failed to change the password"})
	}
}
