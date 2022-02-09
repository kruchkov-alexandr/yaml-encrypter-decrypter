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
var envWhiteSpaces int
var valuesForFormatting bool = false
var tmpYamlText []string

func main() {
	flagFile := flag.String(
		"filename",
		"values.yaml",
		"filename for encode/decode",
	)
	flagKey := flag.String(
		"key",
		"}tf&Wr+Nt}A9g{s",
		"AES key for encrypt/decrypt",
	)
	flagEnv := flag.String("env", "env:", "block-name for encode/decode")
	flagDebug := flag.String("debug", "false", "debug mode, print encode/decode to stdout")
	flagEncryptValue := flag.String("encrypt", "", "value to encrypt")
	flagDecryptValue := flag.String("decrypt", "", "value to decrypt")
	flag.Parse()

	filename := *flagFile
	key := *flagKey
	env := *flagEnv
	debug := *flagDebug
	encryptValue := *flagEncryptValue
	decryptValue := *flagDecryptValue
	// for @kpogonea
	const AES = "AES256:"

	// for @jaxel87, encrypt/decrypt value by flag
	if encryptValue != "" {
		encrypted, err := encryptAES(key, encryptValue)
		fmt.Println(encrypted)
		if err != nil {
			log.Fatalf("something wrong during encrypt")
		}
		os.Exit(0)
	}
	if decryptValue != "" {
		decrypted, err := decryptAES(key, decryptValue)
		fmt.Println(decrypted)
		if err != nil {
			log.Fatalf("something wrong during decrypt")
		}
		os.Exit(0)
	}

	text := readFile(filename)
	for _, eachLn := range text {
		//show current whitespaces before character
		currentWhiteSpaces := countLeadingSpaces(eachLn)
		if envWhiteSpaces > currentWhiteSpaces {
			valuesForFormatting = false
		}
		if valuesForFormatting {
			stringArray := strings.Fields(eachLn)
			whitespaces := strings.Repeat(" ", currentWhiteSpaces)

			encrypted, err := encryptAES(key, stringArray[1])
			if err != nil {
				log.Fatalf("something wrong")
			}
			matchedAesEncrypted, _ := regexp.MatchString(AES, stringArray[1])
			if !matchedAesEncrypted {
				if debug == "true" {
					fmt.Println(whitespaces + stringArray[0] + " " + AES + encrypted)
				}
				tmpYamlText = append(tmpYamlText, whitespaces+stringArray[0]+" "+AES+encrypted)
			} else {
				aesBeforeDecrypt := strings.ReplaceAll(stringArray[1], AES, "")
				decrypted, err := decryptAES(key, aesBeforeDecrypt)
				if err != nil {
					log.Fatalf("something wrong during decrypt")
				}
				if debug == "true" {
					fmt.Println(whitespaces + stringArray[0] + " " + decrypted)
				}
				tmpYamlText = append(tmpYamlText, whitespaces+stringArray[0]+" "+decrypted)
			}
		} else {
			if debug == "true" {
				fmt.Println(eachLn)
			}
			tmpYamlText = append(tmpYamlText, eachLn)
		}
		matchedEnvVariable, _ := regexp.MatchString(env, eachLn)
		if matchedEnvVariable {
			envWhiteSpaces = currentWhiteSpaces + 2
			valuesForFormatting = true
		}
	}

	// if already ok, read temp yaml slice and rewrite target yaml file
	if debug != "true" {
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
}

func countLeadingSpaces(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
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

//func main() {
//	encoded, err := encryptAES("secretkey", "plaintext")
//	fmt.Println(encoded)
//	if err != nil {
//		log.Fatalf("error")
//	}
//	decoded, err2 := decryptAES("secretkey", "30tEfhuJSVRhpG97XCuWgz2okj7L8vQ1s6V9zVUPeDQ=")
//	fmt.Println(decoded)
//	if err2 != nil {
//		log.Fatalf("error")
//	}
//
//}

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
