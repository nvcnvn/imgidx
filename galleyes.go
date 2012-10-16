// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"github.com/openvn/nstuff"
	"tipimage"
	_ "tipimage/gif"
	_ "tipimage/jpeg"
	_ "tipimage/png"
)

func init() {
	indexRouter, _ := nstuff.NewRouter("$", Index)
	hashRouter, _ := nstuff.NewRouter("/hash$", Hash)
	searchRouter, _ := nstuff.NewRouter("/search$", Search)

	var s nstuff.Host
	s.AddRouters(searchRouter, hashRouter, indexRouter)
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
				s.Print(PHash(m))
			} else {
				s.Print("invalid")
				s.Print(err.Error())
			}
		}
	} else {
		s.Print("invalid")
	}
}

func Search(s *nstuff.Host) {

}
