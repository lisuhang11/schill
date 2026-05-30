package redis

import (
	_ "embed"
)

var (
	//go:embed lua/post_like.lua
	PostLikeScript string

	//go:embed lua/post_unlike.lua
	PostUnlikeScript string

	//go:embed lua/post_collect.lua
	PostCollectScript string

	//go:embed lua/post_uncollect.lua
	PostUncollectScript string
)

type LuaScripts struct {
	PostLike      string
	PostUnlike    string
	PostCollect   string
	PostUncollect string
}

func NewLuaScripts() *LuaScripts {
	return &LuaScripts{
		PostLike:      PostLikeScript,
		PostUnlike:    PostUnlikeScript,
		PostCollect:   PostCollectScript,
		PostUncollect: PostUncollectScript,
	}
}
