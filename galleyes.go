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
	crawlRouter, _ := nstuff.NewRouter("/crawl$", Crawl)
	indexRouter, _ := nstuff.NewRouter("$", Index)

	var s nstuff.Host
	s.AddRouters(crawlRouter, hashRouter, indexRouter)
	s.Run()
}

func Index(s *nstuff.Host) {
	html := `<!doctype html>
<html lang="vn">
<head>
	<meta charset="utf-8">
	<title>Image Index</title>
	<link href="/statics/css/cupertino/jquery-ui-1.9.1.custom.css" rel="stylesheet">
	<script src="/statics/js/jquery-1.8.2.js"></script>
	<script src="/statics/js/jquery-ui-1.9.1.custom.js"></script>
	<script>
	$(function() {
      $("#upload_link").click(function(){
	    $('#upload_frame').contents().find("#uploader").click();
	  });
	});
	function ShowResult(query) {
		var arr = jQuery.parseJSON(query);
		appendImage(arr, arr.length, 0);
	}
	function appendImage(arr, len, step) {
		if (step >= len) {
			return;
		}
		$("<img/>", {
			"class": "img_result",
			src: arr[step].SavedLocation,
			title: arr[step].PageLocation
		}).tooltip().hide().appendTo("#results").show(
			'slide',
			'easeInOutCubic',
			1000,
			function() {
				appendImage(arr, len, step+1)
			});		
	}
	</script>
	<style>
	#layout {
    width: 70%;
    max-width: 700px;
    margin-left: auto;
    margin-right: auto;
  	}
	#file_browser {
    height: 100px;
    border: 5px solid #d7ebf9;
    border-radius: 25px;
    text-align: center;
	}
	.img_result {
	width: 100px;
	margin: 10px;
	border-radius: 10px;
	}
	</style>
</head>
<body>
  <div id="layout">
    <div class="ui-widget-content" id="file_browser">
      <h1 id="upload_link">Select an Image!</h1>
      <iframe id="upload_frame" class="ui-helper-hidden" src="hash"></iframe>
    </div>
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
  </div>
</body>
</html>`
	s.Print(html)
}

func Hash(s *nstuff.Host) {
	s.Print(`<!doctype html>
<html lang="vn">
<head>
	<meta charset="utf-8">
	<title>Image Index Uploader</title>
	<script>
	function submit() {
		document.getElementById("form").submit();
	}
	`)
	file, header, err := s.File("file")
	if err == nil {
		if err == nil {
			filetype := header.Header.Get("Content-Type")
			if filetype == "image/jpeg" || filetype == "image/png" || filetype == "image/gif" {
				m, _, err := tipimage.Decode(file)
				if err == nil {
					h := md5.New()
					io.Copy(h, file)
					checksum := base64.URLEncoding.EncodeToString(h.Sum(nil))
					phash, p := PHash(m)
					FindImage(s, checksum, phash, p)
				} else {
					s.Print("error=invalid_image&err_detail=", err.Error())
				}
			}
		} else {
			s.Print("error=invalid_input")
		}
	}
	s.Print(`
	</script>
</head>
<body>
  <form id="form" method="post" enctype="multipart/form-data" action="hash">
    <input id="uploader" type="file" name="file" onchange="submit();" />
  </form>
</body>
</html>
`)
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
