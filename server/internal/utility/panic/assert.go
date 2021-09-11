// Package panic provides the condition check.
package panic

// Assert throws a panic if the expression is not true.
func Assert(exp bool, msg interface{}) {
	if exp != true {
		panic(msg)
	}
}
