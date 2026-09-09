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
	// Family identifies the metadata enum or the prefix of a loose constant.
	Family string
	// Flags reports whether the metadata enum has the Flags attribute.
	Flags bool
	// Kind is the original metadata scalar type, such as uint32 or string.
	Kind string
}
