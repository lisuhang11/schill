package types

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (r *RegisterReq) Validate() error {
	return validate.Struct(r)
}

func (r *LoginReq) Validate() error {
	return validate.Struct(r)
}

func (r *RefreshReq) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateAvatarReq) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateUserProfileInfoReq) Validate() error {
	return validate.Struct(r)
}
