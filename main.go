package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	min, max, count int32
	sum             int64
}

func (r *Result) merge(toMerge Result) {
	r.min = min(r.min, toMerge.min)
	r.max = max(r.max, toMerge.max)
	r.count += toMerge.count
	r.sum += toMerge.sum
}

type ResultMap map[string]Result

func (rs *ResultMap) merge(toMerge ResultMap) {
	for station, result := range toMerge {
		if existingResult, ok := (*rs)[station]; ok {
			existingResult.merge(result)
			(*rs)[station] = existingResult
		} else {
			(*rs)[station] = result
		}
	}
}

/* func (rs *ResultMap) add(station string, value int32) {
	// fmt.Printf("adding: %s, %d \n", station, value)
	if result, ok := (*rs)[station]; ok {
		result.count++
		result.max = max(value, result.max)
		result.min = min(value, result.min)
		result.sum += int64(value)
		(*rs)[station] = result
	} else {
		(*rs)[station] = Result{
			min:   value,
			max:   value,
			count: 1,
			sum:   int64(value),
		}
	}
} */

type computedResult struct {
	city          string
	min, avg, max float64
}

func (rs *ResultMap) toString() string {
	resultArr := make([]computedResult, len((*rs)))

	var count int
	for city, calculated := range *rs {
		resultArr[count] = computedResult{
			city: city,
			min:  round(float64(calculated.min) / 10),
			avg:  round(float64(calculated.sum) / 10 / float64(calculated.count)),
			max:  round(float64(calculated.max) / 10),
		}
		count++
	}
	sort.Slice(resultArr, func(i, j int) bool {
		return resultArr[i].city < resultArr[j].city
	})

	var stringsBuilder strings.Builder
	for _, i := range resultArr {
		stringsBuilder.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f, ", i.city, i.min, i.avg, i.max))
	}
	return stringsBuilder.String()[:stringsBuilder.Len()-2]
}

func round(x float64) float64 {
	rounded := math.Round(x * 10)
	if rounded == 0.0 {
		return 0.0
	}
	return rounded / 10
}

func main() {
	f, err := os.Create("./cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	start := time.Now()

	bufSize := 1024 * 1024 * 12 // optimal performance on my m1 mac air

	// fmt.Println(evaluate("./measurements2_100m.txt", bufSize))
	evaluate("./measurements2.txt", bufSize)
	fmt.Println("Execution time: ", time.Since(start))
}

func evaluate(path string, bufSize int) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	results := make(ResultMap)

	buf := make([]byte, bufSize)
	readStart := 0
	count := 0 // count chunks
	for {
		n, err := f.Read(buf[readStart:])
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if readStart+n == 0 {
			break
		}

		newline := bytes.LastIndexByte(buf[:readStart+n], '\n')
		if newline < 0 {
			break
		}

		chunk := buf[:newline+1]
		leftover := buf[newline+1 : readStart+n]

		count++
		fmt.Print(count, " ")
		chunkResults, err := processChunk(chunk)
		if err != nil {
			log.Fatal(err)
		}
		results.merge(chunkResults)

		readStart = copy(buf, leftover)
	}

	return results.toString()
}

func processChunk(chunk []byte) (ResultMap, error) {
	results := make(ResultMap)
	lines := bytes.Split(chunk, []byte("\n"))
	for _, line := range lines {
		// fmt.Print(string(line))
		if len(line) == 0 {
			break
		}
		lineSplit := bytes.Split(line, []byte(";"))
		// fmt.Print(string(lineSplit[0]), string(lineSplit[1]), "\n")
		station := string(lineSplit[0])
		valueRaw := string(lineSplit[1])
		valueStr := strings.Join(strings.Split(valueRaw, "."), "")
		i, _ := strconv.ParseInt(valueStr, 10, 32)

		value := int32(i)
		// results.add(station, value)
		if result, ok := results[station]; ok {
			result.count++
			result.max = max(value, result.max)
			result.min = min(value, result.min)
			result.sum += int64(value)
			results[station] = result
		} else {
			results[station] = Result{
				min:   value,
				max:   value,
				count: 1,
				sum:   int64(value),
			}
		}
	}
	// fmt.Print(results.toString())
	return results, nil
}
