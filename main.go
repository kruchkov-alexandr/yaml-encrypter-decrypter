package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
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
		"8d9b2dd4c94e8ac7ef742fc0ed162adf49ef8676f906517de1d5085a817ec824",
		"key for encode, only length 32bit",
	)
	flagEnv := flag.String("env", "env:", "block-name for encode/decode")
	flagDebug := flag.String("debug", "false", "debug mode, print encode/decode to stdout")
	flag.Parse()

	filename := *flagFile
	key := *flagKey
	env := *flagEnv
	debug := *flagDebug
	const AES = "AES256-encoded:"

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

			encrypted := encrypt(stringArray[1], key)
			matchedAesEncrypted, _ := regexp.MatchString(AES, stringArray[1])
			if !matchedAesEncrypted {
				if debug == "true" {
					fmt.Println(whitespaces + stringArray[0] + " " + AES + encrypted)
				}
				tmpYamlText = append(tmpYamlText, whitespaces+stringArray[0]+" "+AES+encrypted)
			} else {
				aesBeforeDecrypt := strings.ReplaceAll(stringArray[1], AES, "")
				decrypted := decrypt(aesBeforeDecrypt, key)
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

func encrypt(stringToEncrypt string, keyString string) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Please enter 32 length key")

	}
	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data.
	//The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func decrypt(encryptedString string, keyString string) (decryptedString string) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Fatalf("Incorrect key, please enter correct key, 32 length")
	}

	return fmt.Sprintf("%s", plaintext)
}
