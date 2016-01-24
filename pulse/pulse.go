package pulse

import (
	"fmt"
	"math"
	"sort"
	"time"
	"unicode"
)

type outputFunc func(string)

type unmatchedLog struct {
	line       string
	dateStored time.Time
	reported   bool
}

type revision struct {
	tokenPtr   *token
	tokenIndex int
	variations *[]variation
	text       string
}

type variation struct {
	text       string
	numMatches int64
}

type token struct {
	word       string
	variable   bool
	required   bool
	variations []variation
}

type pattern struct {
	tokens     []token
	numMatches int64
}

type vertex struct {
	x                      int
	y                      int
	startsSequenceOfLength int
}

type vertexDistance struct {
	distance int
	index    int
}

type distArray []vertexDistance

//Channel to receive log data from consuming application
var input <-chan string
var report outputFunc
var patternCreationRate float64
var patternCreationRateIncreasing bool
var inputsSinceLastNewPattern int64
var lastPatternCount int
var unmatched []unmatchedLog
var patterns []pattern

const tokenMapSize int = 2048

var tokenMap [tokenMapSize]map[*pattern]bool

func (s distArray) Len() int           { return len(s) }
func (s distArray) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s distArray) Less(i, j int) bool { return s[i].distance < s[j].distance }

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

//sets initial state of the token map, used to lookup existing patterns
func initTokenMap() {
	for i := 0; i < tokenMapSize; i++ {
		tokenMap[i] = make(map[*pattern]bool)
	}
}

//converts a string into a slice of strings.  symbols and contiguous strings of any other type
//are returned as individual elements.  all whitespace is excluded
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

//adds a vertex to a list of vertices, or updates a variable if the vertex already exists in the list
func addUpdateVertex(newValue vertex, list []vertex) []vertex {
	var done = false
	for i := range list {
		if newValue.x == list[i].x && newValue.y == list[i].y {
			list[i].startsSequenceOfLength = newValue.startsSequenceOfLength
			done = true
			break
		}
	}

	if !done {
		list = append(list, newValue)
	}

	return list
}

//with a suppliced list of verticies and a particular vertex, this algorithm
//locates the vertex in the list that is closest to the supplied vertex,
//however some preferential treatment is given to vertices that begin a
//longer sequence of shared substrings in the inputs being compared
func getNextVertex(value vertex, vertices []vertex) (bool, vertex) {
	x := value.x
	y := value.y

	var distances []vertexDistance
	nextVertexExists := false

	for i := range vertices {
		v := vertices[i]
		if v.x < x || v.y < y {
			continue
		}

		nextVertexExists = true
		distances = append(distances, vertexDistance{(v.x - x) + (v.y - y), i})
	}

	if !nextVertexExists {
		return false, vertex{0, 0, 0}
	}

	sort.Sort(distArray(distances))

	var minDistance = distances[0]
	var nextVertex = vertices[minDistance.index]
	var nextMin = vertexDistance{0, 0}
	if len(distances) > 1 {
		nextMin = distances[1]
		var difference = nextMin.distance - minDistance.distance
		if difference <= 3 && vertices[nextMin.index].startsSequenceOfLength > nextVertex.startsSequenceOfLength {
			nextVertex = vertices[nextMin.index]
		}
	}

	return true, nextVertex
}

//removes the supplied vertex from the list of vertices, and returns the updated list
func removeVertexFromList(val vertex, vertices []vertex) []vertex {
	for i := range vertices {
		if vertices[i].x == val.x && vertices[i].y == val.y {
			vertices = append(vertices[:i], vertices[i+1:]...)
			break
		}
	}
	return vertices
}

