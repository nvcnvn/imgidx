// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
	"bytes"
	"github.com/openvn/nstuff"
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

func fullResult(s *nstuff.Host, img Image) string {
	var buff bytes.Buffer
	buff.WriteString(img.SavedLocation)
	buff.WriteString("|-|-|")
	buff.WriteString(img.Description)
	buff.WriteString("|-|-|")
	buff.WriteString(img.Location)

	pagCol := s.Conn.Storage("Page")
	key, err := pagCol.DecodeKey(img.PageID)
	if err != nil {
		return buff.String()
	}
	var page Page
	err = pagCol.Get(key, &page)
	if err != nil {
		return buff.String()
	}
	buff.WriteString("|-|-|")
	buff.WriteString(page.Location)
	return buff.String()
}
