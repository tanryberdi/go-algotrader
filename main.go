package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/samjtro/go-tda/data"
)

var (
	m  sync.Mutex
	m1 sync.Mutex
	m2 sync.Mutex
	m3 sync.Mutex
	m4 sync.Mutex
	m5 sync.Mutex
	m6 sync.Mutex
)

func main() {
	start := time.Now()
	df, err := data.PriceHistory("TSLA", "month", "3", "daily", "1")

	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println(Set(df))
	fmt.Println(time.Since(start))
}

func Set(df []data.FRAME) []DATA {
	data5 := Chaikin(21, MACD(VWAP(df, RSI(EMA(10, RMA(3, df))))), df)

	return data5
}

type DATA struct {
	Price          float64
	RMA            float64
	EMA            float64
	RSI            float64
	VWAP           float64
	MACD           float64
	Chaikin        float64
	BollingerUpper float64
	BollingerLower float64
	IMI            float64
	MFI            float64
	PCR            float64
	OI             float64
}

// Calculates RMA, creates []DATA structure for use in the rest of the App, returns it
func RMA(n float64, data []data.FRAME) []DATA {
	d := []DATA{}

	m.Lock()

	for i, frame := range data {
		sum := 0.0

		if i >= int(n) {
			for a := int(n); a != 0; a-- {
				c, err := strconv.ParseFloat(data[i-a].Close, 64)

				if err != nil {
					log.Fatalf(err.Error())
				}

				sum += c
			}

			close, err := strconv.ParseFloat(frame.Close, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			d1 := DATA{
				Price: close,
				RMA:   sum / n,
			}

			d = append(d, d1)
		}
	}

	m.Unlock()
	return d
}

// Calculates EMA, adds to []DATA from RMA, return the []DATA
func EMA(n float64, d []DATA) []DATA {
	mult := 2 / (n + 1)

	m1.Lock()

	for i, _ := range d {
		sum := 0.0

		if i == int(n) {
			for a := 2; a != int(n)+1; a++ {
				sum += d[i-a].Price
			}

			a := d[i-1].Price - sum/n
			b := mult + sum/n
			ema := a * b

			d[i].EMA = ema
		} else if i > int(n) {
			prevEma := d[i-2].EMA
			ema := (d[len(d)-1].Price-prevEma)*mult + prevEma

			d[i].EMA = ema
		}
	}

	m1.Unlock()
	return d
}

func RSI(d []DATA) []DATA {
	gain := []float64{}
	loss := []float64{}
	var avgGain, avgLoss float64

	m2.Lock()

	for i, _ := range d {
		if i > 0 {
			diff := d[i].Price - d[i-1].Price

			if diff < 0 {
				loss = append(loss, diff)
			} else {
				gain = append(gain, diff)
			}
		}

		if i > 14 {
			avgGain = AverageGainLoss(i, d, gain)
			avgLoss = AverageGainLoss(i, d, loss)
			rs := avgGain / avgLoss
			d[i].RSI = 100 - (100 / (1 + rs))
		}
	}

	m2.Unlock()
	return d
}

func InitialAverageGainLoss(data []float64) float64 {
	sum := 0.0

	m3.Lock()
	defer m3.Unlock()

	for _, x := range data {
		sum += x
	}

	return sum
}

func AverageGainLoss(i int, d []DATA, data []float64) float64 {
	var initialAvgGainLoss, avgGainLoss float64

	m4.Lock()
	defer m4.Unlock()

	if len(data) >= (i - 13) {
		initialAvgGainLoss = InitialAverageGainLoss(data[(i - 13):])
		avgGainLoss = ((initialAvgGainLoss * 13) + (d[i].Price - d[i-1].Price)) / 14
	} else {
		initialAvgGainLoss = InitialAverageGainLoss(data)
		avgGainLoss = (initialAvgGainLoss*float64(len(data)-1) + (d[i].Price - d[i-1].Price)) / float64(len(data))
	}

	return avgGainLoss
}

func VWAP(df []data.FRAME, d []DATA) []DATA {
	var averagePrice float64

	m5.Lock()
	for i, x := range d {
		for _, y := range df {
			lo, err := strconv.ParseFloat(y.Lo, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			hi, err := strconv.ParseFloat(y.Hi, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			vol, err := strconv.ParseFloat(y.Volume, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			averagePrice += x.Price
			averagePrice += lo
			averagePrice += hi
			vwap := averagePrice / vol

			d[i].VWAP = vwap
		}
	}

	m5.Unlock()
	return d
}

func MACD(d []DATA) []DATA {
	var macd float64
	twentySixDayEMA := EMA(26, d)
	twelveDayEMA := EMA(12, d)

	m6.Lock()

	for _, x := range twentySixDayEMA {
		for _, y := range twelveDayEMA {
			for i, _ := range d {
				macd = y.EMA - x.EMA
				d[i].MACD = macd
			}
		}
	}

	m6.Unlock()
	return d
}

func Chaikin(p int, d []DATA, df []data.FRAME) []DATA {
	var Helper []DATA

	for _, x := range d {
		for _, y := range df {
			lo, err := strconv.ParseFloat(y.Lo, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			hi, err := strconv.ParseFloat(y.Hi, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			vol, err := strconv.ParseFloat(y.Volume, 64)

			if err != nil {
				log.Fatalf(err.Error())
			}

			n := ((x.Price - lo) - (hi - x.Price)) / (hi - lo)
			m := n * (vol * float64(p))
			adl := (m * float64(p-1)) + (m * float64(p))
			d1 := DATA{Price: adl}

			Helper = append(Helper, d1)
		}
	}

	threeDayEMA := EMA(3, Helper)
	tenDayEMA := EMA(10, Helper)

	for _, x := range threeDayEMA {
		for _, y := range tenDayEMA {
			for i, _ := range d {
				d[i].Chaikin = x.EMA - y.EMA
			}
		}
	}

	return d
}

func BollingerUpper() {

}

func BollingerLower() {

}

func IMI() {

}

func MFI() {

}

func PCR() {

}

func OI() {

}
