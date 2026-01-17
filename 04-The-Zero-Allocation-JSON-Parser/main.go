package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
)

type SensorData struct {
	SensorID string  `json:"sensor_id"`
	Readings float64 `json:"readings"`
}

var sensorPool = sync.Pool{
	New: func() any {
		return &SensorData{}
	},
}

func discard(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	_, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	depth := 1
	for depth > 0 {
		to, err := d.Token()
		if err != nil {
			return err
		}
		if del, ok := to.(json.Delim); ok {
			switch del {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
		}
	}
	return nil
}

func SensorParser(r io.Reader) error {
	d := json.NewDecoder(r)

	for d.More() {
		val := sensorPool.Get().(*SensorData)
		// consuming the opening '{'

		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// check to ensure that whatever was consumed is actually a '{'
		if delim, ok := t.(json.Delim); !ok || delim != '{' {
			return fmt.Errorf("expected '{' but got %v", t)
		}

		for d.More() {
			// now we are in the array and are asking if there are more things to consume
			k, err := d.Token()
			if err != nil {
				return err
			}
			key, _ := k.(string)
			switch key {
			case "sensor_id":
				// we then consume the value in the key
				dt, _ := d.Token()
				sensorID := dt.(string)
				val.SensorID = sensorID

			case "readings":
				_, _ = d.Token()
				if d.More() {
					r, err := d.Token()
					if err != nil {
						return err
					}
					keyReading, _ := r.(float64)
					val.Readings = keyReading
				}
				// check if there are more items in the arrray []
				for d.More() {
					discard(d)
				}
				// we consume the final "]"
				if _, err := d.Token(); err != nil {
					return err
				}
			default:
				err := discard(d)
				if err == io.EOF {
					return errors.New("end of file error")
				}
				if err != nil {
					return err
				}
			}

		}
		sensorPool.Put(val)
	}
	return nil
}

func main() {
	err := SensorParser(io.Reader(nil))
	if err != nil {
		log.Print(err)
	}
}
