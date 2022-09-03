package vars_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/nrwiersma/kustomize-plugins/api/v1alpha1"
	"github.com/nrwiersma/kustomize-plugins/plugins/vars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     v1alpha1.VarsTransformer
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "accepts valid config",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{
					{
						Name: "test1",
						Source: &v1alpha1.SourceRef{
							ObjRef:    v1alpha1.ObjectRef{APIVersion: "v1", Kind: "ConfigMap"},
							FieldPath: "data.test",
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "handles no replacements",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{},
			},
			wantErr: assert.Error,
		},
		{
			name: "handles no name",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{
					{
						Source: &v1alpha1.SourceRef{
							ObjRef:    v1alpha1.ObjectRef{APIVersion: "v1", Kind: "ConfigMap"},
							FieldPath: "data.test",
						},
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "handles no source and sources",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{
					{
						Name: "test",
					},
				},
			},
			wantErr: assert.Error,
		},
		{
			name: "handles source and sources",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{
					{
						Name: "test",
						Source: &v1alpha1.SourceRef{
							ObjRef:    v1alpha1.ObjectRef{APIVersion: "v1", Kind: "ConfigMap"},
							FieldPath: "data.test",
						},
						Sources: []v1alpha1.SourceRef{
							{
								ObjRef:    v1alpha1.ObjectRef{APIVersion: "v1", Kind: "ConfigMap"},
								FieldPath: "data.test",
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},

		{
			name: "handles sources and no tempalte",
			cfg: v1alpha1.VarsTransformer{
				Replacements: []v1alpha1.Replacement{
					{
						Name: "test",
						Sources: []v1alpha1.SourceRef{
							{
								ObjRef:    v1alpha1.ObjectRef{APIVersion: "v1", Kind: "ConfigMap"},
								FieldPath: "data.test",
							},
						},
					},
				},
			},
			wantErr: assert.Error,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := vars.New(test.cfg)

			test.wantErr(t, err)
		})
	}
}

func TestPlugin_FilterSource(t *testing.T) {
	in, err := os.ReadFile("testdata/src.yaml")
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/want.src.yaml")
	require.NoError(t, err)

	cfg := v1alpha1.VarsTransformer{
		Replacements: []v1alpha1.Replacement{
			{
				Name: "TEST",
				Source: &v1alpha1.SourceRef{
					ObjRef: v1alpha1.ObjectRef{
						APIVersion: "v1",
						Kind:       "ConfigMap",
						Namespace:  "test",
						Name:       "config",
					},
					FieldPath: "data.src",
				},
			},
		},
	}

	plugin, err := vars.New(cfg)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewReader(in)}},
		Filters: []kio.Filter{plugin},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: buf}},
	}
	err = p.Execute()

	require.NoError(t, err)
	assert.Equal(t, string(want), buf.String())
}

func TestPlugin_FilterSourceHandlesNonExistentSource(t *testing.T) {
	in, err := os.ReadFile("testdata/src.yaml")
	require.NoError(t, err)

	cfg := v1alpha1.VarsTransformer{
		Replacements: []v1alpha1.Replacement{
			{
				Name: "TEST",
				Source: &v1alpha1.SourceRef{
					ObjRef: v1alpha1.ObjectRef{
						APIVersion: "v1",
						Kind:       "ConfigMap",
						Name:       "test",
					},
					FieldPath: "data.src",
				},
			},
		},
	}

	plugin, err := vars.New(cfg)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewReader(in)}},
		Filters: []kio.Filter{plugin},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: buf}},
	}
	err = p.Execute()

	assert.Error(t, err)
}

func TestPlugin_FilterSources(t *testing.T) {
	in, err := os.ReadFile("testdata/srcs.yaml")
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/want.srcs.yaml")
	require.NoError(t, err)

	cfg := v1alpha1.VarsTransformer{
		Replacements: []v1alpha1.Replacement{
			{
				Name: "TEST",
				Sources: []v1alpha1.SourceRef{
					{
						ObjRef: v1alpha1.ObjectRef{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "source",
						},
						FieldPath: "metadata.name",
					},
					{
						ObjRef: v1alpha1.ObjectRef{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "source",
						},
						FieldPath: "spec.ports.[name=src-port].port",
					},
				},
				Template: "%s:%s",
			},
		},
	}

	plugin, err := vars.New(cfg)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewReader(in)}},
		Filters: []kio.Filter{plugin},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: buf}},
	}
	err = p.Execute()

	require.NoError(t, err)
	assert.Equal(t, string(want), buf.String())
}

func TestPlugin_FilterSourcesHandlesNonExistentSource(t *testing.T) {
	in, err := os.ReadFile("testdata/srcs.yaml")
	require.NoError(t, err)

	cfg := v1alpha1.VarsTransformer{
		Replacements: []v1alpha1.Replacement{
			{
				Name: "TEST",
				Sources: []v1alpha1.SourceRef{
					{
						ObjRef: v1alpha1.ObjectRef{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "test",
						},
						FieldPath: "metadata.name",
					},
					{
						ObjRef: v1alpha1.ObjectRef{
							APIVersion: "v1",
							Kind:       "Service",
							Name:       "source",
						},
						FieldPath: "spec.ports.[name=src-port].port",
					},
				},
				Template: "%s:%s",
			},
		},
	}

	plugin, err := vars.New(cfg)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewReader(in)}},
		Filters: []kio.Filter{plugin},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: buf}},
	}
	err = p.Execute()

	assert.Error(t, err)
}
