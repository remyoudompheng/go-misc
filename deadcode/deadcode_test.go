package main

import (
	"go/types"
	"testing"
)

func TestP1(t *testing.T) {
	ctx := new(Context)
	ctx.Load("./testdata/p1")
	objs := ctx.Process()
	compare(t, objs, []string{
		"unused",
		"g",
		"H",
		// "h", // recursive functions are not supported
	})
}

func TestP2(t *testing.T) {
	ctx := new(Context)
	ctx.Load("./testdata/p2")
	objs := ctx.Process()
	compare(t, objs, []string{
		"main",
		"unused",
		"g",
		// "h", // recursive functions are not supported
	})
}

func TestWithTestFiles(t *testing.T) {
	ctx := &Context{withTests: true}
	ctx.Load("./testdata/p3")
	objs := ctx.Process()
	// Only "y" is unused, x is used in tests.
	compare(t, objs, []string{"y"})
}

func compare(t *testing.T, objs []types.Object, names []string) {
	left := make(map[string]bool)
	right := make(map[string]bool)
	for _, o := range objs {
		left[o.Name()] = true
	}
	for _, n := range names {
		right[n] = true
	}

	for _, o := range objs {
		if !right[o.Name()] {
			t.Errorf("%s should not have been reported as unused", o.Name())
		}
	}
	for _, n := range names {
		if !left[n] {
			t.Errorf("unused %s should not have been reported", n)
		}
	}
}
