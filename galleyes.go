// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"crypto/md5"
	"encoding/base64"
	"github.com/openvn/nstuff"
	"io"
	"strconv"
	"tipimage"
	_ "tipimage/gif"
	_ "tipimage/jpeg"
	_ "tipimage/png"
)

func init() {
	hashRouter, _ := nstuff.NewRouter("/hash$", Hash)
	searchRouter, _ := nstuff.NewRouter("/search$", Search)
	crawlRouter, _ := nstuff.NewRouter("/crawl$", Crawl)
	indexRouter, _ := nstuff.NewRouter("$", Index)

	var s nstuff.Host
	s.AddRouters(crawlRouter, searchRouter, hashRouter, indexRouter)
	s.Run()
}

func Index(s *nstuff.Host) {
	html := `<html>
<head>
<title>Search</title>
</head>
<body>
<form method="post" enctype="multipart/form-data" action="hash">
<input type="file" name="file" />
<input type="submit" value="upload" />
</form>
<form method="post" action="crawl">
<input type="text" name="url" />
<input type="submit" value="upload" />
</form>
</body>
</html>`
	s.Print(html)
}

func Hash(s *nstuff.Host) {
	file, header, err := s.File("file")
	if err == nil {
		filetype := header.Header.Get("Content-Type")
		if filetype == "image/jpeg" || filetype == "image/png" || filetype == "image/gif" {
			m, _, err := tipimage.Decode(file)
			if err == nil {
				h := md5.New()
				io.Copy(h, file)

				s.Print("checksum=", base64.URLEncoding.EncodeToString(h.Sum(nil)))
				phash, part := PHash(m)
				s.Print("&phash=", phash, "&p1=", part[0], "&p2=", part[1], "&p3=", part[2])
				s.Print("&p4=", part[3], "&p5=", part[4], "&p6=", part[5], "&p7=", part[6], "&p8=", part[1])
			} else {
				s.Print("error=invalid_image&err_detail=", err.Error())
			}
		}
	} else {
		s.Print("error=invalid_input")
	}
}

func Search(s *nstuff.Host) {
	checksum, err := base64.URLEncoding.DecodeString(s.Get("checksum"))
	phash, err := strconv.ParseInt(s.Get("phash"), 10, 64)
}

func Crawl(s *nstuff.Host) {
	page := s.Post("url")
	if !ValidURL(page) {
		s.Print("error: invalid image url.")
		return
	}
	NewImageIndex(s, page, 0).Index()
}
