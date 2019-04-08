package chinese

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/smhanov/dawg"
)

// WordModel is a structure that can both find all words that are prefixes of a given string,
// and return the log frequencies of those words.
type WordModel struct {
	builder     dawg.Builder
	finder      dawg.Finder
	frequencies []float32
}

// NewWordModel returns a new word model. You must add words to this using AddWord()
// and then call Finish() before using it.
func NewWordModel() *WordModel {
	return &WordModel{
		builder: dawg.New(),
	}
}

// AddWord adds a word and log frequency to the model.
// If frequency of all words is not known, use the length of the word.
// This will cause the segmenter to try to break the text into the fewest
// number of words.
//
// Words must be added in alphabetical order, and must not be repeated.
// Otherwise, it will cause a panic()
func (m *WordModel) AddWord(word string, freqCount float32) {
	m.builder.Add(word)
	m.frequencies = append(m.frequencies, freqCount)
}

// Finish signals that the word model is finished and ready to be used.
func (m *WordModel) Finish() {
	m.finder = m.builder.Finish()
	log.Printf("Model has %v words, %v edges", m.finder.NumAdded(), m.finder.NumEdges())

	// sum frequencies, then convert to log (1/p)
	var sum float64
	for _, freq := range m.frequencies {
		sum += float64(freq)
	}

	log.Printf("freq sum is %v", sum)
	logT := math.Log2(sum)

	for i, freq := range m.frequencies {
		m.frequencies[i] = float32(logT - math.Log2(float64(freq)))
	}
}

// FindAllPrefixesOf finds all prefixes of the input string that are
// words, and returns their log inverse probabilities.
func (m *WordModel) FindAllPrefixesOf(input string) []WordFreq {
	matches := m.finder.FindAllPrefixesOf(input)
	results := make([]WordFreq, len(matches), len(matches))
	for i, match := range matches {
		results[i].Word = match.Word
		results[i].LogProbability = m.frequencies[match.Index]
	}

	return results
}

func readerFromAnything(arg interface{}) (io.Reader, error) {

	switch v := arg.(type) {
	case io.Reader:
		return v, nil

	case string:
		if strings.HasPrefix(v, "https:") || strings.HasPrefix(v, "http:") {
			resp, err := http.Get(v)
			if err != nil {
				return nil, err
			}
			return resp.Body, nil
		}

		var r io.Reader
		f, err := os.Open(v)
		if err != nil {
			return nil, err
		}

		r = f

		ext := filepath.Ext(v)

		if ext == ".gz" {
			r, err = gzip.NewReader(r)
		} else if ext == ".bz2" {
			r = bzip2.NewReader(r)
		}

		if err != nil {
			f.Close()
			return nil, err
		}
	}

	return nil, errors.New("Don't know how to open that thing")
}

// LoadModel returns a model that you open from the given file.
// The model is a text file. Each line is a word and raw frequency (not log)
// separated by space. The format is inferred from the first line.
// If the file ends in .bz2 or .gz, it will be decompressed. If the file is an URL,
// it will be fetched. If the file is an io.Reader, it will be read from.
func LoadModel(args ...interface{}) (*WordModel, error) {
	var arg interface{}
	if len(args) == 0 {
		arg = "https://raw.githubusercontent.com/go-ego/gse/master/data/dict/dictionary.txt"
	} else {
		arg = args[0]
	}

	r, err := readerFromAnything(arg)
	if err != nil {
		return nil, err
	}

	model := NewWordModel()

	scanner := bufio.NewScanner(r)
	var words []WordFreq
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, errors.New("Error: not enough fields")
		}
		word := fields[0]
		freq, _ := strconv.ParseFloat(fields[1], 32)
		if len([]rune(word)) < 2 {
			//freq = 2
		}
		words = append(words, WordFreq{
			Word:           word,
			LogProbability: float32((freq)),
		})
	}

	sort.Slice(words, func(a, b int) bool {
		return words[a].Word < words[b].Word
	})

	prev := ""
	for _, word := range words {
		if word.Word == prev {
			continue
		}
		prev = word.Word
		model.AddWord(word.Word, word.LogProbability)
	}

	model.Finish()

	return model, nil
}
