// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

import (
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
