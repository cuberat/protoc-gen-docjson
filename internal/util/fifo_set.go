package util

// BSD 2-Clause License
//
// Copyright (c) 2023 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice,
//   this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

type FifoSet[T comparable] struct {
	Items   []T
	ItemMap map[T]bool
}

func NewStringSet() *FifoSet[string] {
	size := 2
	return &FifoSet[string]{
		Items:   make([]string, 0, size),
		ItemMap: make(map[string]bool, size),
	}
}

func NewIntSet() *FifoSet[int] {
	size := 2
	return &FifoSet[int]{
		Items:   make([]int, 0, size),
		ItemMap: make(map[int]bool, size),
	}
}

func (s *FifoSet[T]) Add(item T) {
	if _, ok := s.ItemMap[item]; ok {
		return
	}

	s.Items = append(s.Items, item)
	s.ItemMap[item] = true
}

func (s *FifoSet[T]) Update(items []T) {
	for _, item := range items {
		s.Add(item)
	}
}

func (s *FifoSet[T]) GetItems() []T {
	return s.Items
}

func (s *FifoSet[T]) Len() int {
	return len(s.Items)
}
