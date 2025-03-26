package src

import (
	"fmt"
	"strings"
)

// parseMethodComment extracts source and target field names from a comment
func parseMethodComment(methodComment string) (result map[string]string, err error) {
	// one line comment may be have multiple mapmap:
	methodComment = strings.Trim(methodComment, "/")
	methodComment = strings.Trim(methodComment, " ")
	if !strings.HasPrefix(methodComment, "mapmap:") {
		return result, fmt.Errorf("method comment does not contain mapmap")
	}

	result = make(map[string]string)

	// first split by "mapmap:" get target map source items
	parts := strings.SplitSeq(methodComment, "mapmap:")
	for part := range parts {
		if part == "" {
			continue
		}

		targetAndSource := strings.Split(part, ",")

		var sourceField, targetField string
		for _, targetOrSource := range targetAndSource {
			targetOrSource = strings.TrimSpace(targetOrSource)
			if strings.HasPrefix(targetOrSource, "source:") {
				sourceField = strings.TrimPrefix(targetOrSource, "source:")
			} else if strings.HasPrefix(targetOrSource, "target:") {
				targetField = strings.TrimPrefix(targetOrSource, "target:")
			}
		}

		fmt.Println(targetField)
		fmt.Println(sourceField)

		result[targetField] = sourceField
	}

	return result, nil
}
