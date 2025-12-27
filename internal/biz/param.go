package biz

//定义biz层结构体

// ReplyParam 商家回复评价的参数，解耦
type ReplyParam struct {
	ReviewID  int64
	StoreID   int64
	Content   string
	PicInfo   string
	VideoInfo string
	ReplyID  int64
}