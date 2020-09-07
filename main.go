package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var a_siprefix []string = []string{"", "k", "M", "G", "T", "P", "H"}
var a_timeprefix []string = []string{"ns", "µs", "ms", "s"}

var inc int64 = 0

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Text Prefix: ")
	text, _ := reader.ReadString('\n')
	fmt.Print("Hash Prefix: ")
	hash, _ := reader.ReadString('\n')
	fmt.Print("Avg Speed (H/S): ")
	speedS, _ := reader.ReadString('\n')

	text = text[:len(text)-1]
	if len(speedS) > 1 {
		speedS = speedS[:len(speedS)-1]
	}
	hash = hash[:len(hash)-1] // Cut \n

	lc := ([]byte(text))[len(text)-1:][0] // Check for \r
	if lc == 13 {
		text = text[:len(text)-1]
		if len(speedS) > 1 {
			speedS = speedS[:len(speedS)-1]
		}
		hash = hash[:len(hash)-1] // Cut \r
	}

	speed, err := strconv.Atoi(speedS)
	if err == nil {
		bits := len(hash) * 4
		hashes := int64(math.Pow(2, float64(bits)))
		time := int(hashes / int64(speed))
		fmt.Printf("	- Approximate time required: %s\n", formatTime(time))
	}




	ok := true
	allowed := "abcdef0123456789"
	for _, r := range hash {
		if strings.ContainsAny(string(r), allowed) != true {
			ok = false
			fmt.Println("!!! ERROR !!! \"" + strconv.Itoa(int(r)) + "\"")
		}
	}
	if ok != true {
		fmt.Println("Hash Prefix is not hexadecimal.")
		os.Exit(1)
	}

	exitC := make(chan bool)
	textC := make(chan string)

	start := time.Now().UnixNano()

	// <crunchy //

	go info(exitC)
	for i := 0; i < runtime.NumCPU(); i++ {
		go core(text, hash, i, int64(runtime.NumCPU()), exitC, textC)
	}

	// </crunchy> //

	fmt.Println("========== STARTED ===========")

	input := <-textC
	sum := sha256.Sum256([]byte(input))
	dst := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(dst, sum[:])

	end := time.Now().UnixNano()

	diff := end - start

	form1, rem := getPrefix(float64(diff), a_timeprefix)
	fmt.Println("========== FINISHED ==========")
	/*if rem > 60 {
		form1 = formatTime(int64(rem))
		rem = math.Mod(math.Floor(rem), 60)
	}*/

	fmt.Printf("The operation took %.3f%s\n", rem, form1)
	fmt.Println("Results:")
	fmt.Println("	- Input: " + input + "\\n")
	fmt.Println("	- Hash:  " + string(dst))
}

func formatTime(a int) string {
	seconds := a
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	seconds  = seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	hours   := minutes / 60
	minutes = minutes % 60
	if hours < 24 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	days  := hours / 24
	hours  = hours % 24
	if hours < 365 {
		return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, seconds)
	}
	years := days / 365
	days   = years % 365
	return fmt.Sprintf("%dy%dd%dh%dm%ds", years, days, hours, minutes, seconds)
}

/*func formatTime(a int64) string {
	seconds := a

	minutes := seconds / 60
	seconds = seconds % 60


	if minutes > 60 {
		hours := minutes / 60
		minutes = minutes % 60
		return fmt.Sprintf(" seconds, %d minutes and %d hours", minutes, hours)
	} else {
		return fmt.Sprintf(" seconds and %d minutes", minutes)
	}
}*/

func getPrefix(a float64, ma []string) (string, float64) {
	i := 0
	for {
		if a < 1000 {
			return ma[i], a
		}
		a /= 1000
		i++
		if i > len(ma) {
			return ma[i-1], a
		}
	}
}

func info(e chan bool) {
	for {
		select {
		case <-e:
			return
		default:
			ae := time.Now().UnixNano()
			time.Sleep(time.Second)
			ae = time.Now().UnixNano() - ae

			prefix, rem := getPrefix(float64(inc) / (float64(ae) / 1000000), a_siprefix)
			prefix2, rem2 := getPrefix(math.Abs(float64(ae) - 1000000000), a_timeprefix)
			fmt.Printf(  "%.2f%s hashes per second. (S.I.: %.3f%s)\n", rem, prefix, rem2, prefix2)
			inc = int64(0)
		}
	}

}

func core(in string, ou string, i int, m int64, e chan bool, r chan string) {
	var off int64 = int64(i)
	for {
		select {
		case <-e:
			return
		default:
			txt := in + " " + strconv.FormatInt(off, 16) + "\n"

			sum := sha256.Sum256([]byte(txt))

			dst := make([]byte, hex.EncodedLen(len(sum)))
			hex.Encode(dst, sum[:])

			sums := string(dst)

			if strings.HasPrefix(sums, ou) {
				r <- txt
				for xd := 0; xd < int(m); xd++ {
					e <- true
				}
			}

			if (off % (m * 64)) == 0 {
				inc += int64(off / m)
			}

			off += m
		}
	}
}
