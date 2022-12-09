export type EVSEStatusDataForUI = {
  isEVConnected: number;
  isChargingEnabled: number;
  isCharging: number;
  isError: number;
  energyActiveNet_kwh_float: number;
  powerActiveImport_kw_float: number;
};
