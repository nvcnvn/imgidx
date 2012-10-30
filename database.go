// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

type Page struct {
	Location string
}

type Image struct {
	CheckSum      string
	PHash         int64
	Part0         int8
	Part1         int8
	Part2         int8
	Part3         int8
	Part4         int8
	Part5         int8
	Part6         int8
	Part7         int8
	SavedID       string
	SavedLocation string
	Location      string
	Description   string
	PageID        string
}
