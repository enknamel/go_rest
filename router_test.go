package main

import (
	"testing"
)

//the core of the rest router is makePath
func TestMakePath(t *testing.T) {
	r, p, err := makePath("/")
	if err != nil {
		t.Fatal(err)
	}

	if !r.MatchString("/") || r.MatchString("/test") {
		t.Fatal("Matched too much", r)
	}

	if len(p) != 0 {
		t.Fatal("Created a path param when shouldn't have ", p)
	}

	r, p, err = makePath("/lockservice/:resourceName/locks")

	if err != nil {
		t.Fatal(err)
	}

	if len(p) != 1 || p[0] != "resourceName" {
		t.Fatal("failed to create path params ", p)
	}

	if !r.MatchString("/lockservice/123456/locks") || r.MatchString("/lockservice/123456/locks/") {
		t.Fatal("Matched too much", r)
	}
}
