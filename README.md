# gopetri

This is first implementation of petri net for Go.

```
!!! This support only one chip in progress. 
``` 

## Examples.

### Base example.

U can run example and watch changes of states in log.

```
go run example/base/main.go
```

### Base example with visualisation.

We use graphviz for visualisation. If u run next cmd, you will receive graph.dot file with dot lang and graph img.

```
go run example/base/main.go > graph.dot && dot -Tpng graph.dot > graph.png
```
