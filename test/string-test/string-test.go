package main

import (
	"fmt"
	"strings"
)

func main() {
	str := `
[ ID] Interval           Transfer     Bandwidth       Retr
[  4]   0.00-10.00  sec   912 MBytes   765 Mbits/sec  2079             sender
[  4]   0.00-10.00  sec   911 MBytes   764 Mbits/sec                  receiver

iperf Done.
`

	fmt.Println(str)
	var speed string
	split_str := strings.Split(str, " ")
	count := 0
	for i, a := range split_str {
		fmt.Println(a)
		if strings.Contains(a, "bits/sec") {
			count++
			// count = 1 means send rate, and 2 means receive rate
			if count == 2 {
				speed = split_str[i-1]
			}
		}
	}
	fmt.Println(speed)

}
