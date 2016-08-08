package flamingo

type Form struct {
	Title    string
	ID       string
	Subtitle string
	Fields   []FieldGroup
}

type Field interface {
	isField()
}

type FieldGroup interface {
	Items() []Field
}

type fieldGroup struct {
	items []Field
}

func (g *fieldGroup) Items() []Field {
	return g.items
}

func NewButtonGroup(buttons ...Button) FieldGroup {
	var items []Field
	for _, b := range buttons {
		items = append(items, b)
	}
	return &fieldGroup{
		items: items,
	}
}

func NewTextFieldGroup(fields ...TextField) FieldGroup {
	var items []Field
	for _, f := range fields {
		items = append(items, f)
	}
	return &fieldGroup{
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
