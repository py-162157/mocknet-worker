package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	fmt.Println(get_fat_tree_port_order("s10", "s11"))
}

func get_fat_tree_port_order(pod1 string, pod2 string) int {
	k := 4

	var dst int
	flag := false
	src, err := strconv.Atoi(strings.TrimPrefix(pod1, "s"))
	if err != nil {
		panic(err)
	}
	if strings.Contains(pod2, "h") {
		flag = true
		dst, err = strconv.Atoi(strings.Split(strings.Split(pod2, "s")[0], "h")[1])
		if err != nil {
			panic(err)
		}
	} else {
		dst, err = strconv.Atoi(strings.TrimPrefix(pod2, "s"))
		if err != nil {
			panic(err)
		}
	}

	src -= 1
	dst -= 1
	fmt.Println(src, dst)

	if src < k*k/4 {
		// core
		fmt.Println("src is core")
		return (dst - k*k/4) / k
	} else {
		if (src+1)%2 != (k*k/4)%2 {
			// aggregation
			fmt.Println("src is aggregation")
			if dst < k*k/4 {
				// dst is core
				fmt.Println("dst is core")
				return k/2 + dst%(k/2)
			} else {
				// dst is access
				fmt.Println("dst is access")
				return ((dst - k*k/4) % k) / 2
			}
		} else {
			// access
			fmt.Println("src is access")
			if flag {
				// dst is host
				fmt.Println("dst is host")
				return dst
			} else {
				// dst is aggregation
				fmt.Println("dst is aggregation")
				return k/2 + ((dst-k*k/4)%k)/2
			}
		}
	}
}
