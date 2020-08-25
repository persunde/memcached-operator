# memcached-operator
Following operator-sdk quickstart guide as described here:
https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/


## NOTES
The file config/crd/bases/cache.example.com_webservers.yaml is not applied when running make install
Why is that?
It is the required CustomResourceDefinition that needs to be installed BEFORE the Operator can run on the C
kubectl apply -f config/crd/bases/cache.example.com_webservers.yaml


How to build and run the Operator:
export USERNAME=persundecern
make && make generate && make manifests && make install

kubectl apply -f config/crd/bases/cache.example.com_webservers.yaml

export version=v0.1.8

make docker-build IMG=$USERNAME/memcached-operator:$version && make docker-push IMG=$USERNAME/memcached-operator:$version && make deploy IMG=$USERNAME/memcached-operator:$version
