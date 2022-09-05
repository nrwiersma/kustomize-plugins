package vars

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nrwiersma/kustomize-plugins/api/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/utils"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
)

type plugin struct {
	repls []v1alpha1.Replacement
}

// New return a vars transformer.
func New(cfg v1alpha1.VarsTransformer) (kio.Filter, error) {
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return &plugin{
		repls: cfg.Replacements,
	}, nil
}

func validate(cfg v1alpha1.VarsTransformer) error {
	if len(cfg.Replacements) == 0 {
		return errors.New("at least one config replacement is required")
	}
	for _, repl := range cfg.Replacements {
		if repl.Name == "" {
			return errors.New("a replacement must have a name")
		}
		if repl.Source == nil && len(repl.Sources) == 0 {
			return fmt.Errorf("replacement %s must have a source or sources", repl.Name)
		}
		if repl.Source != nil && len(repl.Sources) != 0 {
			return fmt.Errorf("replacement %s may not have a source and sources", repl.Name)
		}
		if len(repl.Sources) != 0 && repl.Template == "" {
			return fmt.Errorf("replacement %s has sources but no template", repl.Name)
		}
	}
	return nil
}

func (p *plugin) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	vals := map[string]string{}
	for _, repl := range p.repls {
		val, err := getValue(repl, items)
		if err != nil {
			return nil, err
		}
		vals["$("+repl.Name+")"] = val
	}

	for _, item := range items {
		walker := walk.Walker{
			Visitor:               replacer{vals: vals},
			Sources:               []*yaml.RNode{item},
			InferAssociativeLists: false,
			VisitKeysAsScalars:    false,
		}
		if _, err := walker.Walk(); err != nil {
			return nil, err
		}
	}

	return items, nil
}

func getValue(repl v1alpha1.Replacement, items []*yaml.RNode) (string, error) {
	if repl.Source != nil {
		return getSourceValue(repl.Name, *repl.Source, items)
	}

	vals := make([]any, 0, len(repl.Sources))
	for _, src := range repl.Sources {
		val, err := getSourceValue(repl.Name, src, items)
		if err != nil {
			return "", err
		}
		vals = append(vals, val)
	}
	return fmt.Sprintf(repl.Template, vals...), nil
}

func getSourceValue(name string, src v1alpha1.SourceRef, items []*yaml.RNode) (string, error) {
	sel := framework.Selector{}
	if v := src.ObjRef.APIVersion; v != "" {
		sel.APIVersions = append(sel.APIVersions, v)
	}
	if v := src.ObjRef.Kind; v != "" {
		sel.Kinds = append(sel.Kinds, v)
	}
	if v := src.ObjRef.Namespace; v != "" {
		sel.Namespaces = append(sel.Namespaces, v)
	}
	if v := src.ObjRef.Name; v != "" {
		sel.Names = append(sel.Names, v)
	}
	res, err := sel.Filter(items)
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "", SourceNotFoundError{Name: name}
	}
	if len(res) > 1 {
		return "", MultipleSourcesError{Name: name}
	}

	paths := utils.SmarterPathSplitter(src.FieldPath, ".")
	rn, err := res[0].Pipe(yaml.Lookup(paths...))
	if err != nil {
		return "", err
	}
	if rn.YNode().Kind != yaml.ScalarNode {
		return "", PathNotFoundError{Name: name}
	}
	return rn.YNode().Value, nil
}

type replacer struct {
	vals map[string]string
}

func (v replacer) VisitMap(srcs walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	return srcs[0], nil
}

func (v replacer) VisitScalar(srcs walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	src := srcs[0]
	if !src.IsStringValue() {
		return src, nil
	}

	val := src.YNode().Value
	for k, v := range v.vals {
		val = strings.ReplaceAll(val, k, v)
	}
	src.YNode().Value = val
	return src, nil
}

func (v replacer) VisitList(srcs walk.Sources, _ *openapi.ResourceSchema, kind walk.ListKind) (*yaml.RNode, error) {
	return srcs[0], nil
}
