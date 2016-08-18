package flamingo

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageMatchString(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	assert.True(t, msg.MatchString("foo bar baz"))
	assert.True(t, msg.MatchString("  foo bar baz  "))
	assert.True(t, msg.MatchString("foo Bar baz"))
}

func TestMessageMatchStringCase(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	assert.True(t, msg.MatchStringCase("foo Bar Baz"))
	assert.True(t, msg.MatchStringCase("  foo Bar Baz  "))
	assert.False(t, msg.MatchStringCase("foo Bar baz"))
}

func TestMessageMatchRegex(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	r := regexp.MustCompile(`foo( Ba[rz])+`)
	r2 := regexp.MustCompile(`foo( ba[rz])+`)
	r3 := regexp.MustCompile(` foo( Ba[rz])+`)
	assert.True(t, msg.MatchRegex(r))
	assert.False(t, msg.MatchRegex(r2))
	assert.False(t, msg.MatchRegex(r3))
}
