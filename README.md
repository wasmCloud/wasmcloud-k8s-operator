# wasmCloud Kubernetes Operator
The wasmCloud Kubernetes Operator is an [operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) used for managing a "wasmCloud application". This type of operator work is typically done through the use of [CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)s. The wasmCloud Kubernetes Operator aims to provide a thin facade atop of the [lattice controller](https://github.com/wasmCloud/lattice-controller) to give people running Kubernetes the ability to manage robust, declarative deployments of capability providers, actors, and their respective configurations.
