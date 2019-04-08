package chinese

import (
	"container/heap"
	"errors"
	"reflect"
	"unicode/utf8"
)

// WordFreq represents a word and a frequency returned from a model
type WordFreq struct {
	Word           string
	LogProbability float32
}

// Model is a dictionary that can find all words that are
// a prefix of the given string.
type Model interface {
	FindAllPrefixesOf(input string) []WordFreq
}

// Segmenter will break chinese text into words, based on a single
// word frequency model that you provide
type Segmenter struct {
	model Model
}

// NewSegmenter returns a new text segmenter. When passed no arguments,
// it loads the default model from the web. Otherwise, you must create
// a model and pass it in as the first argument to use it.
func NewSegmenter(args ...interface{}) *Segmenter {
	var model Model
	var err error
	if len(args) == 0 {
		model, err = LoadModel()
		if err != nil {
			panic(err)
		}
	} else {
		var ok bool
		model, ok = args[0].(Model)
		if !ok {
			panic(errors.New("NewSegmenter: First argument must be a word model"))
		}
	}
	return &Segmenter{model: model}
}

type queue struct {
	dist  []float32
	items []int
}

func (q queue) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q queue) Less(i, j int) bool {
	return q.dist[q.items[i]] < q.dist[q.items[j]]
}

func (q queue) Len() int {
	return len(q.items)
}

func (q *queue) Push(x interface{}) {
	item := x.(int)
	q.items = append(q.items, item)
}

func (q *queue) Pop() interface{} {
	ret := q.items[len(q.items)-1]
	q.items = q.items[:len(q.items)-1]
	return ret
}

// Segment breaks the input string into separate words. Whitespace or other characters
// will be returned as their own entry in the result, so the original input
// can be obtained as the concatenation of the strings in the result.
func (s *Segmenter) Segment(inputStr string) []string {
	length := len(inputStr)
	dist := make([]float32, length+1, length+1)
	prev := make([]int, length+1, length+1)

	// Each entry: 0 -- no node existes
	// 1 -- node has been placed in queue
	// 2 -- node has been processed and is unrecognized
	valid := make([]byte, length+1, length+1)
	q := queue{dist: dist}

	// Add position 0 to queue
	heap.Push(&q, 0)

	jump := func(from, to int, prob float32) {
		// if no node, make one
		if valid[to] == 0 {
			valid[to] = 1
			heap.Push(&q, to)
		}

		// if dist > our dist + edge
		//log.Printf("to=%v %v < %v?", to, prob, dist[to])
		if dist[to] == 0 || prob < dist[to] {
			dist[to] = prob
			prev[to] = from
		}
	}

	// while the queue not empty,
	for q.Len() > 0 {
		// pop position from queue
		pos := heap.Pop(&q).(int)
		//log.Printf("Process position %v", pos)
		if pos == length {
			break
		}

		// find all words starting at that position
		matches := s.model.FindAllPrefixesOf(inputStr[pos:])
		//log.Printf("Got matches %v", matches)
		sofar := dist[pos]
		for _, match := range matches {
			//log.Printf("Process match %v", match)
			// find a node for it.
			l := len(match.Word)

			jump(pos, pos+l, sofar+match.LogProbability)
		}

		if len(matches) == 0 {
			valid[pos] = 2
			_, l := utf8.DecodeRuneInString(inputStr[pos:])
			jump(pos, pos+l, sofar)
		}
	}

	var output []string
	from := prev[length]
	to := length
	for {
		// join together consecutive unrecognized characters
		str := inputStr[from:to]
		for valid[from] == 2 && from > 0 && valid[prev[from]] == 2 {
			from, to = prev[from], from
			str = inputStr[from:to] + str
		}
		output = append(output, str)
		if from == 0 {
			break
		}
		from, to = prev[from], from
	}

	reverseSlice(output)
	return output
}

func reverseSlice(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
