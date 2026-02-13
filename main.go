package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {

	strictFlag := flag.Bool("strict", false, "enable strict mode")
	flag.Parse()
	config := Config{
		Strict: *strictFlag,
	}

	kvMap := make(map[string]string)
	file, err := os.Open(".env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// syntax check
		k, v, err := parseLine(line, lineNum, config)
		if err != nil {
			fmt.Println(err, v)
			break
		}

		// skip comments or empty lines
		if k == "" {
			continue
		} else {
			// // detecting duplicate keys
			// if _, exists := kvMap[k]; exists {
			// 	fmt.Printf("duplicate key in line %d\n", lineNum)
			// } else {
			// 	kvMap[k] = v
			// }
		}

		// fmt.Println("_____________")
	}

	for k, v := range kvMap {
		fmt.Printf("> %s ={ %s }\n", k, v)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func parseLine(line string, lineNum int, cfg Config) (string, string, error) {

	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return "", "", nil
	}

	// check for comment line
	if strings.Contains(line, "#") {
		if strings.Compare(string(line[0]), "#") == 0 {
			log.Printf("comment line %d", lineNum)
			return "", "", nil
		}
	}
	// check for =
	if !strings.Contains(line, "=") {
		fmt.Println("Line number", lineNum, "error: missing =")
		return "", "", fmt.Errorf("missing '=' on line %d", lineNum)
	}
	// im guaranteed two item in the slice
	slice := strings.SplitN(line, "=", 2)

	// checks the structural format of keys
	k, v, err := checkAfterSplit(slice, lineNum, cfg)
	if err != nil {
		return "", "", err
	}

	return k, v, nil
}

func checkAfterSplit(slice []string, lineNum int, cfg Config) (string, string, error) {
	key := string(slice[0])
	value := string(slice[1])

	if len(key) == 0 {
		return "", "", fmt.Errorf("missing {key} on line %d", lineNum)
	}
	if len(value) == 0 {
		log.Printf("missing {value} on line %d", lineNum)
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if cfg.Strict {
		isCleaKey, err := StrictKeyRegex(key)
		if err != nil || !isCleaKey {
			return "", "", fmt.Errorf("key does not respect naming format on line %d", lineNum)
		}

	}

	return key, value, nil
}

func StrictKeyRegex(word string) (bool, error) {

	matched, err := regexp.MatchString("^[A-Z_][A-Z0-9_]*$", word)
	if err != nil {
		return true, err
	}
	return matched, nil
}
