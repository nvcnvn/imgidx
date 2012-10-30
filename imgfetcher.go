// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"appengine/blobstore"
	imgs "appengine/image"
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/base64"
	"github.com/openvn/nstuff"
	"github.com/openvn/nstuff/model"
	"io/ioutil"
	"runtime"
	"sync"
	"tipimage"
	_ "tipimage/gif"
	_ "tipimage/jpeg"
	_ "tipimage/png"
)

type imageInfo struct {
	src string
	alt string
}

type pageInfo struct {
	url  string
	deep int
}

type ImageIndex struct {
	pageList *list.List
	s        *nstuff.Host
	rootURL  string
	deep     int
	wg       sync.WaitGroup
}

func NewImageIndex(s *nstuff.Host, url string, deep int) *ImageIndex {
	i := &ImageIndex{}
	i.s = s
	i.rootURL = url
	i.deep = deep
	return i
}

func (i *ImageIndex) Index() {
	runtime.GOMAXPROCS(20)
	i.pageList = list.New()
	i.pageList.PushBack(pageInfo{i.rootURL, 0})

	for i.pageList.Len() > 0 {
		p := i.pageList.Remove(i.pageList.Front()).(pageInfo)
		key, err := i.s.Conn.Storage("Page").NewQuery().KeysOnly().
			Filter("Location", model.EQ, p.url).GetFirst(nil)
		if err != model.ErrNotFound {
			continue
		}

		resp, err := i.s.Client.Get(p.url)
		if err != nil {
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		if i.deep < 0 {
			i.FetchPage(data, p.deep+1)
		} else {
			if p.deep < i.deep {
				i.FetchPage(data, p.deep+1)
			}
		}
		key, err = i.s.Conn.Storage("Page").Put(&Page{p.url})
		i.FetchSrc(data, key.Encode())
	}
	i.wg.Wait()
}

func (i *ImageIndex) FetchSrc(data []byte, pageID string) {
	// openTag: <img
	openTag := []byte{0x3c, 0x69, 0x6d, 0x67}
	openPos := 0
	closePos := 0
	// srcTag: src
	srcTag := []byte{0x73, 0x72, 0x63}
	srcPos := 0
	// altTag: alt
	altTag := []byte{0x61, 0x6c, 0x74}
	altPos := 0
	// quote: " (0x22)
	quoteOpenPos := 0
	quoteClosePos := 0
	found := bytes.Index(data[openPos:], openTag)
	var tmpSlice []byte
	var url string
	for found = bytes.Index(data[openPos:], openTag); found != -1; found = bytes.Index(data[openPos:], openTag) {
		openPos = openPos + found + 5
		closePos = bytes.IndexByte(data[openPos:], 0x3e)
		tmpSlice = data[openPos : openPos+closePos]

		var alt string
		altPos = bytes.Index(tmpSlice, altTag)
		if altPos != -1 {
			quoteOpenPos = bytes.IndexByte(tmpSlice[altPos+4:], 0x22)
			if quoteOpenPos != -1 {
				quoteClosePos = bytes.IndexByte(tmpSlice[altPos+4+quoteOpenPos+1:], 0x22)
				if quoteClosePos != -1 {
					alt = string(tmpSlice[altPos+4+quoteOpenPos+1 : altPos+4+quoteOpenPos+quoteClosePos+1])
				}
			}
		}

		srcPos = bytes.Index(tmpSlice, srcTag)
		if srcPos != -1 {
			quoteOpenPos = bytes.IndexByte(tmpSlice[srcPos+4:], 0x22)
			if quoteOpenPos != -1 {
				quoteClosePos = bytes.IndexByte(tmpSlice[srcPos+4+quoteOpenPos+1:], 0x22)
				if quoteClosePos != -1 {
					url, _ = FullURL(i.rootURL, string(tmpSlice[srcPos+4+quoteOpenPos+1:srcPos+4+quoteOpenPos+quoteClosePos+1]))
					go i.FetchImage(imageInfo{url, alt}, pageID)
				}
			}
		}
	}
}

func (i *ImageIndex) FetchPage(data []byte, deep int) {
	// openTag: <a
	openTag := []byte{0x3c, 0x61}
	openPos := 0
	closePos := 0
	// hrefTag: href
	hrefTag := []byte{0x68, 0x72, 0x65, 0x66}
	hrefPos := 0
	// quote: " (0x22)
	quoteOpenPos := 0
	quoteClosePos := 0
	found := bytes.Index(data[openPos:], openTag)
	var tmpSlice []byte
	var url string
	for found = bytes.Index(data[openPos:], openTag); found != -1; found = bytes.Index(data[openPos:], openTag) {
		openPos = openPos + found + 3
		closePos = bytes.IndexByte(data[openPos:], 0x3e)
		tmpSlice = data[openPos : openPos+closePos]

		hrefPos = bytes.Index(tmpSlice, hrefTag)
		if hrefPos != -1 {
			quoteOpenPos = bytes.IndexByte(tmpSlice[hrefPos+5:], 0x22)
			if quoteOpenPos != -1 {
				quoteClosePos = bytes.IndexByte(tmpSlice[hrefPos+5+quoteOpenPos+1:], 0x22)
				if quoteClosePos != -1 {
					url, _ = FullURL(i.rootURL, string(tmpSlice[hrefPos+5+quoteOpenPos+1:hrefPos+5+quoteOpenPos+quoteClosePos+1]))
					i.pageList.PushBack(pageInfo{url, deep})
				}
			}
		}
	}
}

func (i *ImageIndex) FetchImage(info imageInfo, pageID string) {
	i.wg.Add(1)
	_, err := i.s.Conn.Storage("Image").NewQuery().KeysOnly().
		Filter("Location", model.EQ, info.src).GetFirst(nil)
	if err != model.ErrNotFound {
		i.s.Log("Error: %s\n", "galleyes: indexed image")
		i.wg.Done()
		return
	}
	resp, err := i.s.Client.Get(info.src)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	mime := resp.Header.Get("Content-Type")
	if mime != "image/png" && mime != "image/jpeg" && mime != "image/gif" {
		i.s.Log("Error: %s - %s - %s\n", "galleyes: not supported image format", mime, info.src)
		i.wg.Done()
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}

	// caculate PHash
	var buff bytes.Buffer
	_, err = buff.Write(data)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	m, _, err := tipimage.Decode(&buff)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	hash, part := PHash(m)

	// caculate md5 checksum
	h := md5.New()
	_, err = h.Write(data)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	checksum := h.Sum(nil)

	// save the original image
	w, err := blobstore.Create(i.s.Context, mime)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	w.Write(data)
	_, err = w.Write(data)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	err = w.Close()
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	key, err := w.Key()
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	link, err := imgs.ServingURL(i.s.Context, key, nil)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}

	// assign value
	img := Image{}
	img.PageID = pageID
	img.SavedID = string(key)
	img.SavedLocation = link.String()
	img.Location = info.src
	img.Description = info.alt
	img.CheckSum = base64.URLEncoding.EncodeToString(checksum)
	img.PHash = hash
	img.Part0 = part[0]
	img.Part1 = part[1]
	img.Part2 = part[2]
	img.Part3 = part[3]
	img.Part4 = part[4]
	img.Part5 = part[5]
	img.Part6 = part[6]
	img.Part7 = part[7]
	_, err = i.s.Conn.Storage("Image").Put(&img)
	if err != nil {
		i.s.Log("Error: %s\n", err.Error())
		i.wg.Done()
		return
	}
	i.wg.Done()
}
