package displayserver

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type EVSEStatusDataForUI struct {
	EnergyActiveNet_kwh_float  float64 `json:"energyActiveNet_kwh_float" yaml:"energyActiveNet_kwh_float"`
	PowerActiveImport_kw_float float64 `json:"powerActiveImport_kw_float" yaml:"powerActiveImport_kw_float"`
}

type UICallbacks struct {
	OnStartButtonPress  func(int)
	OnStopButtonPress   func(int)
	OnGetChargeStatus   func(int) string
	OnGetEVSEsActiveIds func() string
}

var callbacks UICallbacks

func onStart(w http.ResponseWriter, req *http.Request) {
	if evseId, err := getEVSEIdFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Unable to find EVSE with this id")
	} else {
		callbacks.OnStartButtonPress(evseId)
		w.WriteHeader(http.StatusOK)
	}
}

func onStop(w http.ResponseWriter, req *http.Request) {
	if evseId, err := getEVSEIdFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Unable to find EVSE with this id")
	} else {
		callbacks.OnStopButtonPress(evseId)
		w.WriteHeader(http.StatusOK)
	}
}

func onGetChargeStatus(w http.ResponseWriter, req *http.Request) {
	if evseId, err := getEVSEIdFromReq(req); err != nil {
		log.Error("failed to retrieve evseId from URL path parameters")
		fmt.Fprintf(w, "Unable to find EVSE with this id")
	} else {
		fmt.Fprintf(w, callbacks.OnGetChargeStatus(evseId))
		w.WriteHeader(http.StatusOK)
	}
}

func onGetEVSEsActiveIds(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, callbacks.OnGetEVSEsActiveIds())
}

func getEVSEIdFromReq(req *http.Request) (int, error) {
	if evseId, ok := mux.Vars(req)["evseId"]; !ok {
		return -1, errors.New("failed to retrieve evseId from URL path parameters")
	} else {
		if id_integer, err := strconv.Atoi(evseId); !ok {
			return -1, err
		} else {
			return id_integer, nil
		}
	}
}

func Start(_callbacks UICallbacks) {
	callbacks = _callbacks
	http.HandleFunc("/start/{evseId}", onStart)
	http.HandleFunc("/stop/{evseId}", onStop)
	http.HandleFunc("/chargestatus/{evseId}", onGetChargeStatus)
	http.HandleFunc("/evses/active/ids", onGetEVSEsActiveIds)
	http.ListenAndServe(":8090", nil)
	for {
		time.Sleep(time.Millisecond * 50)
	}
}
