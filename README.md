# Kustomize Plugins

A set of kustomize KRM function plugins.

## Installation

To install locally:
* Clone the project
* Run `make install`
* Make sure your `${GOPATH}/bin` in in your PATH environment variable.

Alternatively, download the binaries from the latest release.

## Transformers

### Vars Transformer

A replacement for the depreciated vars kustomize transformer.

Usage:

```yaml
apiVersion: wiersma.io/v1alpha1
kind: VarsTransformer
metadata:
  name: config
  annotations:
    config.kubernetes.io/function: |-
      exec:
        path: vars-transformer
replacements:
  - name: SRC
    sources:
      - objRef: {apiVersion: v1, kind: Service, name: source}
        fieldPath: metadata.name
      - objRef: {apiVersion: v1, kind: Service, name: source}
        fieldPath: spec.ports.[name=src-port].port
    templates: "%s:%s"
  - name: SRC_CFG
    source:
      objRef: {apiVersion: v1, kind: ConfigMap, namespace: default, name: config}
      fieldPath: data.src
```
