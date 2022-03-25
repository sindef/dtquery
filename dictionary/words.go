package dictionary

import (
	"bufio"
	"log"
	"math/rand"
	"os"
)

//Based on length, return a word from a random line in the words.txt file
func Slow(length int) string {
	var matches []string
	var word string
	//Open the file
	file, err := os.Open("dictionary/words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	//Create a new reader for the file
	reader := bufio.NewReader(file)
	//Create a new scanner
	scanner := bufio.NewScanner(reader)

	//Loop through the lines in the file - any words that match the length will be added to the matches slice
	for scanner.Scan() {
		if len(scanner.Text()) == length {
			matches = append(matches, scanner.Text())
		}
	}

	//Given the length of the matches slice, pick a random word from the slice
	if len(matches) > 0 {
		word = matches[rand.Intn(len(matches))]
	} else if len(matches) == 0 {
		log.Fatal("No words of length ", length, " found in words.txt")
	}

	return word
}

//Based on length, return a word from the first matching line in the words.txt file
func Fast(length int) string {
	var word string
	//Open the file
	file, err := os.Open("dictionary/words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	//Create a new reader for the file
	reader := bufio.NewReader(file)
	//Create a new scanner
	scanner := bufio.NewScanner(reader)

	//Loop through the lines in the file - any words that match the length will be added to the matches slice
	for scanner.Scan() {
		if len(scanner.Text()) == length {
			word = scanner.Text()
			//Break out of the loop
			break
		}
	}
	return word
}

//Based on length, return a slice of every possible word of that length
func All(length int) []string {
	var matches []string
	//Open the file
	file, err := os.Open("dictionary/words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	//Create a new reader for the file
	reader := bufio.NewReader(file)
	//Create a new scanner
	scanner := bufio.NewScanner(reader)

	//Loop through the lines in the file - any words that match the length will be added to the matches slice
	for scanner.Scan() {
		if len(scanner.Text()) == length {
			matches = append(matches, scanner.Text())
		}
	}

	return matches
}
