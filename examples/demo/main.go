package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/NoelM/minigo"
	"nhooyr.io/websocket"
)

func main() {
	apiKey := os.Getenv("SNCF_API_KEY")
	apiResponse := Response{}

	err := GetDepartures(apiKey, &apiResponse)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			err := GetDepartures(apiKey, &apiResponse)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
		if err != nil {
			log.Println(err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute*10)
		defer cancel()

		demo(c, ctx, &apiResponse)
	})

	err = http.ListenAndServe("192.168.1.34:3615", fn)
	log.Fatal(err)
}

func demo(c *websocket.Conn, ctx context.Context, apiResp *Response) {
	for {
		buf := []byte{}
		nbLines := 0

		apiResp.Mtx.RLock()
		for _, d := range apiResp.Departures {
			if nbLines+3 > 20 {
				break
			}

			//baseDepTime, _ := time.Parse("20060102T150405", d.Schedule.BaseDepartureDateTime)
			depTime, _ := time.Parse("20060102T150405", d.Schedule.DepartureDateTime)
			header := fmt.Sprintf("%s - %s %s", depTime.Format("15:04"), d.Informations.CommercialMode, d.Informations.Headsign)
			destination := fmt.Sprintf(">> %s", d.Informations.Direction)

			buf = append(buf, minigo.EncodeAttributes(minigo.InversionFond)...)
			buf = append(buf, minigo.EncodeMessage(header)...)
			buf = append(buf, minigo.GetMoveCursorReturn(1)...)
			buf = append(buf, minigo.EncodeAttributes(minigo.FondNormal)...)
			buf = append(buf, minigo.EncodeMessage(destination)...)
			buf = append(buf, minigo.GetMoveCursorReturn(2)...)

			nbLines += 3
		}
		apiResp.Mtx.RUnlock()

		c.Write(ctx, websocket.MessageBinary, buf)
		time.Sleep(10 * time.Second)
	}
}
