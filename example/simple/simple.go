package main

import (
	"flag"
	"fmt"
	"reflect"
	"time"

	watch "github.com/HaesungSeo/goConsulWatch"
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

	cfg, err := watch.New(*addr, *keyType, *key)
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
