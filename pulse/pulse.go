package pulse

import (
	"fmt"
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

func initTokenMap() {
	//fmt.Println("Initializing token map..")
	for i := 0; i < tokenMapSize; i++ {
		tokenMap[i] = make(map[*pattern]bool)
	}
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

func removeVertexFromList(val vertex, vertices []vertex) []vertex {
	for i := range vertices {
		if vertices[i].x == val.x && vertices[i].y == val.y {
			vertices = append(vertices[:i], vertices[i+1:]...)
			break
		}
	}
	return vertices
}

//returns sorted list of tokens in pattern
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

	//fmt.Println("Token map updated with pattern...")
}

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

	/*for j := range longTokens {
		fmt.Printf("\n")
		for i := range shortTokens {
			fmt.Printf("%v ", matrix[i][j])
		}
	}*/

	/*for i := range vertices {
		var vertex = vertices[i]
		fmt.Printf("%v \n", vertex)
	}*/

	foundPattern, vertices = analyzeMatrix(matrix, vertices)
	if foundPattern {
		//fmt.Println("Found a pattern...")
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
					//fmt.Println("Skipped col: " + skippedColText)
				}

				if yDiff > 1 {
					var skipped = longTokens[lastPoint.y+1 : vertex.y]
					for y := range skipped {
						skippedRowText += skipped[y]
					}
					//fmt.Println("Skipped row: " + skippedRowText)
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
			//fmt.Printf("%v \n", vertex)
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

		//fmt.Printf("Pattern creation rate: %v Increasing? %v\n", patternCreationRate, patternCreationRateIncreasing)

		inputsSinceLastNewPattern = 0
		lastPatternCount = numPatterns

		//fmt.Printf("Pattern: %v \n", p)
	}
	return foundPattern
}

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

func matchInputToPattern(p *pattern, words []string, input string) bool {
	var patternTokens = p.tokens
	var lastVariableToken *token
	var matchingTokens = 0
	var requiredTokenCount = 0
	var requiredStaticTokenCount = 0
	var staticTokensMatched = 0
	var reviseVariable = false
	var revisionIndex = -1
	var revisions []revision
	for i := range patternTokens {
		var t = patternTokens[i]
		if t.required {
			requiredTokenCount++
		}

		if t.variable {
			lastVariableToken = &t
			reviseVariable = true
			revisionIndex = i
			continue
		} else {
			requiredStaticTokenCount++
			var index = indexOfWord(t.word, words)
			if index == -1 { //required token was not found in input
				return false
			} else if index == 0 {
				matchingTokens++
				staticTokensMatched++
				words = words[index+1:]
			} else if index > 0 && reviseVariable {
				text := words[:index]
				textStr := ""
				//build string to add or update
				for j := range text {
					textStr = textStr + text[j]
				}

				var lvt = *lastVariableToken
				revisions = append(revisions, revision{&lvt, revisionIndex, &lvt.variations, textStr})

				matchingTokens = matchingTokens + 2 //the variable token and the static one
				staticTokensMatched++

				words = words[index+1:]
			}
		}
	}
	//fmt.Printf("\nRequired tokens in pattern: %v  Tokens matched: %v\n", requiredStaticTokenCount, staticTokensMatched)
	//fmt.Printf("Revisions: %v\n", revisions)

	if requiredTokenCount == matchingTokens {
		for i := range revisions {
			var v = revisions[i].variations
			var str = revisions[i].text
			var token = *(revisions[i].tokenPtr)
			var idx = indexOfWordInVariations(str, *v)
			if idx == -1 {
				*v = append(*v, variation{str, 1})
				token.variations = *v
				p.tokens[revisions[i].tokenIndex] = token
			} else {
				arr := *v
				arr[idx].numMatches++
			}
		}
		p.numMatches++
		//fmt.Printf("Pattern revised: %v", p)
		return true
	}

	if staticTokensMatched >= requiredStaticTokenCount { //all the required static elements matched, but not revising
		p.numMatches++
		reportAnomaly(input)
		return true
	}

	return false
}

func reportAnomaly(line string) {
	//fmt.Printf("Pattern count: %v\n", len(patterns))

	//fmt.Printf("Pattern creation rate: %v, rate increasing? %v", patternCreationRate, patternCreationRateIncreasing)
	if (!patternCreationRateIncreasing || patternCreationRate <= 0.20) && (len(patterns) != 0) {
		fmt.Printf("Reporting anomaly...%v", line)
		report(line)
	}
}

func analyze(line string) {
	index := -1
	maxScore := 0.0
	patternFound := false
	inputsSinceLastNewPattern++

	if len(patterns) == lastPatternCount {
		patternCreationRate = patternCreationRate * 0.95
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

	if float64(tokensInCommon)/float64(len(lineTokens)) >= 0.5 {
		patternFound = matchInputToPattern(mostLikelyPattern, lineTokens, line)
	}

	//if no pattern found, compare to unmatched lines, see if a new pattern can be detected
	if !patternFound {
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
			unmatched = append(unmatched, unmatchedLog{line, time.Now(), false})
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
	/*analyze("monkey x [michaeld] Hello World")
	analyze("monkey x y x [bob] Hello World")
	analyze("Harry is a bad donkey.")
	analyze("Martin is a bad donkey.")
	analyze("Martin is a bad donkey.")
	analyze("monkey x [harrison] Hello World!")
	analyze("John is a bad donkey.")
	analyze("The black bird flies at midnight.")
	analyze("A carriage rides along the lengthy dirt road.")
	analyze("monkey x [michaeld] Hello World")
	analyze("monkey x y x [bob] Hello World")*/
	go func() {
		for value := range in {
			analyze(value)
		}
	}()
}
