package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

var (
	filename   = flag.String("f", "", "input file name")
	clientname = flag.String("c", "", "struct holding reference to a grpc client")
)

func main() {
	flag.Parse()

	b, err := os.ReadFile(*filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", string(b), 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	methods := []string{}
	for _, d := range f.Decls {
		switch t := d.(type) {
		case *ast.FuncDecl:
			if t.Recv == nil {
				// regular function, skip
				continue
			}
			name, typ := receiver(t.Recv.List[0])
			if typ != *clientname {
				continue
			}
			for _, stmt := range t.Body.List {
				m := stmtType(stmt, name)
				if m != "" {
					methods = append(methods, m)
				}
			}
		}
	}
	sort.Strings(methods)
	fmt.Println(strings.Join(methods, ", "))
}

func receiver(recExp *ast.Field) (name string, typ string) {
	// should not happen, since the client call needs a selector expr starting with the receiver
	if len(recExp.Names) == 0 {
		return "", ""
	}
	name = recExp.Names[0].Name
	switch t := recExp.Type.(type) {
	case *ast.StarExpr:
		typ = t.X.(*ast.Ident).Name
	case *ast.Ident:
		typ = t.Name
	}
	return
}

func stmtType(stmt ast.Stmt, target string) string {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for _, e := range s.Rhs {
			if r, found := walkExp(e, target); found {
				return r
			}
		}
	case *ast.ReturnStmt:
		for _, e := range s.Results {
			if r, found := walkExp(e, target); found {
				return r
			}
		}
	case *ast.ExprStmt:
		if r, found := walkExp(s.X, target); found {
			return r
		}
	}
	return ""
}

func walkExp(expr ast.Expr, target string) (string, bool) {
	switch t := expr.(type) {
	case *ast.CallExpr:
		return walkExp(t.Fun, target)
	case *ast.SelectorExpr:
		_, b := walkExp(t.X, target)
		if b {
			return t.Sel.Name, true
		}
	case *ast.Ident:
		if t.Name == target {
			return t.Name, true
		}
	}
	return "", false
}
