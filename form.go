package flamingo

type Form struct {
	Title   string
	Text    string
	Combine bool
	Color   string
	Fields  []FieldGroup
}

type Field interface {
	isField()
}

type FieldGroup interface {
	ID() string
	Items() []Field
	Type() FieldGroupType
}

type FieldGroupType byte

const (
	ButtonGroup FieldGroupType = 1 << iota
	TextFieldGroup
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

type ButtonType byte

const (
	DefaultButton ButtonType = iota
	PrimaryButton
	DangerButton
)

type Button struct {
	Text         string
	Name         string
	Value        string
	Type         ButtonType
	Confirmation *Confirmation
}

func (b Button) isField() {}

type Confirmation struct {
	Title   string
	Text    string
	Ok      string
	Dismiss string
}

type TextField struct {
	Title string
	Value string
	Short bool
}

func (f TextField) isField() {}
