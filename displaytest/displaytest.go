package displaytest

import (
	"time"

	"github.com/gregszalay/ocpp-charging-station-go/displayserver"
	log "github.com/sirupsen/logrus"
)

var evses []displayserver.EVSEStatusDataForUI

func RunDisplayTest() {

	evses = []displayserver.EVSEStatusDataForUI{
		{
			IsEVConnected:              1,
			IsChargingEnabled:          0,
			IsCharging:                 0,
			IsError:                    0,
			EnergyActiveNet_kwh_float:  float64(12.56) / 100,
			PowerActiveImport_kw_float: float64(687.65) / 100,
		},
		{
			IsEVConnected:              1,
			IsChargingEnabled:          0,
			IsCharging:                 0,
			IsError:                    0,
			EnergyActiveNet_kwh_float:  float64(12.56) / 100,
			PowerActiveImport_kw_float: float64(687.65) / 100,
		},
		{
			IsEVConnected:              1,
			IsChargingEnabled:          0,
			IsCharging:                 0,
			IsError:                    0,
			EnergyActiveNet_kwh_float:  float64(12.56) / 100,
			PowerActiveImport_kw_float: float64(687.65) / 100,
		},
		{
			IsEVConnected:              1,
			IsChargingEnabled:          0,
			IsCharging:                 0,
			IsError:                    0,
			EnergyActiveNet_kwh_float:  float64(12.56) / 100,
			PowerActiveImport_kw_float: float64(687.65) / 100,
		},
	}

	// Set up the UI logic
	UI_callbacks := &displayserver.UICallbacks{
		OnStartButtonPress: func(evseId int, rfid string) {
			log.Info("Start button pressed for evse: ", evseId)
			if evseId > len(evses)-1 {
				log.Error("No such evse")
				return

			}
			evses[evseId].IsChargingEnabled = 1
			evses[evseId].IsCharging = 1
		},
		OnStopButtonPress: func(evseId int, rfid string) {
			log.Info("Stop button pressed for evse: ", evseId)
			if evseId > len(evses)-1 {
				log.Error("No such evse")
				return
			}
			evses[evseId].IsChargingEnabled = 0
			evses[evseId].IsCharging = 0
		},
		OnGetChargeStatus: func(evseId int) displayserver.EVSEStatusDataForUI {
			if evseId > len(evses)-1 {
				log.Error("No such evse")
				return displayserver.EVSEStatusDataForUI{}
			}
			return evses[evseId]
		},
		OnGetEVSEsActiveIds: func() []int {
			result := make([]int, 0)
			for index, _ := range evses {
				result = append(result, index)
			}
			return result
		},
	}

	go displayserver.Start(*UI_callbacks)

	go func() {
		for {
			for _, evseStatus := range evses {
				if evseStatus.IsCharging == 1 && evseStatus.IsError == 0 {
					evseStatus.PowerActiveImport_kw_float = 23.5
					evseStatus.EnergyActiveNet_kwh_float += 0.2
				}
			}
			time.Sleep(time.Second)
		}
	}()

}
