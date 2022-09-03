package vars

import (
	"fmt"
)

// MultipleSourcesError is returned if multiple sources were found.
type MultipleSourcesError struct {
	Name string
}

func (e MultipleSourcesError) Error() string {
	return fmt.Sprintf("found more than one source matching %s", e.Name)
}

// SourceNotFoundError is returned if a replacement source does not exist in the resource list.
type SourceNotFoundError struct {
	Name string
}

func (e SourceNotFoundError) Error() string {
	return fmt.Sprintf("no source found matching %s", e.Name)
}
