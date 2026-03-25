package types

import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (r *FollowReq) Validate() error {
	return validate.Struct(r)
}

func (r *UnfollowReq) Validate() error {
	return validate.Struct(r)
}

func (r *CheckFollowStatusReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetFollowingListReq) Validate() error {
	return validate.Struct(r)
}

func (r *GetFollowerListReq) Validate() error {
	return validate.Struct(r)
}

func (r *BatchCheckFollowStatusReq) Validate() error {
	return validate.Struct(r)
}
