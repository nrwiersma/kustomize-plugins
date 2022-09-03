package main

import (
	"fmt"
	"os"

	"github.com/nrwiersma/kustomize-plugins/api/v1alpha1"
	"github.com/nrwiersma/kustomize-plugins/plugins/vars"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	fn := func(rl *framework.ResourceList) error {
		obj, err := rl.FunctionConfig.Map()
		if err != nil {
			return err
		}
		cfg := v1alpha1.VarsTransformer{}
		if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj, &cfg); err != nil {
			return err
		}

		plugin, err := vars.New(cfg)
		if err != nil {
			return err
		}

		rl.Items, err = plugin.Filter(rl.Items)
		return err
	}

	cmd := command.Build(framework.ResourceListProcessorFunc(fn), command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
