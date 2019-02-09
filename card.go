package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

const (
	//StartURL the first url to visit
	StartURL = "http://sv.bagoum.com/cardSort"
	//Domain the remote server Domain
	Domain = "http://sv.bagoum.com"
	//Store Flags
	BaseArtWithBorderStoreFlag = iota
	BaseArtStoreFlag
	EvoArtWithBorderStoreFlag
	EvoArtStoreFlag
	JpSoundTrackStoreFlag
	EnSoundTrackStoreFlag
	KoSoundTrackStoreFlag
)

var (
	//Ua the default user agent
	Ua                  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.81 Safari/537.36"
	DeafaultStoreConfig = *NewStoreFlag()
	getBorderedArtRe    = regexp.MustCompile(`.*C_(?P<name>\d+)\.png`)
	getCardNameRe       = regexp.MustCompile(`.*/0/(?P<name>.*)`)
	getName             = regexp.MustCompile(`(?P<name>\S+).*`) //It's tag before a line of audio players
	getCardName         = regexp.MustCompile(`.*/(?P<name>.*)`)
	uaSetter            = colly.UserAgent(Ua)
)

//Server connect the server and do the collection
type Server struct {
	frontPageCollector *colly.Collector
	Err                error
	threads            chan struct{}
}

//NewServer create a new Server instance
func NewServer(config StoreConfig) *Server {
	ret := new(Server)
	ret.frontPageCollector = colly.NewCollector()
	ret.threads = make(chan struct{}, config.ThreadCount)
	uaSetter(ret.frontPageCollector)

	err := os.Chdir(config.Path)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(config.Path, 0755)
		if err != nil {
			panic("chnage work directory failed")
		}
		_ = os.Chdir(config.Path)
	} else if err != nil {
		panic("change work directory failed")
	}

	ret.frontPageCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if !strings.HasPrefix(e.Attr("href"), "/card") {
			return
		}
		fmt.Println(Domain + e.Attr("href"))
		card := NewCard(Domain + e.Attr("href"))
		ret.threads <- struct{}{}
		go func() {
			ret.SetErr(card.Store(config))
			<-ret.threads
		}()
	})

	ret.frontPageCollector.OnError(func(r *colly.Response, err error) {
		ret.SetErr(fmt.Errorf("Connet to server failed.URL %s error string %v", r.Request.URL, err))
	})
	return ret
}

//SetErr record the server error
func (server *Server) SetErr(err error) {
	if server.Err == nil {
		server.Err = err
	}
}

//Do do the collection
func (server *Server) Do() {
	server.frontPageCollector.Visit(StartURL)
}

//Card store card status
type Card struct {
	cardPageCollector *colly.Collector
	url               string
	mux               sync.Mutex
	//under this part is the url to store
	Name                 string
	BaseArtWithBorderURL string
	EvoArtWithBorderURL  string
	BaseArtURL           string
	EvoArtURL            string
	JpSoundTrackURL      map[string]string
	EnSoundTrackURL      map[string]string
	KoSoundTrackURL      map[string]string
	/* Not Support Yet
	cardType             string
	rarity               string
	set                  string
	cost                 string
	baseStats            string
	baseEffect           string
	evoStatus            string
	evoEffect            string
	baseDescription string
	evoDescription string
	*/
}

