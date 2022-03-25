package dns

import (
	"bytes"
	"dtquery/dictionary"
	"math/rand"
)

//Function to generate random domain names
func Random() string {
	var buffer bytes.Buffer
	//Generate a random number between 1 and 14
	length := uint8(rand.Intn(14) + 1)
	//Generate a random string of length length
	for i := uint8(0); i < length; i++ {
		buffer.WriteByte(byte(rand.Intn(26) + 97))
		//Add a period before the penultimate character
		if i == length-3 {
			buffer.WriteByte(46)
		}
	}
	return buffer.String()
}

//Function to generate the first .com domain of specified length
func QuickWord(length int, TLD string) string {
	word := dictionary.Fast(length)
	return word + TLD
}

//Function to generate a random .com domain of specified length
func RandomWords(length int, TLD string) string {
	word := dictionary.Slow(length)
	return word + TLD
}

//Function to return a slice of every possible .com domain of specified length
func AllWords(length int, TLD string) []string {
	words := dictionary.All(length)
	for i := 0; i < len(words); i++ {
		words[i] = words[i] + TLD
	}
	return words
}
