package testutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContext creates a test context with timeout
func TestContext(t *testing.T) context.Context {
	ctx := context.Background()
	return ctx
}

// AssertNoPanic asserts that a function doesn't panic
func AssertNoPanic(t *testing.T, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function panicked: %v", r)
		}
	}()
	fn()
}

// AssertErrorContains asserts that an error contains a specific substring
func AssertErrorContains(t *testing.T, err error, substring string) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), substring)
}

// AssertNoError is a helper that fails the test if err is not nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	require.NoError(t, err, msgAndArgs...)
}

// AssertEqual is a helper that asserts two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotEmpty asserts that a value is not empty
func AssertNotEmpty(t *testing.T, value interface{}, msgAndArgs ...interface{}) {
	assert.NotEmpty(t, value, msgAndArgs...)
}

// AssertGreaterThan asserts that actual > expected
func AssertGreaterThan(t *testing.T, actual, expected int, msgAndArgs ...interface{}) {
	assert.Greater(t, actual, expected, msgAndArgs...)
}

// AssertPaginationValid asserts pagination metadata is valid
func AssertPaginationValid(t *testing.T, currentPage, totalPages, totalResults, resultsPerPage int) {
	assert.Greater(t, currentPage, 0, "current page should be positive")
	assert.GreaterOrEqual(t, totalPages, 0, "total pages should be non-negative")
	assert.GreaterOrEqual(t, totalResults, 0, "total results should be non-negative")
	assert.Greater(t, resultsPerPage, 0, "results per page should be positive")

	if totalResults > 0 {
		expectedPages := (totalResults + resultsPerPage - 1) / resultsPerPage
		assert.Equal(t, expectedPages, totalPages, "total pages calculation should be correct")
	}
}

// AssertValidUUID asserts that a string is a valid UUID
func AssertValidUUID(t *testing.T, value string) {
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, value, "should be valid UUID")
}

// TestTableCase represents a table-driven test case
type TestTableCase struct {
	Name          string
	Setup         func()
	Input         interface{}
	ExpectedError error
	ExpectedValue interface{}
	Cleanup       func()
}

// RunTableTests runs table-driven tests
func RunTableTests(t *testing.T, cases []TestTableCase, testFunc func(*testing.T, TestTableCase)) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup()
			}

			testFunc(t, tc)

			if tc.Cleanup != nil {
				tc.Cleanup()
			}
		})
	}
}

// BenchmarkTableCase represents a benchmark test case
type BenchmarkTableCase struct {
	Name  string
	Setup func()
	Bench func(b *testing.B)
}

// RunBenchmarkTests runs benchmark tests
func RunBenchmarkTests(b *testing.B, cases []BenchmarkTableCase) {
	for _, bc := range cases {
		b.Run(bc.Name, func(b *testing.B) {
			if bc.Setup != nil {
				bc.Setup()
			}
			bc.Bench(b)
		})
	}
}
