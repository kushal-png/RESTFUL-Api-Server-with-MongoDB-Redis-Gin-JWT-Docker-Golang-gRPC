package grpcApi

import (
	"context"
	"errors"
	"fmt"
	"project/initializers"
	models "project/model"
	"project/pb"
	services "project/service"
	"project/utils"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	AuthService services.AuthService
	UserService services.UserService
}

func NewGrpcAuthServer(authService services.AuthService, userService services.UserService) *AuthServer {
	return &AuthServer{
		AuthService: authService,
		UserService: userService,
	}
}

func (a *AuthServer) SignUpUser(ctx context.Context, req *pb.SignUpInput) (*pb.Response, error) {
	newUser := &models.SignUpInput{
		Name:            req.GetName(),
		Email:           req.GetEmail(),
		Password:        req.GetPassword(),
		PasswordConfirm: req.GetPasswordConfirm(),
	}
	fmt.Println(req.GetName(), req.GetEmail(), req.GetPassword(), req.GetPasswordConfirm())

	code := utils.Encode(utils.GenerateRandomString(20))
	newUser.VerificationCode = code

	res, err := a.AuthService.SignUpUser(newUser)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	//send email
	emailData := utils.EmailData{
		Code:    code,
		Subject: "Your accout verification code",
	}
	err = utils.SendEmail(res, &emailData)
	if err != nil {
		return nil, errors.New("failed to send the email")
	}

	response := &pb.Response{
		Status:  "success",
		Message: "We have sent the verification mail",
	}

	return response, nil
}

func (a *AuthServer) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.Response, error) {
	code := req.GetVerificationCode()
	err := a.AuthService.VerifyUser(code)
	if err != nil {
		return nil, errors.New("failed to verify")
	}

	response := &pb.Response{
		Status:  "success",
		Message: "Verified",
	}

	return response, nil
}

func (a *AuthServer) SignInUser(ctx context.Context, req *pb.SignInInput) (*pb.TokenResponse, error) {
	payload := &models.SignInInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	fmt.Println(req.GetEmail(), req.GetPassword(), payload)

	res, err := a.UserService.GetUserByMail(payload.Email)
	if err != nil || res == nil {
		return nil, errors.New("no user with this email exists")
	}

	err = utils.VerifyPassword(payload.Password, res.Password)
	if err != nil {
		return nil, errors.New("password Incorrect")
	}

	if !res.Verified {
		return nil, errors.New("please verify the user")
	}

	config, _ := initializers.LoadConfig(".")
	access_token, err := utils.CreateToken(config.AccessTokenExpiresIn, res.ID, config.AccessTokenPrivateKey)
	if err != nil {
		return nil, errors.New("failed to create access token")
	}

	refresh_token, err := utils.CreateToken(config.RefreshTokenExpiresIn, res.ID, config.RefreshTokenPrivateKey)
	if err != nil {
		return nil, errors.New("failed to create refersh token")
	}

	response := &pb.TokenResponse{
		Status:       "success",
		Message:      "Verified",
		AccessToken:  access_token,
		RefreshToken: refresh_token,
	}

	return response, nil
}

func (a *AuthServer) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.Response, error) {
	fmt.Println("enter")
	payload := &models.ForgotPasswordInput{
		Email: req.GetEmail(),
	}

	res, err := a.UserService.GetUserByMail(payload.Email)
	if err != nil {
		return nil, errors.New("failed to find user")
	}

	if !res.Verified {
		return nil, errors.New("not verified")
	}

	code := utils.Encode(utils.GenerateRandomString(20))
	err = a.AuthService.ForgotPassword(res.Email, code)
	if err != nil {
		return nil, errors.New("failed to send reset code")
	}

	emailData := utils.EmailData{
		Code:    code,
		Subject: "Your accout reser password code",
	}
	err = utils.SendEmail(res, &emailData)
	if err != nil {
		return nil, errors.New("failed to send mail")
	}

	response := &pb.Response{
		Status:  "success",
		Message: "Code sent on email",
	}

	return response, nil
}

func (a *AuthServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.Response, error) {
	payload := &models.ResetPasswordInput{
		Password:        req.GetReq().Password,
		PasswordConfirm: req.GetReq().PasswordConfirm,
	}

	code := req.GetResetCode()

	if payload.Password != payload.PasswordConfirm {
		return nil, errors.New("passwords donnot match")
	}

	hashed_pw := utils.HashPassword(payload.Password)

	err := a.AuthService.ResetPassword(hashed_pw, code)
	if err != nil {
		return nil, errors.New("failed")
	}

	response := &pb.Response{
		Status:  "success",
		Message: "Changes Successfully",
	}

	return response, nil
}

func (a *AuthServer) PracticeChecker(ctx context.Context, in *pb.SignUpInput) (*pb.Response, error) {
	fmt.Println(in.GetEmail(), in.GetName(), in.GetPassword(), in.GetPasswordConfirm())
	response := &pb.Response{
		Status:  "success",
		Message: "Changes Successfully",
	}

	return response, nil
}
