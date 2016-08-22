package flamingo

// Form is the way to post complex structures with the bot.
// Forms can have title, text, color, footer, several Fields
// (buttons, text fields, images, ...), etc.
type Form struct {
	// Title is the title of the form.
	Title string
	// Text is the text of the form.
	Text string
	// Combine will force all the form to be in a single message.
	// The behavior of this is actually client-dependant.
	Combine bool
	// Color is the color of the form. This is client-specific.
	Color string
	// Footer is a small text on the footer of the form.
	Footer string
	// Fields is a collection of field groups on the form.
	Fields []FieldGroup
}

// Field is an element of a FieldGroup.
type Field interface {
	isField()
}

// FieldGroup is a collection of Fields of a concrete type.
type FieldGroup interface {
	// ID returns the ID of the group (only valid for ButtonGroup)
	ID() string
	// Items returns all the fields in the group.
	Items() []Field
	// Type returns the type of the group.
	Type() FieldGroupType
}

// FieldGroupType is the type of elements contained in the group.
type FieldGroupType byte

const (
	// ButtonGroup is a group of buttons.
	ButtonGroup FieldGroupType = 1 << iota
	// TextFieldGroup is a group of text fields.
	TextFieldGroup
	// ImageGroup is a single image.
	ImageGroup
)

type fieldGroup struct {
	kind  FieldGroupType
	id    string
	items []Field
}

func (g *fieldGroup) Items() []Field {
	return g.items
}

func (g *fieldGroup) ID() string {
	return g.id
}

func (g *fieldGroup) Type() FieldGroupType {
	return g.kind
}

// NewButtonGroup creates a FieldGroup with the given buttons.
func NewButtonGroup(id string, buttons ...Button) FieldGroup {
	var items []Field
	for _, b := range buttons {
		items = append(items, b)
	}
	return &fieldGroup{
		kind:  ButtonGroup,
		id:    id,
		items: items,
	}
}

// NewTextFieldGroup creates a FieldGroup with the given text fields.
func NewTextFieldGroup(fields ...TextField) FieldGroup {
	var items []Field
	for _, f := range fields {
		items = append(items, f)
	}
	return &fieldGroup{
		kind:  TextFieldGroup,
		items: items,
	}
}

// ButtonType is the kind of button. Even though there are several
// types of buttons this is really very platform-specific.
type ButtonType byte

const (
	// DefaultButton is the default button.
	DefaultButton ButtonType = iota
	// PrimaryButton is a button with an accent color to stand out.
	PrimaryButton
	// DangerButton is a typically red button to indicate a dangerous action.
	DangerButton
)

// Button is a single button element.
type Button struct {
	// Text is the visible text of the button.
	Text string
	// Name is the name of the button.
	Name string
	// Value is the value of the button.
	Value string
	// Type is the type of the button.
	Type ButtonType
	// Confirmation defines a confirmation popup to be shown after the button
	// is clicked. This is platform-specific behaviour.
	Confirmation *Confirmation
}

// NewButton creates a new default button. Value and name will be the same, if
// that does not work for you, you will need to create a Button by hand.
func NewButton(text, value string) Button {
	return Button{
		Text:  text,
		Value: value,
		Name:  value,
	}
}

// NewPrimaryButton creates a new primary button. Value and name will be the same, if
// that does not work for you, you will need to create a Button by hand.
func NewPrimaryButton(text, value string) Button {
	b := NewButton(text, value)
	b.Type = PrimaryButton
	return b
}

// NewDangerButton creates a new danger button. Value and name will be the same, if
// that does not work for you, you will need to create a Button by hand.
func NewDangerButton(text, value string) Button {
	b := NewButton(text, value)
	b.Type = DangerButton
	return b
}

func (b Button) isField() {}

// Confirmation is a confirmation popup to be shown after a button is clicked.
// Slack-specific at the moment.
type Confirmation struct {
	// Title of the confirm window.
	Title string
	// Text of the confirm window.
	Text string
	// Ok button text of the confirm window.
	Ok string
	// Dismiss button text of the confirm window.
	Dismiss string
}

// TextField is a field which displays a label with its text value.
type TextField struct {
	// Title is the label of the field.
	Title string
	// Value is the text value of the field.
	Value string
	// Short is a platform-specific feature. If available, will be rendered as
	// a short field instead of a wide one.
	Short bool
}

// NewTextField creates a new textfield.
func NewTextField(title, value string) TextField {
	return TextField{
		Title: title,
		Value: value,
	}
}

// NewShortTextField creates a new short text field.
func NewShortTextField(title, value string) TextField {
	return TextField{
		Title: title,
		Value: value,
		Short: true,
	}
}

func (f TextField) isField() {}

// Image is an image to be posted.
type Image struct {
	// URL is the URL of the image.
	URL string
	// Text of the image.
	Text string
	// ThumbnailURL is the URL of the thumbnail to be displayed.
	ThumbnailURL string
}

func (f Image) isField() {}

// ID always returns an empty string since Images have no ID. This method is just
// implemented to fulfill the FieldGroup interface.
func (f Image) ID() string { return "" }

// Items returns a slice with only the image itself, which is both a fieldgroup and
// a field.
func (f Image) Items() []Field { return []Field{f} }

// Type returns the ImageGroup type.
func (f Image) Type() FieldGroupType { return ImageGroup }
