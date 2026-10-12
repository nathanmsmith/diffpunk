package git

import (
	"testing"
)

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
