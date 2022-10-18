package main

import (
	"encoding/binary"
	"fmt"
	"github.com/coalalib/coalago"
	"github.com/kklash/genetic"
	log "github.com/ndmsystems/golog"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

import _ "net/http/pprof"

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	log.Init("zone", "emv", "name")
	mode := os.Args[1]

	//mode := "client"
	f, _ := os.Create(fmt.Sprintf("pprof_%s", mode))
	//addr := "165.22.120.237"
	addr := "147.182.133.37"
	//port := 5683
	port := 5555
	uri := "/tests/large"
	switch mode {
	case "server":

		srv := coalago.NewServer()
		fcontent, err := os.ReadFile("testfile20mb")
		if err != nil {
			print(err)
		}
		println(len(fcontent))
		println("9825")

		srv.AddGETResource(uri, func(message *coalago.CoAPMessage) *coalago.CoAPResourceHandlerResult {
			println("connect")
			return coalago.NewResponse(coalago.NewBytesPayload(fcontent), coalago.CoapCodeContent)
		})
		srv.AddPOSTResource(uri, func(message *coalago.CoAPMessage) *coalago.CoAPResourceHandlerResult {
			if message != nil {
				os.WriteFile("testfileDone", message.GetPayload(), 7777)
			}
			return coalago.NewResponse(coalago.NewEmptyPayload(), coalago.CoapCodeContent)
		})
		pprof.StartCPUProfile(f)
		srv.Listen(fmt.Sprintf(":%d", port))
		print("start wait")
		time.Sleep(time.Second * 30)
		print("stop")
		pprof.StopCPUProfile()

	case "client":
		cl := coalago.NewClient()
		var resp *coalago.Response
		var err error
		start := time.Now()
		resp, err = cl.GET(fmt.Sprintf("coap://%s:%d%s", addr, port, uri))
		if err != nil {
			print(err.Error())
		}
		if resp != nil {
			println(fmt.Sprintf("speed = %d Kbits", 1000.0*(int64(len(resp.Body))/time.Since(start).Milliseconds())/128))
			os.WriteFile("testfileDone", resp.Body, 7777)
		}
	case "poster":
		coalago.P, _ = strconv.ParseFloat(os.Args[2], 64)
		coalago.I, _ = strconv.ParseFloat(os.Args[3], 64)
		coalago.D, _ = strconv.ParseFloat(os.Args[4], 64)
		cl := coalago.NewClient()
		var resp *coalago.Response
		var err error
		fcontent, err := os.ReadFile("testfile20mb")
		pprof.StartCPUProfile(f)
		start := time.Now()
		resp, err = cl.POST(fcontent, fmt.Sprintf("coap://%s:%d%s", addr, port, uri))
		pprof.StopCPUProfile()
		println(fmt.Sprintf("speed = %d Kbits", 1000.0*(int64(len(fcontent))/time.Since(start).Milliseconds())/128))
		if err != nil {
			print("main.go : ")
			println(err.Error())
		}
		if resp != nil {
			println(string(resp.Body))
		}
	case "serverpost":
		cl := coalago.NewServer()
		pprof.StartCPUProfile(f)
		cl.AddPOSTResource(uri, func(message *coalago.CoAPMessage) *coalago.CoAPResourceHandlerResult {
			if message != nil {
				os.WriteFile("testfileDone", message.GetPayload(), 7777)
			}
			return coalago.NewResponse(coalago.NewEmptyPayload(), coalago.CoapCodeContent)
		})
		cl.Listen(fmt.Sprintf(":%d", port))
		time.Sleep(time.Second * 20)
		pprof.StopCPUProfile()
		return
	case "genetic_post":
		cl := coalago.NewClient()
		fcontent, _ := os.ReadFile("testfile20mb")
		genAlgo := genetic.NewPopulation(
			20,
			func() []byte {
				genome := make([]byte, 0, 32)
				tmp := make([]byte, 8)
				//ws := rand.Int63n(830) + 70
				ws := 70
				genome = genome[:0]
				binary.LittleEndian.PutUint64(tmp[:], uint64(ws))

				P := rand.Float64()
				I := rand.Float64()
				D := rand.Float64()
				println(fmt.Sprintf("WS: %d, P: %f, I: %f, D: %f", ws, P, I, D))

				genome = append(genome, tmp...)
				genome = append(genome, Float64ToByte(P)...)
				genome = append(genome, Float64ToByte(I)...)
				genome = append(genome, Float64ToByte(D)...)
				println(genome)
				return genome
			},
			genetic.CrossoverFunc[[]byte](ByteCrossover),

			//genetic.UniformCrossover[[]byte],
			genetic.StaticFitnessFunc(func(guess []byte) int {
				coalago.DEFAULT_WINDOW_SIZE = int(int64(binary.LittleEndian.Uint64(guess[:8])))
				coalago.P = ByteToFloat64(guess[8:16])
				coalago.I = ByteToFloat64(guess[16:24])
				coalago.D = ByteToFloat64(guess[24:])
				println(fmt.Sprintf("L WS: %d, P: %f, I: %f, D: %f", coalago.DEFAULT_WINDOW_SIZE, coalago.P, coalago.I, coalago.D))
				cl.POST(fcontent, fmt.Sprintf("coap://%s:%d%s", addr, port, uri))
				//binary.LittleEndian.PutUint64(guess[:8], uint64(coalago.RezWindowSize))
				if coalago.ReturnStatus {
					return int(coalago.BytesPerSec)
				} else {
					return 0 //int(coalago.BytesTotal / 100)
				}

			}),
			genetic.TournamentSelection[[]byte](3),
			func(guess []byte) {
				return
			},
		)
		genAlgo.Evolve(13107200, 10, 2)
		bestSolution, bestFitness := genAlgo.Best()
		coalago.DEFAULT_WINDOW_SIZE = int(int64(binary.LittleEndian.Uint64(bestSolution[:8])))
		coalago.P = ByteToFloat64(bestSolution[8:16])
		coalago.I = ByteToFloat64(bestSolution[16:24])
		coalago.D = ByteToFloat64(bestSolution[24:])
		println(fmt.Sprintf("WS: %d, P: %f, I: %f, D: %f, Best : %d", coalago.DEFAULT_WINDOW_SIZE, coalago.P, coalago.I, coalago.D, bestFitness))
	}

}

func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}
func ByteToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)

	return math.Float64frombits(bits)
}

func ByteCrossover(male, female []byte) ([]byte, []byte) {
	dnaLength := len(male)
	if len(female) != dnaLength {
		panic("cannot do point-based crossover with mismatching DNA length")
	}
	println("custom crossover")
	offspring1 := make([]byte, dnaLength)
	offspring2 := make([]byte, dnaLength)
	offspring2 = offspring2[:0]
	offspring1 = offspring1[:0]
	for i := 0; i < dnaLength; i += 8 {
		if rand.Float64() > 0.5 {
			offspring1 = append(offspring1, male[i:i+8]...)
			offspring2 = append(offspring2, female[i:i+8]...)
		} else {
			offspring1 = append(offspring1, female[i:i+8]...)
			offspring2 = append(offspring2, male[i:i+8]...)
		}
	}

	return offspring1, offspring2
}

//coap://165.22.120.237:5683/tests/large")
//9825
//147.182.133.37
//5.129.81.97
//138.197.191.160
