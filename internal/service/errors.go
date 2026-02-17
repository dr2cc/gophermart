package service

import "errors"

var (
	ErrOrderUploadedThisUser      = errors.New("order already uploaded this user") //200
	ErrOrderAcceptedForProcessing = errors.New("order accepted for processing")    //202
	ErrInvalidOrderFormat         = errors.New("invalid order format")             //400
	ErrUserNotAuthenticated       = errors.New("the user is not authenticated")    //401
	ErrOrderAlreadyExists         = errors.New("order already exists")             //409
	ErrInvalidOrderNumberFormat   = errors.New("invalid order number format")      //422
	ErrInternalServerError        = errors.New("internal server error")            //500
)