//NewCard create a new card instance
func NewCard(url string) *Card {
	ret := new(Card)
	ret.cardPageCollector = colly.NewCollector()
	ret.url = url
	ret.Name = string(getCardName.FindSubmatch([]byte(url))[1])
	emptyMap := map[string]string{}
	ret.JpSoundTrackURL, ret.EnSoundTrackURL, ret.KoSoundTrackURL = emptyMap, emptyMap, emptyMap
	uaSetter(ret.cardPageCollector)

	//collection logic
	ret.cardPageCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.HasPrefix(href, "https://shadowverse-portal.com") {
			ret.BaseArtWithBorderURL = href
		}
		if strings.HasPrefix(href, "/getRawImage") {
			ret.BaseArtURL = Domain + href
		}
	})
	ret.cardPageCollector.OnHTML("tr", func(e *colly.HTMLElement) {
		has := false
		e.ForEachWithBreak("audio", func(_ int, _ *colly.HTMLElement) bool {
			has = true
			return false
		})
		if !has {
			return
		}
		name := string(getName.FindSubmatch([]byte(e.ChildText("td")))[1])
		e.ForEach("source[src]", func(_ int, e *colly.HTMLElement) {
			src := Domain + e.Attr("src")
			if strings.Contains(src, "/j/") {
				ret.JpSoundTrackURL[name] = src
			}
			if strings.Contains(src, "/e/") {
				ret.EnSoundTrackURL[name] = src
			}
			if strings.Contains(src, "/k/") {
				ret.KoSoundTrackURL[name] = src
			}
		})
	})
	ret.cardPageCollector.OnError(func(r *colly.Response, err error) {
		panic(fmt.Sprintf("Card Error: URL %s Error %v",
			r.Request.URL, err))
	})
	ret.cardPageCollector.Visit(url)

	res1 := getBorderedArtRe.FindSubmatch([]byte(ret.BaseArtWithBorderURL))
	res2 := getCardNameRe.FindSubmatch([]byte(ret.BaseArtURL))

	if res1 == nil || res2 == nil {
		panic("New Card Error: regular expressions do not match.")
	}

	//FIXME here need multi language support
	ret.EvoArtWithBorderURL = "https://shadowverse-portal.com/image/card/ja/E_" + string(res1[1]) + ".png"
	ret.EvoArtURL = "http://sv.bagoum.com/getRawImage/1/0/" + string(res2[1])

	return ret
}

//Store store the card infomation to disk
func (card *Card) Store(config StoreConfig) error {
	card.mux.Lock()
	defer card.mux.Unlock()

	err := os.Mkdir(card.Name, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	setErr := func(err_ error) {
		if err == nil {
			err = err_
		}
	}

	if config.Content[BaseArtWithBorderStoreFlag] {
		setErr(card.download(card.BaseArtWithBorderURL, card.Name+"/"+"base-with-border.png"))
	}
	if config.Content[BaseArtStoreFlag] {
		setErr(card.download(card.BaseArtURL, card.Name+"/"+"base-art.png"))
	}
	if config.Content[EvoArtStoreFlag] {
		setErr(card.download(card.EvoArtURL, card.Name+"/"+"evo-art.png"))
	}
	if config.Content[EvoArtWithBorderStoreFlag] {
		setErr(card.download(card.EvoArtWithBorderURL, card.Name+"/"+"evo-art-with-border.png"))
	}
	if config.Content[JpSoundTrackStoreFlag] {
		setErr(card.downloadSoundTrack(card.JpSoundTrackURL, "jp"))
	}
	if config.Content[EnSoundTrackStoreFlag] {
		setErr(card.downloadSoundTrack(card.EnSoundTrackURL, "en"))
	}
	if config.Content[KoSoundTrackStoreFlag] {
		setErr(card.downloadSoundTrack(card.KoSoundTrackURL, "ko"))
	}
	return err
}

func (card *Card) downloadSoundTrack(urls map[string]string, prefix string) error {
	var err error
	setErr := func(err_ error) {
		if err == nil {
			err = err_
		}
	}
	for item, url := range urls {
		setErr(card.download(url, card.Name+"/"+item+"-"+prefix+".mp3"))
	}
	return err
}

func (card *Card) download(url, fileName string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req.Header.Add("User-Agent", Ua)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileName, data, 0755)
	return err
}

type StoreConfig struct {
	Content     map[int]bool
	Path        string
	ThreadCount int
}

func NewStoreFlag() *StoreConfig {
	ret := new(StoreConfig)
	ret.Content = map[int]bool{}
	ret.Content[BaseArtStoreFlag] = true
	ret.Content[BaseArtWithBorderStoreFlag] = true
	ret.Content[EvoArtStoreFlag] = true
	ret.Content[EvoArtWithBorderStoreFlag] = true
	ret.Content[JpSoundTrackStoreFlag] = true
	ret.Content[EnSoundTrackStoreFlag] = true
	ret.Content[KoSoundTrackStoreFlag] = true
	ret.Path = "work"
	ret.ThreadCount = 20
	return ret
}

func (config *StoreConfig) Disable(flag int) {
	config.Content[flag] = false
}

func (config *StoreConfig) Enable(flag int) {
	config.Content[flag] = true
}

func (config *StoreConfig) SetPath(path string) {
	config.Path = path
}