//returns sorted list of tokens in pattern, sorted in the order they appear in both strings
func analyzeMatrix(matrix [][]int, vertices []vertex) (bool, []vertex) {
	//start with {0, 0}
	var tokens []vertex
	if matrix[0][0] > 0 {
		tokens = append(tokens, vertices[0])
		vertices = removeVertexFromList(vertices[0], vertices)
	}
	var start = vertex{0, 0, 0}
	var foundNextPoint, nextPoint = getNextVertex(start, vertices)
	for foundNextPoint {
		tokens = append(tokens, nextPoint)
		vertices = removeVertexFromList(nextPoint, vertices)
		foundNextPoint, nextPoint = getNextVertex(nextPoint, vertices)
	}
	return float64(len(tokens)) > float64(len(matrix[0])/2), tokens
}

//using a list of words and a pointer to a pattern, update the
//token map so that the words can be used to later locate the pattern
func updateTokenMap(words []token, ref *pattern) {
	for i := range words {
		if words[i].variable {
			continue
		}

		chars := []rune(words[i].word)
		sum := 0
		for j := range chars {
			value := int(chars[j])
			sum += value
		}

		sum = sum % tokenMapSize

		var pm = tokenMap[sum]
		pm[ref] = true
	}
}

//returns all patterns that a particular word is part of, using the token map
func patternsFromToken(word string) []*pattern {
	chars := []rune(word)
	sum := 0
	for i := range chars {
		value := int(chars[i])
		sum += value
	}

	sum = sum % tokenMapSize
	var pm = tokenMap[sum]
	keys := make([]*pattern, 0, len(pm))
	for k := range pm {
		if pm[k] == true {
			keys = append(keys, k)
		}
	}

	return keys
}

//match a pattern against a new input, revising the pattern under certain circumstances
func matchPattern(pat pattern, longTokens []string, input string) bool {
	foundPattern := false
	var vertices []vertex
	var shortTokens []string
	for i := range pat.tokens {
		shortTokens = append(shortTokens, pat.tokens[i].word)
	}

	matrix := make([][]int, len(shortTokens))
	for i := range shortTokens {
		matrix[i] = make([]int, len(longTokens))
		for j := range matrix[i] {
			var matches = 0
			if shortTokens[i] == longTokens[j] {
				matches++
				vertices = addUpdateVertex(vertex{i, j, matches}, vertices)
				var prevRow = j - 1
				var prevCol = i - 1
				for prevRow > 0 && prevCol > 0 {
					if shortTokens[prevCol] == longTokens[prevRow] {
						matches++
						vertices = addUpdateVertex(vertex{prevCol, prevRow, matches}, vertices)
						prevRow--
						prevCol--
					} else {
						break
					}
				}
			}
			matrix[i][j] = matches
		}
	}

	foundPattern, vertices = analyzeMatrix(matrix, vertices)
	var newPattern pattern
	if foundPattern {
		lastPoint := vertex{-1, -1, 0}
		for i := range vertices {
			var skippedBeginning = i == 0 && vertices[i].x != 0 && vertices[i].y != 0
			var vertex = vertices[i]
			var distance = (vertex.x - lastPoint.x) + (vertex.y - lastPoint.y)
			if distance <= 2 && !skippedBeginning {
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				newPattern.tokens = append(newPattern.tokens, token{text, false, true, nil})
			} else {
				xDiff := vertex.x - lastPoint.x
				yDiff := vertex.y - lastPoint.y
				skippedColText := ""
				skippedRowText := ""
				if xDiff > 1 {
					var skipped = shortTokens[lastPoint.x+1 : vertex.x]
					for x := range skipped {
						skippedColText += skipped[x]
					}
					if skippedColText == "!WILDCARD!" {
						skippedColText = ""
					}
				}

				if yDiff > 1 {
					var skipped = longTokens[lastPoint.y+1 : vertex.y]
					for y := range skipped {
						skippedRowText += skipped[y]
					}
				}

				var variableText []variation
				if skippedColText != "" {
					variableText = append(variableText, variation{skippedColText, 1})
				}
				if skippedRowText != "" {
					variableText = append(variableText, variation{skippedRowText, 1})
				}
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				//add wildcard token to sequence
				newPattern.tokens = append(newPattern.tokens, token{"!WILDCARD!", true, len(variableText) > 1, variableText})
				//add static token to sequence
				newPattern.tokens = append(newPattern.tokens, token{text, false, true, nil})
			}
		}

		if len(newPattern.tokens) <= len(pat.tokens) {
			for i := range newPattern.tokens {
				var originalToken = pat.tokens[i]
				var newToken = newPattern.tokens[i]
				var newText string
				if newToken.variable && len(newToken.variations) == 1 {
					newText = newToken.variations[0].text
				}

				if originalToken.variable && newToken.variable {
					accountedForNewValue := false
					for j := range originalToken.variations {
						if originalToken.variations[j].text == newText {
							originalToken.variations[j].numMatches++
							accountedForNewValue = true
						}
					}
					if !accountedForNewValue {
						originalToken.variations = append(originalToken.variations, variation{newText, 1})
					}
				} else if newToken.variable && !originalToken.variable {
					originalToken.word = "!WILDCARD!"
					originalToken.variable = true
					for j := range newToken.variations {
						originalToken.variations = append(originalToken.variations, variation{newToken.variations[j].text, 1})
					}
				}

				pat.tokens[i] = originalToken
				pat.numMatches++
			}
		} else {
			//determine how close the patterns are
			var diff = math.Abs(float64(len(pat.tokens)) - float64(len(newPattern.tokens)))
			var maxLength = float64(max(len(pat.tokens), len(newPattern.tokens)))
			if ((maxLength - diff) / maxLength) >= 0.90 {
				return true
			}

			//a match was made above a certain threshold between the pattern and the input, but the length of tokens is too far off
			reportAnomaly(input)
			return false
		}

		return true
	}
	return false
}

