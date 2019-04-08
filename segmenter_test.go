package chinese_test

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/smhanov/chinese"
)

// ExampleSegmentation will create a simple model with some chinese words. Then it will split a sentence.
func ExampleNewWordModel() {
	model := chinese.NewWordModel()
	model.AddWord("他", 1)
	model.AddWord("儿", 1)
	model.AddWord("儿子", 2)
	model.AddWord("叫", 1)
	model.AddWord("名", 1)
	model.AddWord("名字", 2)
	model.AddWord("四", 1)
	model.AddWord("子", 1)
	model.AddWord("字", 1)
	model.AddWord("岁", 1)
	model.AddWord("的", 1)
	model.Finish()

	segmenter := chinese.NewSegmenter(model)
	segments := segmenter.Segment("我儿子四岁。他的名字叫Zack。")

	fmt.Printf("%s", strings.Join(segments, " "))
	// Output:
	// 我 儿子 四 岁 。 他 的 名字 叫 Zack。
}

func TestSegmentEnglish(t *testing.T) {
	model := chinese.NewWordModel()
	dict := "/usr/share/dict/words"
	if _, err := os.Stat(dict); os.IsNotExist(err) {
		t.Logf("Skipping full dictionary test; can't find %s", dict)
		return
	}

	file, err := os.Open(dict)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		model.AddWord(word, float32(len(word)))
	}

	model.Finish()

	segmenter := chinese.NewSegmenter(model)
	segments := segmenter.Segment("helloworld1500times..howareyou?")
	log.Printf("%v", segments)

}

/*func ExampleGse() {
	runtime.GC()
	PrintMemUsage()
	var seg gse.Segmenter
	seg.LoadDict()
	segments := seg.Segment([]byte("我儿子四岁。他的名字叫Zack。"))
	fmt.Println(gse.ToString(segments, true))
	PrintMemUsage()
	panic(nil)
	// Output:
	// 我 儿子 四岁 。 他 的 名字 叫 Zack。
}*/

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("\nAlloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// In this example, we load the default model (from the web) and use it to segment some text.
func ExampleNewSegmenter() {
	segments := chinese.NewSegmenter().Segment("我儿子四岁。他的名字叫Zack。")
	fmt.Printf("%s\n", strings.Join(segments, " "))
	// Output:
	// 我 儿子 四岁 。 他 的 名字 叫 Zack。
}
