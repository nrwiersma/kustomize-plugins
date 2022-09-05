package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// VarsTransformer contains configuration for a vars transformer.
type VarsTransformer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Replacements []Replacement `json:"replacements" yaml:"replacements"`
}

// Replacement contains configuration for a vars replacement.
type Replacement struct {
	Name     string      `json:"name" yaml:"name"`
	Source   *SourceRef  `json:"source" yaml:"source"`
	Sources  []SourceRef `json:"sources" yaml:"sources"`
	Template string      `json:"template" yaml:"template"`
}

// SourceRef describes a source document and path.
type SourceRef struct {
	ObjRef    ObjectRef `json:"objRef" yaml:"objRef"`
	FieldPath string    `json:"fieldPath" yaml:"fieldPath"`
}

// ObjectRef describes an object.
type ObjectRef struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}
