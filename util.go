// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"encoding/json"
	"github.com/openvn/nstuff"
	"github.com/openvn/nstuff/model"
	"net/url"
	"path"
)

func ValidURL(url string) bool {
	if len(url) > 3 {
		return true
	}
	return false
}

func FullURL(abs, href string) (string, error) {
	u, err := url.Parse(abs)
	if err != nil {
		return "", err
	}
	dir, _ := path.Split(u.Path)

	result := ""
	if u.Scheme != "" {
		result += u.Scheme + ":"
	}
	if u.Opaque != "" {
		result += u.Opaque
	} else {
		if u.Host != "" || u.User != nil {
			result += "//"
			if u := u.User; u != nil {
				result += u.String() + "@"
			}
			result += u.Host
		}
		result += path.Join(dir, href)
	}

	return result, nil
}

type JsonImage struct {
	SavedLocation string
	Description   string
	Location      string
	PageLocation  string
}

func fullResult(s *nstuff.Host, img []Image) string {
	n := len(img)
	result := make([]JsonImage, n, n)
	pagCol := s.Conn.Storage("Page")
	var page Page
	for i := 0; i < n; i++ {
		result[i].Description = img[i].Description
		result[i].Location = img[i].Location
		result[i].SavedLocation = img[i].SavedLocation
		key, err := pagCol.DecodeKey(img[i].PageID)
		if err != nil {
			continue
		}
		err = pagCol.Get(key, &page)
		if err != nil {
			continue
		}
		result[i].PageLocation = page.Location
	}
	b, err := json.Marshal(result)
	if err == nil {
		return string(b)
	}
	return ""
}

func FindImage(s *nstuff.Host, checksum string, phash int64, p [8]int8) {
	var err error
	imgColl := s.Conn.Storage("Image")
	//find exacts image
	var exacts []Image
	_, err = imgColl.NewQuery().Filter("CheckSum", model.EQ, checksum).GetAll(&exacts)
	if err == nil {
		s.Print("parent.ShowResult('", fullResult(s, exacts), "');\n")
	}
	//find very similars image
	var similars []Image
	_, err = imgColl.NewQuery().Filter("CheckSum", model.GT, checksum).
		Filter("PHash", model.EQ, phash).GetAll(&similars)
	if err == nil {
		s.Print("parent.ShowResult('", fullResult(s, similars), "');\n")
	}
	_, err = imgColl.NewQuery().Filter("CheckSum", model.LT, checksum).
		Filter("PHash", model.EQ, phash).GetAll(&similars)
	if err == nil {
		s.Print("parent.ShowResult('", fullResult(s, similars), "');\n")
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
			s.Print("parent.ShowResult('", fullResult(s, maybes), "');\n")
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
			s.Print("parent.ShowResult('", fullResult(s, maybes), "');\n")
		}
	}
}
