// Copyright 2011 The Open-VN.org Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package galleyes

type Image struct {
	CheckSum    []byte
	PHash       int64
	Part1       int8
	Part2       int8
	Part3       int8
	Part4       int8
	Part5       int8
	Part6       int8
	Part7       int8
	Part8       int8
	ID          string
	Saved       string
	Location    string
	Source      string
	Description string
}
