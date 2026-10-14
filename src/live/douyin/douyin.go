package douyin

import (
	"net/url"

	"github.com/bililive-go/bililive-go/src/live"
	"github.com/bililive-go/bililive-go/src/live/internal"
)

const (
	domain       = "live.douyin.com"
	domainForApp = "v.douyin.com"
	cnName       = "抖音"
)

func init() {
	live.Register(domain, new(builder))
	live.Register(domainForApp, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL) (live.Live, error) {
	ret := &Live{
		BaseLive: internal.NewBaseLive(url),
	}
	ret.btoolsLive = NewBtoolsLive(ret)
	return ret, nil
}

type Live struct {
	internal.BaseLive
	btoolsLive btoolsLive
}

// GetRoomID 返回 bililive-tools 已解析的抖音 web_rid。
// 短链接路径只是分享码，不能用于连接弹幕 WebSocket。
func (l *Live) GetRoomID() string {
	return l.btoolsLive.roomId
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	return l.btoolsLive.GetInfo()
}

func (l *Live) GetStreamInfos() (us []*live.StreamUrlInfo, err error) {
	return l.btoolsLive.GetStreamInfos()
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
