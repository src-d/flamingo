package flamingo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicies(t *testing.T) {
	require := require.New(t)
	p := IgnorePolicy()
	require.False(p.Reply)

	p = ReplyPolicy("foo")
	require.True(p.Reply)
	require.Equal("foo", p.Message)
}

func TestButtonGroup(t *testing.T) {
	require := require.New(t)

	g := NewButtonGroup(
		"foo",
		NewButton("a", "a"),
		NewPrimaryButton("b", "b"),
		NewDangerButton("c", "c"),
	)

	require.Equal("foo", g.ID())
	require.Equal(3, len(g.Items()))
	require.Equal(ButtonGroup, g.Type())
}

func TestImage(t *testing.T) {
	require := require.New(t)

	g := Image{}

	require.Equal("", g.ID())
	require.Equal(1, len(g.Items()))
	require.Equal(ImageGroup, g.Type())
}

func TestText(t *testing.T) {
	require := require.New(t)

	g := Text("fooooo")

	require.Equal("", g.ID())
	require.Equal(1, len(g.Items()))
	require.Equal(TextGroup, g.Type())
}

func TestTextFieldGroup(t *testing.T) {
	require := require.New(t)

	g := NewTextFieldGroup(
		NewTextField("a", "a"),
		NewTextField("b", "b"),
		NewShortTextField("c", "c"),
	)

	require.Equal("", g.ID())
	require.Equal(3, len(g.Items()))
	require.Equal(TextFieldGroup, g.Type())
}

func TestNewOutgoingMessage(t *testing.T) {
	require.Equal(t, "foo", NewOutgoingMessage("foo").Text)
}