//looks for a pattern between two input strings, and learns the new pattern if
//a certain threshold value is reached when the matrix is analyzed.
func findPattern(shortTokens []string, longTokens []string) bool {
	foundPattern := false
	var vertices []vertex
	matrix := make([][]int, len(shortTokens))
	for i := range shortTokens {
		matrix[i] = make([]int, len(longTokens))
		for j := range matrix[i] {
			var matches = 0
			if shortTokens[i] == longTokens[j] {
				matches++
				vertices = addUpdateVertex(vertex{i, j, matches}, vertices)
				var prevRow = j - 1
				var prevCol = i - 1
				for prevRow > 0 && prevCol > 0 {
					if shortTokens[prevCol] == longTokens[prevRow] {
						matches++
						vertices = addUpdateVertex(vertex{prevCol, prevRow, matches}, vertices)
						prevRow--
						prevCol--
					} else {
						break
					}
				}
			}
			matrix[i][j] = matches
		}
	}

	foundPattern, vertices = analyzeMatrix(matrix, vertices)
	if foundPattern {
		var p pattern

		lastPoint := vertex{-1, -1, 0}
		for i := range vertices {
			var skippedBeginning = i == 0 && vertices[i].x != 0 && vertices[i].y != 0
			var vertex = vertices[i]
			var distance = (vertex.x - lastPoint.x) + (vertex.y - lastPoint.y)
			if distance <= 2 && !skippedBeginning {
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				p.tokens = append(p.tokens, token{text, false, true, nil})
			} else {
				xDiff := vertex.x - lastPoint.x
				yDiff := vertex.y - lastPoint.y

				skippedColText := ""
				skippedRowText := ""
				if xDiff > 1 {
					var skipped = shortTokens[lastPoint.x+1 : vertex.x]
					for x := range skipped {
						skippedColText += skipped[x]
					}
				}

				if yDiff > 1 {
					var skipped = longTokens[lastPoint.y+1 : vertex.y]
					for y := range skipped {
						skippedRowText += skipped[y]
					}
				}

				var variableText []variation
				if skippedColText != "" {
					variableText = append(variableText, variation{skippedColText, 1})
				}
				if skippedRowText != "" {
					variableText = append(variableText, variation{skippedRowText, 1})
				}
				lastPoint = vertex
				text := shortTokens[lastPoint.x]
				//add wildcard token to sequence
				p.tokens = append(p.tokens, token{"!WILDCARD!", true, len(variableText) > 1, variableText})
				//add static token to sequence
				p.tokens = append(p.tokens, token{text, false, true, nil})
			}
		}

		p.numMatches = 1
		patterns = append(patterns, p)

		var reference = &p
		updateTokenMap(p.tokens, reference)

		var numPatterns = len(patterns)
		var rate = 1.0 / float64(inputsSinceLastNewPattern)
		var newAvgRate = ((float64(numPatterns) * patternCreationRate) + rate) / float64(numPatterns+1)
		patternCreationRateIncreasing = newAvgRate > patternCreationRate
		patternCreationRate = newAvgRate

		inputsSinceLastNewPattern = 0
		lastPatternCount = numPatterns
	}
	return foundPattern
}

