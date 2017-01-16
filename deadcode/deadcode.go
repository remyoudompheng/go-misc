package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/loader"
)

var exitCode int

var (
	withTestFiles bool
)

func main() {
	flag.BoolVar(&withTestFiles, "test", false, "include test files")
	flag.Parse()
	if flag.NArg() == 0 {
		doDirs([]string{"."}, withTestFiles)
	} else {
		for _, name := range flag.Args() {
			// Is it a directory?
			if fi, err := os.Stat(name); err != nil || !fi.IsDir() {
				fatalf("not a directory: %s", name)
			}
		}
		doDirs(flag.Args(), withTestFiles)
	}
	os.Exit(exitCode)
}

// error formats the error to standard error, adding program
// identification and a newline
func errorf(pos token.Position, format string, args ...interface{}) {
	pwd, _ := os.Getwd()
	f, err := filepath.Rel(pwd, pos.Filename)
	if err == nil {
		pos.Filename = f
	}
	fmt.Fprintf(os.Stderr, pos.String()+": "+format+"\n", args...)
	exitCode = 2
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func doDirs(names []string, withTests bool) []types.Object {
	var conf loader.Config
	for _, name := range names {
		if withTests {
			conf.ImportWithTests(name)
		} else {
			conf.Import(name)
		}
	}
	prog, err := conf.Load()
	if err != nil {
		fatalf("cannot load packages: %s", err)
	}
	var allUnused []types.Object
	for _, pkg := range prog.Imported {
		unused := doPackage(prog, pkg)
		for _, obj := range unused {
			errorf(prog.Fset.Position(obj.Pos()), "%s is unused", obj.Name())
		}
		allUnused = append(allUnused, unused...)
	}
	return allUnused
}

type Package struct {
	p    *ast.Package
	fs   *token.FileSet
	decl map[string]ast.Node
	used map[string]bool
}

func doPackage(prog *loader.Program, pkg *loader.PackageInfo) []types.Object {
	used := make(map[types.Object]bool)
	for _, file := range pkg.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pkg.Info.Uses[id]
			if obj != nil {
				used[obj] = true
			}
			return false
		})
	}

	global := pkg.Pkg.Scope()
	var unused []types.Object
	for _, name := range global.Names() {
		if pkg.Pkg.Name() == "main" && name == "main" {
			continue
		}
		obj := global.Lookup(name)
		if !used[obj] && (pkg.Pkg.Name() == "main" || !ast.IsExported(name)) {
			unused = append(unused, obj)
		}
	}
	return unused
}
