package ast

type Template struct {
	Body Policy
}

type LinkedPolicy struct {
	TemplateID string
	LinkID     string
	Policy     *Policy
}
