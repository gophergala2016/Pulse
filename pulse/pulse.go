package pulse

import "unicode"

type outputFunc func(string)

type variation struct {
	text      string
	frequency float64
}

type token struct {
	word       string
	variable   bool
	required   bool
	variations []variation
}

type pattern struct {
	tokens    []token
	frequency float64
}

//Channel to receive log data from consuming application
var input <-chan string
var report outputFunc
var patternCreationRate float64
var unmatched []string
var tokenMap [2048][]string

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func getTokens(value string) []string {
	var buffer []rune
	var result []string
	chars := []rune(value)
	for i, r := range chars {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			if len(buffer) > 0 {
				result = append(result, string(buffer))
				buffer = nil
			}
			result = append(result, string(r))
		} else if unicode.IsSpace(r) {
			if len(buffer) > 0 {
				result = append(result, string(buffer))
			}
			buffer = nil
		} else {
			buffer = append(buffer, r)
			if i == len(chars)-1 {
				result = append(result, string(buffer))
			}
		}
	}
	return result
}

func findPattern(shortTokens []string, longTokens []string) bool {
	foundPattern := false
	matrix := make([][]int, len(shortTokens))
	for i := range shortTokens {
		matrix[i] = make([]int, len(longTokens))
	}

	return foundPattern
}

func analyze(line string) {
	index := -1
	maxScore := 0.0
	patternFound := false

	//search for existing pattern using token map

	//if no pattern found, compare to unmatched lines, see if a new pattern can be detected
	for i := range unmatched {
		var compare = unmatched[i]
		var distance = ld(line, compare)
		var maxLength = max(len(line), len(compare))
		var score = float64(maxLength-distance) / float64(maxLength)
		if score > maxScore {
			maxScore = score
			index = i
		}
	}

	if maxScore >= 0.5 {
		report("Looking for pattern...")
		var lineTokens = getTokens(line)
		var unmatchedTokens = getTokens(unmatched[index])
		if len(lineTokens) < len(unmatchedTokens) {
			patternFound = findPattern(lineTokens, unmatchedTokens)
		} else {
			patternFound = findPattern(unmatchedTokens, lineTokens)
		}
	}

	if !patternFound {
		unmatched = append(unmatched, line)
		report("Added line to unmatched")
	}
}

//Copied from http://rosettacode.org/wiki/Levenshtein_distance#Go
func ld(s, t string) int {
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}

	}
	return d[len(s)][len(t)]
}

//Run starts the pulse package
func Run(in <-chan string, out outputFunc) {
	input = in
	report = out
	go func() {
		for value := range in {
			analyze(value)
			//report(value)
		}
	}()
}
