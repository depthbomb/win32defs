package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

func emittedSymbols(packages map[string][]generatedConstant) map[string][]string {
	result := make(map[string][]string, len(packages))
	for packageName, constants := range packages {
		for _, constant := range constants {
			result[packageName] = append(result[packageName], constant.Name)
		}
		sort.Strings(result[packageName])
	}

	return result
}

func previousSymbols(root string) (map[string][]string, error) {
	result := make(map[string][]string)
	path := filepath.Join(root, "catalog", "zz_generated.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	}

	if err != nil {
		return nil, err
	}

	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}

		var packageName, name string
		for _, element := range literal.Elts {
			field, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := field.Key.(*ast.Ident)
			if !ok {
				continue
			}

			value, ok := field.Value.(*ast.BasicLit)
			if !ok || value.Kind != token.STRING {
				continue
			}

			text, _ := strconv.Unquote(value.Value)
			switch key.Name {
			case "Package":
				packageName = text
			case "Name":
				name = text
			}
		}

		if packageName != "" && name != "" {
			result[packageName] = append(result[packageName], name)
		}

		return true
	})
	for key := range result {
		sort.Strings(result[key])
		result[key] = slices.Compact(result[key])
	}

	return result, nil
}

func validateGeneration(root string, report generationReport, acceptChanges bool) error {
	if acceptChanges {
		return nil
	}

	var previous generationReport
	contents, err := os.ReadFile(filepath.Join(root, "internal", "source", "report.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err == nil {
		if err := json.Unmarshal(contents, &previous); err != nil {
			return fmt.Errorf("read baseline generation report: %w", err)
		}
	}

	if previous.Symbols == nil {
		previous.Symbols, err = previousSymbols(root)
		if err != nil {
			return err
		}
	}

	var failures []string
	for packageName, names := range previous.Symbols {
		for _, name := range names {
			if _, exists := slices.BinarySearch(report.Symbols[packageName], name); !exists {
				failures = append(failures, "removed symbol: "+packageName+"."+name)
			}
		}
	}

	for _, item := range report.Collisions {
		encoded, _ := json.Marshal(item)
		known := slices.ContainsFunc(previous.Collisions, func(old collision) bool {
			oldEncoded, _ := json.Marshal(old)

			return string(encoded) == string(oldEncoded)
		})
		if !known {
			failures = append(failures, "new collision: "+string(encoded))
		}
	}

	for _, item := range report.Rejected {
		if !slices.Contains(previous.Rejected, item) {
			encoded, _ := json.Marshal(item)
			failures = append(failures, "rejected definition: "+string(encoded))
		}
	}

	for kind, count := range report.Skipped {
		if count > previous.Skipped[kind] {
			failures = append(failures, fmt.Sprintf("new omissions: %s increased from %d to %d", kind, previous.Skipped[kind], count))
		}
	}

	if len(failures) != 0 {
		sort.Strings(failures)

		return fmt.Errorf("generation validation failed; live output was not changed:\n%s\nUse -accept-changes only after reviewing these changes", strings.Join(failures, "\n"))
	}

	return nil
}

func rejectedStructuredDefinitions(export metadataExport) []rejectedDefinition {
	var rejected []rejectedDefinition
	for _, item := range export.GUIDs {
		if _, _, skipped := collectGUIDs([]metadataGUID{item}); skipped != 0 {
			rejected = append(rejected, rejectedDefinition{
				Package:   "guid",
				Namespace: item.Namespace,
				Name:      item.Name,
				Value:     item.Value,
				Reason:    "invalid GUID value",
			})
		}
	}

	for _, item := range export.ConstantMethods {
		if _, skipped := collectConstantMethods([]metadataConstantMethod{item}); skipped != 0 {
			rejected = append(rejected, rejectedDefinition{
				Package:   "process",
				Namespace: item.Namespace,
				Name:      item.Name,
				Value:     item.Value,
				Reason:    "unsupported constant-returning method",
			})
		}
	}

	seenKeys := make(map[string]string)
	for _, item := range export.Initializers {
		packageName, reason := "", ""
		switch {
		case hasAnySuffix(item.ManagedType, ".PROPERTYKEY", ".DEVPROPKEY"):
			packageName = "propertykey"
			if strings.HasSuffix(item.ManagedType, ".DEVPROPKEY") {
				packageName = "devpropkey"
			}

			parts, id, err := parsePropertyKey(item.Value)
			key := packageName + "." + sanitizeIdentifier(item.Name)
			value := fmt.Sprintf("%v/%d", parts, id)
			if err != nil {
				reason = err.Error()
			} else if old, exists := seenKeys[key]; exists && old != value {
				reason = "conflicting structured key definition"
			} else {
				seenKeys[key] = value
			}
		case strings.HasSuffix(item.ManagedType, ".SID_IDENTIFIER_AUTHORITY"):
			packageName = "security"
			if _, skipped := collectAuthorities([]metadataInitializer{item}); skipped != 0 {
				reason = "invalid SID authority initializer"
			}
		case hasAnySuffix(item.ManagedType, ".CONDITION_VARIABLE", ".INIT_ONCE", ".SRWLOCK"):
			packageName = "syncinit"
			if _, skipped := collectSyncInitializers([]metadataInitializer{item}); skipped != 0 {
				reason = "unsupported nonzero synchronization initializer"
			}
		}

		if reason != "" {
			rejected = append(rejected, rejectedDefinition{
				Package:   packageName,
				Namespace: item.Namespace,
				Name:      item.Name,
				Value:     item.Value,
				Reason:    reason,
			})
		}
	}

	return rejected
}