//simple index helper function to find a string in a slice
func indexOfWord(value string, words []string) int {
	for i := range words {
		if words[i] == value {
			return i
		}
	}
	return -1
}

func indexOfWordInVariations(value string, words []variation) int {
	for i := range words {
		if words[i].text == value {
			return i
		}
	}
	return -1
}

func reportAnomaly(line string) {
	fmt.Printf("\nPattern count: %v\n", len(patterns))

	//fmt.Printf("Pattern creation rate: %v, rate increasing? %v", patternCreationRate, patternCreationRateIncreasing)
	if (!patternCreationRateIncreasing || patternCreationRate <= 0.20) && (len(patterns) != 0) {
		fmt.Printf("\nReporting anomaly...%v\n", line)
		report(line)
	}
}

func analyze(line string) {
	index := -1
	maxScore := 0.0
	patternFound := false
	inputsSinceLastNewPattern++

	if len(patterns) == lastPatternCount {
		patternCreationRate = patternCreationRate * 0.99
	}

	//search for existing pattern using token map
	var tokenMatches = make(map[*pattern]int)
	var lineTokens = getTokens(line)
	for i := range lineTokens {
		var patterns = patternsFromToken(lineTokens[i])
		for j := range patterns {
			var p = patterns[j]
			tokenMatches[p] = tokenMatches[p] + 1
		}
	}

	var mostLikelyPattern *pattern
	var tokensInCommon int

	for k := range tokenMatches {
		if tokenMatches[k] > tokensInCommon {
			tokensInCommon = tokenMatches[k]
			mostLikelyPattern = k
		}
	}

	//fmt.Printf("Tokens in common: %v Tokens in line: %v", tokensInCommon, lineTokens)
	if float64(tokensInCommon)/float64(len(lineTokens)) >= 0.5 {
		//patternFound = matchInputToPattern(mostLikelyPattern, lineTokens, line)
		patternFound = matchPattern(*mostLikelyPattern, lineTokens, line)
	}

	//if no pattern found, compare to unmatched lines, see if a new pattern can be detected
	if !patternFound {
		//fmt.Println("Beginning levenstein distance comparison...")
		for i := range unmatched {
			var compare = unmatched[i].line
			var distance = ld(line, compare)
			var timeUnmatched = time.Since(unmatched[i].dateStored).Seconds()
			if timeUnmatched > 30.0 && !unmatched[i].reported {
				reportAnomaly(unmatched[i].line)
				unmatched[i].reported = true
			}
			var maxLength = max(len(line), len(compare))
			var score = float64(maxLength-distance) / float64(maxLength)
			if score > maxScore {
				maxScore = score
				index = i
			}
		}

		if maxScore >= 0.5 {
			var unmatchedTokens = getTokens(unmatched[index].line)
			if len(lineTokens) < len(unmatchedTokens) {
				patternFound = findPattern(lineTokens, unmatchedTokens)
			} else {
				patternFound = findPattern(unmatchedTokens, lineTokens)
			}
		}

		if !patternFound {
			unmatched = append(unmatched, unmatchedLog{line, time.Now(), true})
			reportAnomaly(line)
		} else { //remove unmatched line from unmatched slice
			unmatched = append(unmatched[:index], unmatched[index+1:]...)
		}
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
	initTokenMap()
	go func() {
		for value := range in {
			analyze(value)
		}
	}()
}
