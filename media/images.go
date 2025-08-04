package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"sync"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type Images struct {
	media map[string]*ImagesItem

	lock sync.Mutex
}

func NewImages() *Images {
	imgs := &Images{}

	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("webp", "webp", webp.Decode, webp.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("jpg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("tiff", "tiff", tiff.Decode, tiff.DecodeConfig)
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)

	imgs.media = make(map[string]*ImagesItem)

	return imgs
}

func (imgs *Images) Destroy() {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for _, it := range imgs.media {
		it.Destroy()
	}
}

func (imgs *Images) UpdateFileTimes() {
	imgs.lock.Lock()
	var ims []*ImagesItem
	for _, it := range imgs.media {
		ims = append(ims, it)
	}
	imgs.lock.Unlock()

	//slow
	for _, it := range ims {
		inf, err := os.Stat(it.path)
		if err == nil && inf != nil {
			it.check_file_time = inf.ModTime().UnixNano()
		}
	}
}

func (imgs *Images) Maintenance(min_time int64, fnImgChanged func(path string)) {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	for path, it := range imgs.media {

		diff := (it.check_file_time != it.open_file_time)
		if diff {
			fnImgChanged(path)
		}

		if (it.last_use_time > 0 && it.last_use_time < min_time) || diff {
			//fmt.Println("Maintenance() removing " + it.path)
			it.Destroy()
			delete(imgs.media, path)
		}
	}
}

func (imgs *Images) Find(path string) *ImagesItem {
	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	item, found := imgs.media[path]
	if found {
		return item
	}
	return nil
}

func (imgs *Images) Add(path string, blob []byte) (*ImagesItem, error) {
	//find
	item := imgs.Find(path)
	if item != nil {
		return item, nil
	}

	//create new media
	item, err := NewImagesItem(path, blob)
	if err != nil {
		item = &ImagesItem{path: path} //add it anyway, because file can be rewritten later. Error is return below.
	}

	imgs.lock.Lock()
	defer imgs.lock.Unlock()

	//add
	imgs.media[path] = item

	return item, err
}

type ImagesItem struct {
	path string

	width  int
	height int
	data   []byte //rgba

	last_use_time int64

	open_file_time  int64
	check_file_time int64
}

func NewImagesItem(path string, blob []byte) (*ImagesItem, error) {
	sp := &ImagesItem{path: path}

	sp.last_use_time = time.Now().UnixNano()

	//create new media
	var img image.Image
	if len(blob) > 0 {
		var err error
		img, _, err = image.Decode(bytes.NewReader(blob))
		if err != nil {
			return nil, err
		}
	} else if path != "" {
		imgf, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer imgf.Close()

		img, _, err = image.Decode(imgf)
		if err != nil {
			return nil, err
		}

		//file_time
		inf, err := imgf.Stat()
		if err == nil && inf != nil {
			sp.open_file_time = inf.ModTime().UnixNano()
			sp.check_file_time = sp.open_file_time
		}
	} else {
		return nil, fmt.Errorf("'%s' image format is not supported", string(path))
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

	sp.width = rgba.Rect.Size().X
	sp.height = rgba.Rect.Size().Y
	sp.data = rgba.Pix

	return sp, nil
}
func (sp *ImagesItem) Destroy() {
}
