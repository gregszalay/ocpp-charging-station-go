package displayserver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type EVSEStatusDataForUI struct {
	EnergyActiveNet_kwh_float  float64 `json:"energyActiveNet_kwh_float" yaml:"energyActiveNet_kwh_float"`
	PowerActiveImport_kw_float float64 `json:"powerActiveImport_kw_float" yaml:"powerActiveImport_kw_float"`
}

type UICallbacks struct {
	OnStartButtonPress func(int)
	OnStopButtonPress  func(int)
	OnGetChargeStatus  func(int) string
}

var callbacks UICallbacks

func onStart(w http.ResponseWriter, req *http.Request) {
	// get id from path parameter
	vars := mux.Vars(req)
	evseId, ok := vars["evseId"]
	if !ok {
		log.Error("failed to retrieve evseId from URL path parameters")
	}
	log.Info("Attempting to start TX for EVSE with id ", evseId)
	id_integer, err := strconv.Atoi(evseId)
	if err != nil {
		log.Error("Unable to convert evseId form string to int")
		return
	}
	callbacks.OnStartButtonPress(id_integer)
	w.WriteHeader(http.StatusOK)
}

func onStop(w http.ResponseWriter, req *http.Request) {
	// get id from path parameter
	vars := mux.Vars(req)
	evseId, ok := vars["evseId"]
	if !ok {
		log.Error("failed to retrieve evseId from URL path parameters")
	}
	log.Info("Attempting to start TX for EVSE with id ", evseId)
	id_integer, err := strconv.Atoi(evseId)
	if err != nil {
		log.Error("Unable to convert evseId form string to int")
		return
	}
	callbacks.OnStopButtonPress(id_integer)
	w.WriteHeader(http.StatusOK)
}

func onGetChargeStatus(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, callbacks.OnGetChargeStatus())
}

func Start(_callbacks UICallbacks) {
	callbacks = _callbacks
	http.HandleFunc("/start/{evseId}", onStart)
	http.HandleFunc("/stop/{evseId}", onStop)
	http.HandleFunc("/chargestatus/{evseId}", onGetChargeStatus)
	http.ListenAndServe(":8090", nil)
}
