# goConsulWatch
golang Consul Subscription example with using consul watch

# import in your project
```go
import (
    "github.com/HaesungSeo/goConsulWatch"
)
```

# example code
## simple 
```go
package main

import (
    "flag"
    "fmt"
    "reflect"
    "time"

    goConsulWatch "github.com/HaesungSeo/goConsulWatch"
)

func isMapSame(src, dst map[string]string) bool {
    return reflect.DeepEqual(src, dst)
}

func diff(src, dst map[string]string) map[string]string {
    differ := make(map[string]string)
    for k, val1 := range src {
        val2, ok := dst[k]
        if !ok {
            differ[k] = fmt.Sprintf("[%s] removed", k)
        }
        if val1 != val2 {
            differ[k] = fmt.Sprintf("[%s] [%s] -->[%s]", k, val1, val2)
        }
    }
    for k, val2 := range dst {
        _, ok := src[k]
        if !ok {
            differ[k] = fmt.Sprintf("[%s] New [%s]", k, val2)
        }
        // maybe already compared
    }
    return differ
}

// Watch example
func main() {
    addr := flag.String("s", ":8080", "consul server address")
    keyType := flag.String("t", "keyprefix", "consul watch type, one of key keyprefix")
    key := flag.String("k", "/", "consul kv key")

    flag.Parse()

    cfg, err := goConsulWatch.NewCache(*addr, *keyType, *key)
    if err != nil {
        fmt.Printf("ERROR: %s\n", err)
        return
    }

    old := cfg.KVCopy()
    fmt.Printf("%s\n", old)
    for {
        select {
        case <-time.After(1 * time.Second):
            cur := cfg.KVCopy()
            if !isMapSame(cur, old) {
                diff := diff(old, cur)
                for _, v := range diff {
                    fmt.Printf("%s\n", v)
                }
            }
            old = cur
        }
    }
}
```

with two terminal, Run binary simple at first terminal, <BR> 
then Run consul cli at the other one. watch the result

first terminal
```bash
./simple -k foo -t key
map[]
[foo] New [hello]
[foo] [hello] -->[world]
[foo] [world] -->[and goodbye]
[foo] [and goodbye] -->[]
[foo] New [again]
```

second terminal
```bash
$ consul kv put foo hello
Success! Data written to: foo
$ consul kv put foo world
Success! Data written to: foo
$ consul kv put foo "and goodbye"
Success! Data written to: foo
$ consul kv delete foo
Success! Deleted key: foo
$ consul kv put foo again
Success! Data written to: foo
$
```
