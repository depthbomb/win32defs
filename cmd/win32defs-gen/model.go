package main

type metadataExport struct {
	Constants       []metadataConstant       `json:"constants"`
	ConstantMethods []metadataConstantMethod `json:"constant_methods"`
	Initializers    []metadataInitializer    `json:"initializers"`
	GUIDs           []metadataGUID           `json:"guids"`
}

type metadataConstant struct {
	Namespace     string `json:"namespace"`
	DeclaringType string `json:"declaring_type"`
	Name          string `json:"name"`
	ManagedType   string `json:"managed_type"`
	NativeType    string `json:"native_type"`
	Documentation string `json:"documentation"`
	Comment       string `json:"comment"`
	Kind          string `json:"kind"`
	Value         string `json:"value"`
	Enum          bool   `json:"enum"`
}

type metadataInitializer struct {
	Namespace     string `json:"namespace"`
	DeclaringType string `json:"declaring_type"`
	Name          string `json:"name"`
	ManagedType   string `json:"managed_type"`
	Value         string `json:"value"`
}

type metadataConstantMethod struct {
	Namespace     string `json:"namespace"`
	DeclaringType string `json:"declaring_type"`
	Name          string `json:"name"`
	ReturnType    string `json:"return_type"`
	Documentation string `json:"documentation"`
	Comment       string `json:"comment"`
	Value         string `json:"value"`
}

type metadataGUID struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Value     string `json:"value"`
}

type sourcePackage struct {
	Package string `json:"package"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	URL     string `json:"url"`
}

type sourceLock struct {
	Metadata      sourcePackage `json:"metadata"`
	Documentation sourcePackage `json:"documentation"`
}

type packageSpec struct {
	Name          string
	Doc           string
	TypeName      string
	Underlying    string
	Width         uint
	Signed        bool
	Declare       bool
	Match         func(metadataConstant) bool
	CanonicalRank func(string) int
}

type generatedConstant struct {
	Name          string
	Type          string
	Expression    string
	LookupKey     string
	Numeric       bool
	Namespace     string
	DeclaringType string
	Documentation string
	Comment       string
}

type generatedConstantMethod struct {
	Name          string
	Expression    string
	Documentation string
	Comment       string
}

type generatedAuthority struct {
	Name  string
	Value [6]byte
}

type generatedZeroInitializer struct {
	Name string
}

type constantCoverage struct {
	TotalRows               int            `json:"total_rows"`
	MatchedRows             int            `json:"matched_rows"`
	UnclassifiedRows        int            `json:"unclassified_rows"`
	UnclassifiedByNamespace map[string]int `json:"unclassified_by_namespace"`
}

type collision struct {
	Package     string   `json:"package"`
	Name        string   `json:"name"`
	Definitions []string `json:"definitions"`
}

type generationReport struct {
	Source               sourceLock       `json:"source"`
	PackageCount         map[string]int   `json:"package_count"`
	GUIDCount            int              `json:"guid_count"`
	GUIDSourceCount      int              `json:"guid_source_count"`
	PropertyKeys         int              `json:"property_key_count"`
	DeviceKeys           int              `json:"device_property_key_count"`
	Documented           int              `json:"documented_constant_count"`
	Coverage             constantCoverage `json:"constant_coverage"`
	MethodCount          int              `json:"constant_method_count"`
	AuthorityCount       int              `json:"sid_identifier_authority_count"`
	SyncInitializerCount int              `json:"synchronization_initializer_count"`
	Collisions           []collision      `json:"collisions"`
	Skipped              map[string]int   `json:"skipped"`
}
