package flamingo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicies(t *testing.T) {
	assert := assert.New(t)
	p := IgnorePolicy()
	assert.False(p.Reply)

	p = ReplyPolicy("foo")
	assert.True(p.Reply)
	assert.Equal("foo", p.Message)
}

func TestButtonGroup(t *testing.T) {
	assert := assert.New(t)

	g := NewButtonGroup(
		"foo",
		NewButton("a", "a"),
		NewPrimaryButton("b", "b"),
		NewDangerButton("c", "c"),
	)

	assert.Equal("foo", g.ID())
	assert.Equal(3, len(g.Items()))
	assert.Equal(ButtonGroup, g.Type())
}

func TestImage(t *testing.T) {
	assert := assert.New(t)

	g := Image{}

	assert.Equal("", g.ID())
	assert.Equal(1, len(g.Items()))
	assert.Equal(ImageGroup, g.Type())
}

func TestTextFieldGroup(t *testing.T) {
	assert := assert.New(t)

	g := NewTextFieldGroup(
		NewTextField("a", "a"),
		NewTextField("b", "b"),
		NewShortTextField("c", "c"),
	)

	assert.Equal("", g.ID())
	assert.Equal(3, len(g.Items()))
	assert.Equal(TextFieldGroup, g.Type())
}

func TestNewOutgoingMessage(t *testing.T) {
	assert.Equal(t, "foo", NewOutgoingMessage("foo").Text)
}
