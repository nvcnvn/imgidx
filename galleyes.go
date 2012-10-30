// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"crypto/md5"
	"encoding/base64"
	"github.com/openvn/nstuff"
	"github.com/openvn/nstuff/model"
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
	resultsRouter, _ := nstuff.NewRouter("/results$", Results)
	crawlRouter, _ := nstuff.NewRouter("/crawl$", Crawl)
	indexRouter, _ := nstuff.NewRouter("$", Index)

	var s nstuff.Host
	s.AddRouters(resultsRouter, crawlRouter, searchRouter, hashRouter, indexRouter)
	s.Run()
}

func Index(s *nstuff.Host) {
	html := `<html>
<head>
<title>Search</title>
</head>
<body>
<script>
function ShowResult(query) {
	var xmlhttp;
	if (window.XMLHttpRequest)
	  {// code for IE7+, Firefox, Chrome, Opera, Safari
	  xmlhttp=new XMLHttpRequest();
	  }
	else
	  {// code for IE6, IE5
	  xmlhttp=new ActiveXObject("Microsoft.XMLHTTP");
	  }
	xmlhttp.onreadystatechange=function()
	  {
	  if (xmlhttp.readyState==4 && xmlhttp.status==200)
	    {
	    document.getElementById("results").innerHTML=xmlhttp.responseText;
	    }
	  }
	xmlhttp.open("GET","results?"+query,true);
	xmlhttp.send();
}
</script>
<iframe src="search"><a href="search">Click here to Search!!!</a></iframe>
<div id="results"></div>
<form method="post" action="crawl">
<input type="text" name="url" />
<select name="deep">
	<option value="-1">All</option>
	<option value="0" selected="selected">Only this</option>
	<option value="1">1 more levels</option>
	<option value="3">3 more levels</option>
	<option value="9">9 more levels</option>
	<option value="27">27 more levels</option>
</select>
<input type="submit" value="upload" />
</form>
</body>
</html>`
	s.Print(html)
}

func Hash(s *nstuff.Host) {
	s.Print(`<html><head></head>
<body>
<script>
parent.ShowResult("`)
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
				s.Print("&phash=", phash, "&p0=", part[0], "&p1=", part[1], "&p2=", part[2])
				s.Print("&p3=", part[3], "&p4=", part[4], "&p5=", part[5], "&p6=", part[6], "&p7=", part[7])
			} else {
				s.Print("error=invalid_image&err_detail=", err.Error())
			}
		}
	} else {
		s.Print("error=invalid_input")
	}
	s.Print(`")
</script>
</body>
</html>
`)
}

func Search(s *nstuff.Host) {
	html := `<html>
<head><title>Image Search</title></head>
<body>
<form method="post" enctype="multipart/form-data" action="hash">
<input type="file" name="file" /><input type="submit" value="Search" />
</form>
</body>
</html>`
	s.Print(html)
}

func Results(s *nstuff.Host) {
	var err error
	checksum := s.Get("checksum")
	var phash int64
	phash, err = strconv.ParseInt(s.Get("phash"), 10, 64)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	var p [8]int64
	p[0], err = strconv.ParseInt(s.Get("p0"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[1], err = strconv.ParseInt(s.Get("p1"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[2], err = strconv.ParseInt(s.Get("p2"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[3], err = strconv.ParseInt(s.Get("p3"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[4], err = strconv.ParseInt(s.Get("p4"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[5], err = strconv.ParseInt(s.Get("p5"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[6], err = strconv.ParseInt(s.Get("p6"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	p[7], err = strconv.ParseInt(s.Get("p7"), 10, 8)
	if err != nil {
		s.Log("Error: %s", err.Error())
		return
	}
	imgColl := s.Conn.Storage("Image")

	//find exacts image
	var exacts []Image
	_, err = imgColl.NewQuery().Filter("CheckSum", model.EQ, checksum).GetAll(&exacts)
	if err == nil {
		for _, value := range exacts {
			s.Print(fullResult(s, value), "\n")
		}
	}
	//find very similars image
	var similars []Image
	_, err = imgColl.NewQuery().Filter("CheckSum", model.GT, checksum).
		Filter("PHash", model.EQ, phash).GetAll(&similars)
	if err == nil {
		for _, value := range similars {
			s.Print(fullResult(s, value), "\n")
		}
	}
	_, err = imgColl.NewQuery().Filter("CheckSum", model.LT, checksum).
		Filter("PHash", model.EQ, phash).GetAll(&similars)
	if err == nil {
		for _, value := range similars {
			s.Print(fullResult(s, value), "\n")
		}
	}
	//find some maybe image
	part := [8]string{"Part0", "Part1", "Part2", "Part3", "Part4", "Part5", "Part6", "Part7"}
	var maybes []Image
	for i := 0; i < 8; i++ {
		a := imgColl.NewQuery()
		for j := 0; j < 8; j++ {
			if j == i {
				a.Filter(part[j], model.GT, p[j])
			} else {
				a.Filter(part[j], model.EQ, p[j])
			}
		}
		_, err = a.GetAll(&maybes)
		if err == nil {
			for _, value := range maybes {
				s.Print(fullResult(s, value), "\n")
			}
		}
		b := imgColl.NewQuery()
		for j := 0; j < 8; j++ {
			if j == i {
				b.Filter(part[j], model.LT, p[j])
			} else {
				b.Filter(part[j], model.EQ, p[j])
			}
		}
		_, err = b.GetAll(&maybes)
		if err == nil {
			for _, value := range maybes {
				s.Print(fullResult(s, value), "\n")
			}
		}
	}

}

func Crawl(s *nstuff.Host) {
	page := s.Post("url")
	if !ValidURL(page) {
		s.Print("error: invalid image url.")
		return
	}
	deep, err := strconv.Atoi(s.Post("deep"))
	if err != nil {
		deep = 0
	}
	NewImageIndex(s, page, deep).Index()
}
