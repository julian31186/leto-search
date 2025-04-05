package search

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"regexp"
	"sync"
	"sort"
)

var wg sync.WaitGroup

type Phrase struct {
	Word string
	Freq int
	Doc  string
	Wiki string
	Desc string
}

type TermBody struct {
	Desc string `json:"description"`
	Wiki string `json:"wiki"`
}

type Result struct {
	Term string
	Wiki string
}

func (r Result) String() string {
	return fmt.Sprintf("Term %v, Wiki: %v", r.Term, r.Wiki)
}

type KeyValue struct {
	Key string
	Val int
}

func Search(searchSentence string, index map[string][]Phrase, resultLimit int) ([]Result,error) {
	
	var results []Result

	docToBody := make(map[string]TermBody)
	termCounter := make(map[string]int)

	var termFreqList []KeyValue

	tokens := strings.Fields(searchSentence)
	
	for _,token := range tokens {
		
		token := strings.ToLower(token)

		val,ok := index[token]
		if !ok {
			continue
		}

		for _,p := range val {

			docToBody[p.Doc] = TermBody{
				p.Desc,
				p.Wiki,
			}

			termCounter[p.Doc] += p.Freq
		}
	}

	for k,v := range termCounter {
		termFreqList = append(termFreqList, KeyValue{
			Key: k,
			Val: v,
		})
	}

	// Sorts in descending order
	sort.SliceStable(termFreqList, func(i,j int) bool {
		if termFreqList[i].Val != termFreqList[j].Val {
			return termFreqList[i].Val > termFreqList[j].Val
		}
		return termFreqList[i].Key < termFreqList[j].Key
	})

	// Extracts resultLimit results and returns []Result
	for i := 0; i < len(termFreqList); i += 1{
		if len(results) == resultLimit {
			break
		} else {
			results = append(results, Result{
				Term: termFreqList[i].Key,
				Wiki: docToBody[termFreqList[i].Key].Wiki,
			})
		}
	}

	return results,nil
}

func ReadFile(filePath string) (map[string]TermBody,error) {

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil,err
	}

	defer jsonFile.Close()
	
	fileInfo,_ := jsonFile.Stat()
	fileSize := fileInfo.Size()

	buffer := make([]byte, fileSize)

	bytesRead,err := jsonFile.Read(buffer)
	if err != nil {
		return nil,err
	}
	if bytesRead < int(fileSize) {
		return nil,fmt.Errorf("could not read all bytes of json file")
	}

	var data map[string]TermBody

	err = json.Unmarshal(buffer, &data)
	if err != nil {
		return nil,err
	}

	return data,nil
}

func BuildIndex() (map[string][]Phrase,error) {

	stopWords := map[string]bool{
		"a":          true,
		"an":         true,
		"and":        true,
		"are":        true,
		"as":         true,
		"at":         true,
		"be":         true,
		"but":        true,
		"by":         true,
		"for":        true,
		"from":       true,
		"in":         true,
		"is":         true,
		"it":         true,
		"of":         true,
		"on":         true,
		"or":         true,
		"that":       true,
		"the":        true,
		"this":       true,
		"to":         true,
		"was":        true,
		"were":       true,
		"will":       true,
		"with":       true,
		"i":          true,
		"you":        true,
		"he":         true,
		"she":        true,
		"we":         true,
		"they":       true,
		"me":         true,
		"him":        true,
		"her":        true,
		"us":         true,
		"them":       true,
		"my":         true,
		"your":       true,
		"his":        true,
		"its":        true,
		"our":        true,
		"their":      true,
		"am":         true,
		"been":       true,
		"being":      true,
		"have":       true,
		"has":        true,
		"had":        true,
		"do":         true,
		"does":       true,
		"did":        true,
		"shall":      true,
		"should":     true,
		"would":      true,
		"could":      true,
		"about":      true,
		"against":    true,
		"between":    true,
		"into":       true,
		"through":    true,
		"during":     true,
		"before":     true,
		"after":      true,
		"above":      true,
		"below":      true,
		"up":         true,
		"down":       true,
		"out":        true,
		"off":        true,
		"over":       true,
		"under":      true,
		"again":      true,
		"further":    true,
		"then":       true,
		"once":       true,
		"here":       true,
		"there":      true,
		"when":       true,
		"where":      true,
		"why":        true,
		"how":        true,
		"all":        true,
		"any":        true,
		"both":       true,
		"each":       true,
		"few":        true,
		"more":       true,
		"most":       true,
		"other":      true,
		"some":       true,
		"such":       true,
		"no":         true,
		"nor":        true,
		"not":        true,
		"only":       true,
		"own":        true,
		"same":       true,
		"so":         true,
		"than":       true,
		"too":        true,
		"very":       true,
		"can":        true,
		"just":       true,
		"don":        true,
		"now":        true,
	}

	phrases := make(chan Phrase)

	index := make(map[string][]Phrase)

	data,err := ReadFile("data.json")
	if err != nil {
		return nil,err
	}

	// Tokenize
	for key,value := range data {
		wg.Add(1)
		go tokenize(key,value,stopWords,phrases)
	}

	go func() {
		wg.Wait()
		close(phrases)
	}()

	// Merges Phrases from different docs into the index map
	MergeLoop:
	for {
		select {
		case phrase,ok := <- phrases:
			if !ok {
				break MergeLoop
			}
			_,ok = index[phrase.Word]
			if !ok {
				index[phrase.Word] = []Phrase{}
			}

			index[phrase.Word] = append(index[phrase.Word], phrase)
		}
	}

	return index,nil
}

func cleanString(input string) string {
	re := regexp.MustCompile(`[,\'.;:?!—\-()\[\]{}"\/\\%&*+=<>\n\t\r]`)
	cleaned := re.ReplaceAllString(input, "")
	cleaned = strings.ReplaceAll(cleaned, "  ", " ") // remove double spaces.
	return cleaned
}

func tokenize(term string, val TermBody, stopWords map[string]bool, phrases chan Phrase) {
	defer wg.Done()
	val.Desc = cleanString(val.Desc)

	tokens := strings.Fields(val.Desc)

	counter := make(map[string]int)
	
	for _,token := range tokens {

		token := strings.ToLower(token)

		_,ok := stopWords[token]
		// Token is a stop word
		if ok {
			continue
		} 

		_,ok = counter[token]
		if !ok {
			counter[token] = 0
		}

		counter[token] += 1
	}

	for k,v := range counter {
		phrases <- Phrase{
			Word : k,
			Freq : v,
			Doc  : term,
			Wiki : val.Wiki,
			Desc : val.Desc,
		}
	}
}