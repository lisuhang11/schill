package types

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (r *CreateCommentReq) Validate() error {
	return validate.Struct(r)
}

func (r *DeleteCommentReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetCommentListReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetReplyListReq) Validate() error {
	return validate.Struct(r)
}

func (r *VoteCommentReq) Validate() error {
	return validate.Struct(r)
}
