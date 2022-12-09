package displayserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type EVSEStatusDataForUI struct {
	IsEVConnected              int     `json:"isEVConnected" yaml:"isEVConnected"`
	IsChargingEnabled          int     `json:"isChargingEnabled" yaml:"isChargingEnabled"`
	IsCharging                 int     `json:"isCharging" yaml:"isCharging"`
	IsError                    int     `json:"isError" yaml:"isError"`
	EnergyActiveNet_kwh_float  float64 `json:"energyActiveNet_kwh_float" yaml:"energyActiveNet_kwh_float"`
	PowerActiveImport_kw_float float64 `json:"powerActiveImport_kw_float" yaml:"powerActiveImport_kw_float"`
}

type UICallbacks struct {
	OnStartButtonPress  func(int, string)
	OnStopButtonPress   func(int, string)
	OnGetChargeStatus   func(int) EVSEStatusDataForUI
	OnGetEVSEsActiveIds func() []int
}

var callbacks UICallbacks

type RFID_BODY struct {
	Rfid string `json:"rfid" yaml:"rfid"`
}

func allowCORS(w http.ResponseWriter) {
	//Allow CORS here By * or specific origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	//w.WriteHeader(http.StatusOK)
}

func onStart(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	if evseId, err := getEVSEIDFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Received illegal value for EVSE Id!")
		return
	} else {
		var rfid RFID_BODY
		err := json.NewDecoder(req.Body).Decode(&rfid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Info("Starting evse ", evseId, ". RFID: ", rfid.Rfid)
		callbacks.OnStartButtonPress(evseId, rfid.Rfid)
		//w.WriteHeader(http.StatusOK)
	}
}

func onStop(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	if evseId, err := getEVSEIDFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Received illegal value for EVSE Id!")
	} else {
		var rfid RFID_BODY
		err := json.NewDecoder(req.Body).Decode(&rfid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Info("Stopping evse ", evseId, ". RFID: ", rfid.Rfid)
		callbacks.OnStopButtonPress(evseId, rfid.Rfid)
		//w.WriteHeader(http.StatusOK)
	}
}

func onGetChargeStatus(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	if evseId, err := getEVSEIDFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Received illegal value for EVSE Id!")
	} else {
		data := callbacks.OnGetChargeStatus(evseId)
		json_str, err := json.Marshal(data)
		if err != nil {
			log.Error("Failed to marshal UI data")
		}
		//w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json") // and this
		w.Write(json_str)
	}
}

func onGetEVSEsActiveIds(w http.ResponseWriter, req *http.Request) {
	allowCORS(w)
	data := callbacks.OnGetEVSEsActiveIds()
	json_str, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal UI data")
	}
	w.Header().Set("Content-Type", "application/json") // and this
	w.Write(json_str)
}

func getEVSEIDFromReq(req *http.Request) (int, error) {
	if evseId, ok := mux.Vars(req)["EVSEID"]; !ok {
		return -1, errors.New("failed to retrieve RFID from URL path parameters")
	} else {
		if id_integer, err := strconv.Atoi(evseId); !ok {
			return -1, err
		} else {
			return id_integer, nil
		}
	}
}
func getRFIDFromReq(req *http.Request) (string, error) {
	if rfid, ok := mux.Vars(req)["RFID"]; !ok {
		return "", errors.New("failed to retrieve RFID from URL path parameters")
	} else {
		return rfid, nil
	}
}

func Start(_callbacks UICallbacks) {
	callbacks = _callbacks

	var waitgroup sync.WaitGroup

	waitgroup.Add(1)
	go func() {
		fmt.Println("Creating http server...")
		router := NewRouter()
		log.Fatal(http.ListenAndServe(":8090", router))
		waitgroup.Done()
	}()

}
