// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"crypto/md5"
	"github.com/openvn/nstuff"
	"io"
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
				s.Print(PHash(m))
				s.Print(h.Sum(nil))
			} else {
				s.Print("invalid")
				s.Print(err.Error())
			}
		}
	} else {
		s.Print("invalid")
	}
}

type Barer interface {
	Bar() int
}

type Bar string

type Test struct {
	X int
	Y Bar
}

func Search(s *nstuff.Host) {
	test := Test{5, Bar("ABC")}
	x, err := s.Conn.Storage("Image").Put(&test)
	if err != nil {
		s.Print(err.Error())
		return
	}
	s.Print(x.Encode())
}

func Crawl(s *nstuff.Host) {
	page := s.Post("url")
	if !ValidURL(page) {
		s.Print("error: invalid image url.")
		return
	}
	IndexImage(s, page, 0)
}
