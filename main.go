package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

//global variables
//var valuesForFormatting bool = false
var tmpYamlText []string
var envIndent int = -5555
var currentIndent int = -77777

const AES = "AES256:"

func main() {
	flagKey := flag.String(
		"key",
		"}tf&Wr+Nt}A9g{s",
		"AES key for encrypt/decrypt",
	)
	key := *flagKey

	flagFile := flag.String(
		"filename",
		"values.yaml",
		"filename for encode/decode",
	)
	filename := *flagFile

	flagEnv := flag.String("env", "env:", "block-name for encode/decode")
	env := *flagEnv

	flag.Parse()

	// disable timestamp in stdout
	log.SetFlags(0)

	tetragonal := 333
	value := "test"
	if tetragonal == 1 {
		encryptOneValue(key, value)
		decryptOneValue(key, value)
	}

	// read file
	text := readFile(filename)
	// calculate indents for each line in YAML file
	for _, eachLn := range text {

		// current indent
		currentIndent = countIndent(eachLn)
		//log.Println(currentIndent)

		// check if current line is env block
		if matchEnvBlock(strings.TrimSpace(eachLn), env) {
			envIndent = currentIndent
			log.Println(eachLn)
			tmpYamlText = append(tmpYamlText, eachLn)
			continue
		}

		if len(eachLn) != 0 || matchCharacter(strings.TrimSpace(eachLn), "#") {
			if currentIndent == envIndent+2 {
				parsedString := parseEachLine(eachLn, key)
				log.Println(parsedString)
				tmpYamlText = append(tmpYamlText, parsedString)
			} else {
				envIndent = -5645
				log.Println(eachLn)
				tmpYamlText = append(tmpYamlText, eachLn)

			}
		}

		//check if next line after env block with new indent +2

	}

	//if already ok, read temp yaml slice and rewrite target yaml file
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed open file: %s", err)
	}
	datawriter := bufio.NewWriter(file)
	for _, data := range tmpYamlText {
		_, _ = datawriter.WriteString(data + "\n")
	}
	datawriter.Flush()
	file.Close()
}

func parseEachLine(eachLn string, key string) string {
	var parsedLine string

	// disable timestamp in stdout
	log.SetFlags(0)

	// current indent
	currentIndent = countIndent(eachLn)
	// split string to array
	stringArray := strings.Fields(eachLn)
	// concatenate whitespaces
	whitespaces := strings.Repeat(" ", currentIndent)

	// skip if line is empty
	if len(eachLn) == 0 {
		parsedLine = eachLn
		return parsedLine
	}

	// skip if line is comment, started with #
	if matchCharacter(strings.TrimSpace(eachLn), "#") {
		parsedLine = eachLn
		return parsedLine
	}

	// skip if Value is empty and not contains quotes
	if len(stringArray) == 1 {
		parsedLine = eachLn
		return parsedLine
	}

	// skip if Value is empty and contains quotes
	if len(stringArray) == 2 && stringArray[1] == "\"\"" {
		parsedLine = eachLn
		return parsedLine
	}

	// convert if Value is not empty, but contains quote
	if len(stringArray) >= 2 && stringArray[1] != "\"\"" && matchContains(strings.TrimSpace(eachLn), "\"") {
		regexTemplate := regexp.MustCompile(`"[^"]+"`)
		oldValueString := strings.Join(regexTemplate.FindAllString(eachLn, 1), "")
		encryptedValue, err := encryptAES(key, oldValueString)
		newValue := AES + encryptedValue
		if err != nil {
			log.Fatalf("something wrong, cannot encrypt")
		}
		stringReplaced := strings.ReplaceAll(eachLn, oldValueString, newValue)
		parsedLine = stringReplaced
		return parsedLine
	}

	if len(stringArray) >= 2 && stringArray[1] != "\"\"" && !matchContains(strings.TrimSpace(eachLn), "\"") {

		encryptedValue, err := encryptAES(key, stringArray[1])
		if err != nil {
			log.Fatalf("something wrong, cannot encrypt")
		}
		stringArray[1] = AES + encryptedValue
		parsedLine = whitespaces + strings.Join(stringArray[:], " ")
		return parsedLine
	}
	return parsedLine
}

func encryptOneValue(key string, value string) string {
	//for @jaxel87, encrypt/decrypt value by flag without encrypt/decrypt file
	encrypted, err := encryptAES(key, value)
	if err != nil {
		log.Fatalf("something wrong during encrypt")
	}
	return encrypted
}

func decryptOneValue(key string, value string) string {
	//for @jaxel87, encrypt/decrypt value by flag without encrypt/decrypt file
	decrypted, err := decryptAES(key, value)
	fmt.Println(decrypted)
	if err != nil {
		log.Fatalf("something wrong during decrypt")
	}
	return decrypted
}

// calculate indents for line
func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// check if line is env block
func matchEnvBlock(line string, env string) bool {
	if strings.HasPrefix(line, env) {
		return true
	} else {
		return false
	}
}

// match if line with character
func matchCharacter(line string, character string) bool {
	if strings.HasPrefix(line, character) {
		return true
	} else {
		return false
	}
}

// match if line contains quotes
func matchContains(line string, character string) bool {
	if strings.Contains(line, character) {
		return true
	} else {
		return false
	}
}

func readFile(filename string) (text []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open file")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()
	return text

}

// helm native function
func encryptAES(password string, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := make([]byte, 32)
	copy(key, []byte(password))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	content := []byte(plaintext)
	blockSize := block.BlockSize()
	padding := blockSize - len(content)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	content = append(content, padtext...)

	ciphertext := make([]byte, aes.BlockSize+len(content))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], content)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// helm native function
func decryptAES(password string, crypt64 string) (string, error) {
	if crypt64 == "" {
		return "", nil
	}

	key := make([]byte, 32)
	copy(key, []byte(password))

	crypt, err := base64.StdEncoding.DecodeString(crypt64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	iv := crypt[:aes.BlockSize]
	crypt = crypt[aes.BlockSize:]
	decrypted := make([]byte, len(crypt))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, crypt)

	return string(decrypted[:len(decrypted)-int(decrypted[len(decrypted)-1])]), nil
}
