package main

import (
	"fmt"

	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	NoBordered  bool   `long:"no-bordered" description:"不要下载有边框版本图片（边框图片的名字是日语）"`
	NoBase      bool   `long:"no-base" description:"不要下载未进化版本图片"`
	NoEvo       bool   `long:"no-evo" description:"不要下载进化版本图片"`
	NoRaw       bool   `long:"no-law" description:"不要下载原始大图"`
	NoSound     bool   `long:"no-sound" desciption:"不要下载配音"`
	NoJp        bool   `long:"no-jp" description:"不要下载日语配音"`
	NoEn        bool   `long:"no-en" description:"不要下载英语配音"`
	NoKo        bool   `long:"no-ko" description:"不要下载韩语配音"`
	Path        string `short:"p" long:"path" default:"work" description:"指定存放目录（默认是当前目录下的work目录，目录不存在会自动创建，目录必须为空目录）"`
	TheardCount int    `short:"c" long:"count" default:"20" description:"下载线程数。默认：20"`
}

func main() {
	var o Options
	_, err := flags.Parse(&o)
	if err != nil {
		fmt.Println("输入的参数有错误。如果你在查看帮助信息，请忽略。")
		return
	}
	flag := NewStoreFlag()
	switch true {
	case o.NoBordered:
		flag.Disable(BaseArtWithBorderStoreFlag)
		flag.Disable(EvoArtWithBorderStoreFlag)
	case o.NoBase:
		flag.Disable(BaseArtStoreFlag)
		flag.Disable(BaseArtWithBorderStoreFlag)
	case o.NoEvo:
		flag.Disable(EvoArtStoreFlag)
		flag.Disable(EvoArtWithBorderStoreFlag)
	case o.NoRaw:
		flag.Disable(BaseArtStoreFlag)
		flag.Disable(EvoArtStoreFlag)
	case o.NoJp:
		flag.Disable(JpSoundTrackStoreFlag)
	case o.NoEn:
		flag.Disable(EnSoundTrackStoreFlag)
	case o.NoKo:
		flag.Disable(KoSoundTrackStoreFlag)
	case o.NoSound:
		flag.Disable(JpSoundTrackStoreFlag)
		flag.Disable(EnSoundTrackStoreFlag)
		flag.Disable(KoSoundTrackStoreFlag)
	}
	if o.Path != "" {
		flag.Path = o.Path
	}
	if o.TheardCount != 0 {
		flag.ThreadCount = o.TheardCount
	}
	server := NewServer(*flag)
	server.Do()
	if server.Err != nil {
		fmt.Printf("Error %v\n", server.Err)
	}
}
