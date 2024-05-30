package filter

import (
	"github.com/qiancijun/trash/searchEngine/demo"
	"github.com/qiancijun/trash/searchEngine/demo/video_search/common"
)

type ViewFilter struct {
}

func (ViewFilter) Apply(ctx *common.VideoSearchContext) {
	request := ctx.Request
	if request == nil {
		return
	}
	if request.ViewFrom >= request.ViewTo {
		return
	}
	vidoes := make([]*demo.BiliVideo, 0, len(ctx.Videos))
	for _, video := range ctx.Videos {
		if video.View >= int32(request.ViewFrom) && video.View <= int32(request.ViewTo) {
			vidoes = append(vidoes, video)
		}
	}
	ctx.Videos = vidoes
}
