package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "somefile"
	pathkey := CASPathTransformFunc(key)
	expectedPathname := "54dd0/11392/f2957/925eb/f27fc/6d86b/55a31/8dd16"
	expectedFilename := "54dd011392f2957925ebf27fc6d86b55a318dd16"
	if pathkey.Pathname != expectedPathname {
		t.Errorf("have %s want %s", pathkey.Pathname, expectedPathname)
	}
	if pathkey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathkey.Filename, expectedFilename)
	}
}

func TestStore(t *testing.T) {
	s := newStore()

	defer teardown(t, s)

	for i := 0; i < 50; i++ {
		fmt.Printf("running test %d 🚀\n", i)

		key := fmt.Sprintf("test_%d", i)
		data := []byte("a lot of data")

		// Write test
		if err := s.Write(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		fmt.Println("write test passed ✅")

		// Has test 1
		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key %s", key)
		}
		fmt.Println("has_exists test passed ✅")

		// Read test
		r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("read test passed ✅")

		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		// Delete test
		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
		fmt.Println("delete test passed ✅")

		// Has test 2
		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
		fmt.Println("has_no_file test passed ✅")
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
