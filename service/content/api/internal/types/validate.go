package types

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (r *CreatePostReq) Validate() error {
	return validate.Struct(r)
}

func (r *UpdatePostReq) Validate() error {
	return validate.Struct(r)
}

func (r *DeletePostReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetPostDetailReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetPostListReq) Validate() error {
	return validate.Struct(r)
}

func (r *BatchGetPostReq) Validate() error {
	return validate.Struct(r)
}

func (r *IncViewCountReq) Validate() error {
	return validate.Struct(r)
}
