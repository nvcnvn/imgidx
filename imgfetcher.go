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
	"errors"
	"github.com/openvn/nstuff"
	"github.com/openvn/nstuff/model"
	"io/ioutil"
	"tipimage"
	_ "tipimage/gif"
	_ "tipimage/jpeg"
	_ "tipimage/png"
)

type ImageInfo struct {
	Src string
	Alt string
}

func IndexImage(s *nstuff.Host, url string, deep int) {
	pageList := list.New()
	pageList.PushBack(url)
	for pageList.Len() > 0 {
		url = pageList.Remove(pageList.Front()).(string)
		key, err := s.Conn.Storage("Page").NewQuery().KeysOnly().
			Filter("Location", model.EQ, url).GetFirst(nil)
		if err != model.ErrNotFound {
			s.Log("--%s--\n", key)
			return
		}

		key, err = s.Conn.Storage("Page").Put(&Page{url})
		resp, err := s.Client.Get(url)
		if err != nil {
			return
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		imgChan := make(chan ImageInfo)
		go FetchSrc(s, data, imgChan)
		pageChan := make(chan string)
		go FetchPage(s, data, pageChan)
		for i := range imgChan {
			s.Print(i, "\n")
		}
		for i := range pageChan {
			s.Print(i, "\n")
		}
	}
}

func FetchSrc(s *nstuff.Host, data []byte, ImgInf chan ImageInfo) {
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
					ImgInf <- ImageInfo{string(tmpSlice[srcPos+4+quoteOpenPos+1 : srcPos+4+quoteOpenPos+quoteClosePos+1]), alt}
				}
			}
		}
	}
	close(ImgInf)
}

func FetchPage(s *nstuff.Host, data []byte, PageInf chan string) {
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
					PageInf <- string(tmpSlice[hrefPos+5+quoteOpenPos+1 : hrefPos+5+quoteOpenPos+quoteClosePos+1])
				}
			}
		}
	}
	close(PageInf)
}

func FetchImage(s *nstuff.Host, url string, img *Image) error {
	resp, err := s.Client.Get(url)
	if err != nil {
		return err
	}
	mime := resp.Header.Get("Content-Type")
	if mime != "image/png" && mime != "image/jpeg" && mime != "image/gif" {
		return errors.New("galleyes: not supported image format")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// caculate PHash
	var buff bytes.Buffer
	_, err = buff.Write(data)
	if err != nil {
		return err
	}
	m, _, err := tipimage.Decode(&buff)
	if err != nil {
		s.Print(ioutil.ReadAll(&buff))
		return err
	}
	hash, part := PHash(m)

	// caculate md5 checksum
	h := md5.New()
	_, err = h.Write(data)
	if err != nil {
		return err
	}
	checksum := h.Sum(nil)

	// save the original image
	w, err := blobstore.Create(s.Context, mime)
	if err != nil {
		return err
	}
	w.Write(data)
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	key, err := w.Key()
	if err != nil {
		return err
	}
	link, err := imgs.ServingURL(s.Context, key, nil)
	if err != nil {
		return err
	}

	// assign value
	img.SavedID = string(key)
	img.SavedLocation = link.String()
	img.Location = url
	img.CheckSum = checksum
	img.PHash = hash
	img.Part1 = part[0]
	img.Part2 = part[1]
	img.Part3 = part[2]
	img.Part4 = part[3]
	img.Part5 = part[4]
	img.Part6 = part[5]
	img.Part7 = part[6]
	img.Part8 = part[7]
	return nil
}
