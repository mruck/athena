package util

import "strings"

// GET for a spec.operation
const GET = "get"

// DELETE for a spec.operation
const DELETE = "delete"

// POST for a spec.operation
const POST = "post"

// PATCH for a spec.operation
const PATCH = "patch"

// PUT for a spec.operation
const PUT = "put"

// HEAD for a spec.operation
const HEAD = "head"

// CompareMethods compares lowercased http method
func CompareMethods(method1 string, method2 string) bool {
	return strings.EqualFold(method1, method2)
}
