# memcached-operator
Following operator-sdk quickstart guide as described here:
https://sdk.operatorframework.io/docs/golang/quickstart/


## NOTES
The file config/crd/bases/cache.example.com_webservers.yaml is not applied when running make install
Why is that?
It is the required CustomResourceDefinition that needs to be installed BEFORE the Operator can run on the C
kubectl apply -f config/crd/bases/cache.example.com_webservers.yaml