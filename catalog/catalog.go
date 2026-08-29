// Package catalog provides opt-in provenance and documentation for generated
// definitions without adding lookup data to constant-only consumers.
package catalog

// Definition describes one generated constant or enum member.
type Definition struct {
	Package       string
	Name          string
	Value         string
	Namespace     string
	DeclaringType string
	Documentation string
	Comment       string
}
