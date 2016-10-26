package flamingo

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageMatchString(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	require.True(t, msg.MatchString("foo bar baz"))
	require.True(t, msg.MatchString("  foo bar baz  "))
	require.True(t, msg.MatchString("foo Bar baz"))
}

func TestMessageMatchStringCase(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	require.False(t, msg.MatchStringCase("foo Bar baz"))
	require.True(t, msg.MatchStringCase("foo Bar Baz"))
	require.True(t, msg.MatchStringCase("  foo Bar Baz  "))
}

func TestMessageMatchRegex(t *testing.T) {
	msg := Message{Text: " foo Bar Baz  "}
	r := regexp.MustCompile(`foo( Ba[rz])+`)
	r2 := regexp.MustCompile(`foo( ba[rz])+`)
	r3 := regexp.MustCompile(` foo( Ba[rz])+`)
	require.True(t, msg.MatchRegex(r))
	require.False(t, msg.MatchRegex(r2))
	require.False(t, msg.MatchRegex(r3))
}
