// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

func ValidURL(url string) bool {
	if len(url) > 3 {
		return true
	}
	return false
}
