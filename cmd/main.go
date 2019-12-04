package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	aap_v2 "github.com/zakimal/aap.v2"
	"github.com/zakimal/aap.v2/log"
	"os"
	"time"
)

func main() {
	idFlag := flag.Uint64("id", 0, "ID")
	masterFlag := flag.Bool("master", false, "are you master?")

	flag.Parse()

	if *masterFlag { // is master
		master, err := aap_v2.NewMaster(*idFlag)
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("Start master %d", master.ID())

		go master.Listen()
		log.Info().Msgf("Listening for peers on %s", master.Address())

		time.Sleep(time.Second)

		config, err := os.Open(fmt.Sprintf("config/peers/%d.csv", master.ID()))
		if err != nil {
			panic(err)
		}
		defer config.Close()
		peerReader := csv.NewReader(config)
		peers, err := peerReader.Read()
		if err != nil {
			panic(err)
		}
		for _, peer := range peers {
			worker, err := master.Dial(peer)
			if err != nil {
				panic(err)
			}
			log.Info().Msgf("Dialed peer %d", worker.ID())
		}

		time.Sleep(time.Second)

		master.Run()
	} else { // is worker
		worker, err := aap_v2.NewWorker(*idFlag)
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("Start worker %d", worker.ID())

		go worker.Listen()
		log.Info().Msgf("Listening for peers on %s", worker.Address())

		time.Sleep(time.Second)

		config, err := os.Open(fmt.Sprintf("config/peers/%d.csv", worker.ID()))
		if err != nil {
			panic(err)
		}
		defer config.Close()
		peerReader := csv.NewReader(config)
		peers, err := peerReader.Read()
		if err != nil {
			panic(err)
		}
		for _, peer := range peers {
			worker, err := worker.Dial(peer)
			if err != nil {
				panic(err)
			}
			log.Info().Msgf("Dialed peer %d", worker.ID())
		}

		time.Sleep(time.Second)

		worker.Run()
	}
}
