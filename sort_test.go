package main

import (
	"reflect"
	"testing"
)

func TestReviewOrder(t *testing.T) {
	names := []string{"tests/unrelated_test.rb", "src/deep/beta.ts", "tests/deep/beta.test.ts", "test/alpha_test.rb", "alpha.rb", "README.md", "test/other_test.rb", "deep/deeper/gamma.py", "deep/deeper/test_gamma.py"}
	files := make([]File, len(names))
	for i, name := range names {
		files[i].Name = name
	}
	sorted := sortFiles(files)
	got := make([]string, len(sorted))
	for i, f := range sorted {
		got[i] = f.Name
	}
	want := []string{"README.md", "alpha.rb", "test/alpha_test.rb", "deep/deeper/gamma.py", "deep/deeper/test_gamma.py", "src/deep/beta.ts", "tests/deep/beta.test.ts", "test/other_test.rb", "tests/unrelated_test.rb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestPairingDoesNotCrossModules(t *testing.T) {
	sorted := sortFiles([]File{{Name: "backend/src/foo.go"}, {Name: "frontend/src/foo.go"}, {Name: "frontend/tests/foo_test.go"}, {Name: "backend/tests/foo_test.go"}})
	got := []string{}
	for _, f := range sorted {
		got = append(got, f.Name)
	}
	want := []string{"backend/src/foo.go", "backend/tests/foo_test.go", "frontend/src/foo.go", "frontend/tests/foo_test.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestAmbiguousTestsRemainUnpaired(t *testing.T) {
	sorted := sortFiles([]File{{Name: "lib/foo.rb"}, {Name: "src/foo.rb"}, {Name: "test/foo_test.rb"}})
	if sorted[2].Name != "test/foo_test.rb" {
		t.Fatalf("ambiguous test paired: %v", sorted)
	}
}

func TestBuiltInConventions(t *testing.T) {
	for test, impl := range map[string]string{
		"foo.test.ts": "foo.ts", "foo.spec.tsx": "foo.tsx", "__tests__/foo.ts": "foo.ts", "tests/test_foo.py": "foo.py", "spec/foo_spec.rb": "foo.rb", "foo_test.go": "foo.go", "src/test/java/FooTest.java": "java/Foo.java", "packages/widget/tests/foo.test.js": "packages/widget/foo.js",
	} {
		got, isTest := testIdentity(test)
		if !isTest || got != impl {
			t.Errorf("%s: got %s, %v; want %s", test, got, isTest, impl)
		}
	}
}

func TestJavaSourceAndTestTrees(t *testing.T) {
	sorted := sortFiles([]File{{Name: "src/test/java/example/FooTest.java"}, {Name: "src/main/java/example/Foo.java"}, {Name: "src/main/java/example/Bar.java"}})
	if sorted[1].Name != "src/main/java/example/Foo.java" || sorted[2].Name != "src/test/java/example/FooTest.java" {
		t.Fatalf("Java pairing: %v", sorted)
	}
}
