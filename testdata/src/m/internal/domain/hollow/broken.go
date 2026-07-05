// broken.go carries the .go suffix but its package clause below is
// deliberately malformed, so go/parser rejects it and the layout analyzer must
// not count this directory as hollow's domain counterpart.
package
