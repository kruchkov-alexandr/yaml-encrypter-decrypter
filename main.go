package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

//global variables
var tmpYamlText []string
var envIndent int = -5555
var currentIndent int = -7777

const AES = "AES256:"

func main() {

	flagKey := flag.String(
		"key",
		"",
		"AES key-password for encrypt/decrypt",
	)
	flagDryRun := flag.Bool(
		"dry-run",
		false,
		"dry-run to STDOUT",
	)
	flagFile := flag.String(
		"filename",
		"",
		"filename for encode/decode",
	)
	flagEnv := flag.String("env", "secret:", "YAML block-name for encode/decode")
	flagValue := flag.String(
		"value",
		"",
		"value for encrypt/decrypt",
	)
	flagOperation := flag.String(
		"operation",
		"",
		"Available operations: encrypt, decrypt",
	)

	flag.Parse()

	// disable timestamp in stdout
	log.SetFlags(0)

	if *flagKey == "" {
		log.Println("Please specify environment variable \"YED-PASSWORD\" ")
		os.Exit(0)
	}

	// decrypt/encrypt value
	if *flagValue != "" {
		if matchContains(*flagValue, AES) {
			log.Println(decryptOneValue(*flagKey, strings.TrimPrefix(*flagValue, AES)))
			os.Exit(0)
		} else {
			log.Println(AES + encryptOneValue(*flagKey, *flagValue))
			os.Exit(0)
		}
	}

	if *flagOperation == "" {
		log.Println("Please, specify operation: encrypt or decrypt")
		os.Exit(0)
	}

	// read file
	text := readFile(*flagFile)
	for _, eachLn := range text {

		// disable double-encode issue
		if matchContains(eachLn, AES) && *flagOperation == "encrypt" {
			log.Printf("Cannot encode file %v!\n"+
				"It seems that string \"%v\" already encoded!\n", *flagFile, eachLn)
			os.Exit(1)
		}

		// current indent
		currentIndent = countIndent(eachLn)
		//log.Println(currentIndent)

		// check if current line is env block
		if matchPrefixEnvBlock(strings.TrimSpace(eachLn), *flagEnv) {
			envIndent = currentIndent
			if *flagDryRun {
				log.Println(eachLn)
			} else {
				tmpYamlText = append(tmpYamlText, eachLn)
			}
			continue
		}

		// main logic
		if len(eachLn) != 0 || matchPrefixCharacter(strings.TrimSpace(eachLn), "#") {
			if currentIndent == envIndent+2 {
				parsedString := parseEachLine(eachLn, *flagKey, *flagOperation)
				if *flagDryRun {
					log.Println(parsedString)
				} else {
					tmpYamlText = append(tmpYamlText, parsedString)
				}
			} else {
				// flush envIndent
				envIndent = -5645
				if *flagDryRun {
					log.Println(eachLn)
				} else {
					tmpYamlText = append(tmpYamlText, eachLn)
				}
			}
		} else {
			if *flagDryRun {
				log.Println(eachLn)
			} else {
				tmpYamlText = append(tmpYamlText, eachLn)
			}
		}
	}

	//if already ok, read temp yaml slice and rewrite target yaml file
	if !*flagDryRun {
		file, err := os.OpenFile(*flagFile, os.O_TRUNC|os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
}

func parseEachLine(eachLn string, key string, operation string) string {
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
	if len(strings.TrimSpace(eachLn)) == 0 {
		parsedLine = eachLn
		return parsedLine
	}
	// skip if line is empty with indent
	if len(eachLn) == 0 {
		parsedLine = eachLn
		return parsedLine
	}
	// skip if line is comment, started with #
	if matchPrefixCharacter(strings.TrimSpace(eachLn), "#") {
		parsedLine = eachLn
		return parsedLine
	}

	// skip if line is comment, started with ---
	if matchPrefixCharacter(strings.TrimSpace(eachLn), "---") {
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

	// convert if Value is not empty, but contains quotes
	if len(stringArray) >= 2 && stringArray[1] != "\"\"" && matchContains(strings.TrimSpace(eachLn), "\"") {
		regexTemplate := regexp.MustCompile(`"[^"]+"`)
		oldValueString := strings.Join(regexTemplate.FindAllString(eachLn, 1), "")

		var result string
		if operation == "encrypt" {
			encryptedValue, err := encryptAES(key, oldValueString)
			if err != nil {
				log.Fatalln("Cannot encrypt value, reason:", err)
			}
			result = encryptedValue
			result = AES + result
		} else if operation == "decrypt" {
			decryptedValue, err := decryptAES(key, strings.TrimPrefix(stringArray[1], AES))
			if err != nil {
				log.Fatalln("Cannot decrypt key/value. \n"+
					"Please check encoded variables in YAML file, reason", err)
			}
			result = decryptedValue
		} else {
			log.Println("Please, specify operation: encrypt or decrypt")
			os.Exit(0)
		}
		stringReplaced := strings.ReplaceAll(eachLn, oldValueString, result)
		parsedLine = stringReplaced
		return parsedLine
	}
	// convert if Value is not empty, but NOT contains quotes
	if len(stringArray) >= 2 && stringArray[1] != "\"\"" && !matchContains(strings.TrimSpace(eachLn), "\"") {

		var result string

		if operation == "encrypt" {
			encryptedValue, err := encryptAES(key, stringArray[1])
			if err != nil {
				log.Fatalln("Something wrong, cannot encrypt file, 3", err)
			}

			result = encryptedValue
			stringArray[1] = AES + result
		} else if operation == "decrypt" {
			decryptedValue, err := decryptAES(key, strings.TrimPrefix(stringArray[1], AES))
			if err != nil {
				log.Fatalln("Cannot decrypt key/value. \n"+
					"Please check encoded variables in YAML file, reason", err)
			}
			result = decryptedValue
			stringArray[1] = result
		} else {
			log.Println("Please, specify operation: encrypt or decrypt")
			os.Exit(0)
		}

		parsedLine = whitespaces + strings.Join(stringArray[:], " ")
		return parsedLine
	}
	return parsedLine
}

//for @jaxel87, encrypt value by flag without encryption file
func encryptOneValue(key string, value string) string {
	encrypted, err := encryptAES(key, value)
	if err != nil {
		log.Fatalf("Something wrong during encrypt value")
	}
	return encrypted
}

//for @jaxel87, decrypt value by flag without decryption file
func decryptOneValue(key string, value string) string {
	decrypted, err := decryptAES(key, value)
	if err != nil {
		log.Fatalf("Something wrong during decrypt value")
	}
	return decrypted
}

// calculate indents for line
func countIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// match if line is ENV block
func matchPrefixEnvBlock(line string, env string) bool {
	if strings.HasPrefix(line, env) {
		return true
	} else {
		return false
	}
}

// match if line with character
func matchPrefixCharacter(line string, character string) bool {
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
