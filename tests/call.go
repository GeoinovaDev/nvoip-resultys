package main

import (
	"fmt"

	"github.com/GeoinovaDev/lower-resultys/net/request"
	"github.com/GeoinovaDev/nvoip-resultys/nvoip"
)

func main() {
	request.ProxyURL = "http://127.0.0.1:8888"
	dialer := nvoip.New("45438001", "d@P7-43238W!dE2CB-Pj7edRKOCeLa-OCeQ#JL3", 100, 10)
	dialer.Timeout = 180
	dialer.CallerID = "45438001"
	// telefone := "62981876424"
	telefone := "6239426655"

	parameters := nvoip.RequestParameter{
		PhoneFrom: "45438001",
		PhoneTo:   telefone,
		Audios: []nvoip.AudioParameter{
			nvoip.AudioParameter{
				TextOrAudioUrl: "https://gravacoes-nvoip.s3.sa-east-1.amazonaws.com/tests/audio1.mp3",
				Position:       1,
			},
			nvoip.AudioParameter{
				TextOrAudioUrl: "manoel",
				Position:       2,
			},
			nvoip.AudioParameter{
				TextOrAudioUrl: "https://gravacoes-nvoip.s3.sa-east-1.amazonaws.com/tests/audio2.mp3",
				Position:       3,
			},
		},
		Dtmf: []nvoip.DtmfParameter{
			nvoip.DtmfParameter{
				TextOrAudioUrl: "https://gravacoes-nvoip.s3.sa-east-1.amazonaws.com/tests/audio3.mp3",
				Position:       4,
				MaxTime:        "4000",
				Timeout:        "30",
				MinNumberKey:   "0",
				MaxNumberKey:   "1",
			},
		},
	}

	dialer.CallQueued(parameters, func(response *nvoip.ResponseParameter, err error) {
		fmt.Printf("%+v", response)
	})

	fmt.Scanln()
	// nvoip := client.New("fee84e5862d54c5915831797de2e1f19d8541", 5)
	// parameters := client.RequestParameter{
	// 	PhoneFrom: "45438001",
	// 	PhoneTo:   "62982334440",
	// 	Audios: []client.AudioParameter{
	// 		client.AudioParameter{
	// 			TextOrAudioUrl: "https://new.resultys.com.br/audio/antigo/audio1.mp3",
	// 			Position:       1,
	// 		},
	// 		client.AudioParameter{
	// 			TextOrAudioUrl: "geoinova soluções",
	// 			Position:       2,
	// 		},
	// 		client.AudioParameter{
	// 			TextOrAudioUrl: "https://new.resultys.com.br/audio/antigo/audio3.mp3",
	// 			Position:       4,
	// 		},
	// 	},
	// 	Dtmf: []client.DtmfParameter{
	// 		client.DtmfParameter{
	// 			TextOrAudioUrl: "https://new.resultys.com.br/audio/antigo/audio2.mp3",
	// 			Position:       3,
	// 			MaxTime:        "4000",
	// 			Timeout:        "30",
	// 			MinNumberKey:   "0",
	// 			MaxNumberKey:   "1",
	// 		},
	// 	},
	// }

	// nvoip.CallQueued(parameters, func(response *client.ResponseParameter, err error) {
	// 	if err != nil {
	// 		println("error")
	// 	}

	// 	println(response)
	// })
	// // response, err := nvoip.Call(parameters)

	// fmt.Scanln()
}
