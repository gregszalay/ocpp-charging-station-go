import Box from "@mui/material/Box";
import React, { ReactElement, useState } from "react";
import Stack from "@mui/material/Stack";
import { CircularProgress } from "@mui/material";
import { EVSEStatusDataForUI } from "./typedefs/EVSEStatusDataForUI";
import { RFID } from "./typedefs/RFID";
import TextField from "@mui/material/TextField";
import Paper from "@mui/material/Paper";
import Grid from "@mui/material/Grid";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import appTheme from "../app/theme/AppTheme";
import { useNavigate, useParams } from "react-router-dom";

export default function EVSE() {
  let navigate = useNavigate();
  let { evseId } = useParams();
  const [error, setError] = useState<any>(null);
  const [isLoaded, setIsLoaded] = useState<boolean>(false);
  const [isRFIDReadInProgressStart, setIsRFIDReadInProgressStart] =
    useState<boolean>(false);
  const [isRFIDReadInProgressStop, setIsRFIDReadInProgressStop] =
    useState<boolean>(false);
  const [evseInfo, setevseInfo] = useState<EVSEStatusDataForUI>();

  React.useEffect(() => {
    setInterval(() => {
      if (!isRFIDReadInProgressStart && !isRFIDReadInProgressStop) {
        fetch("http://127.0.0.1:8090/chargestatus/" + evseId)
          .then((res) => {
            console.log(res);
            return res.json();
          })
          .then(
            (result) => {
              setIsLoaded(true);
              console.log(result);
              setevseInfo(result);
            },
            (error) => {
              console.log(error);
              setIsLoaded(true);
              setError(error);
            }
          );
      }
    }, 500);
  }, [isRFIDReadInProgressStart, isRFIDReadInProgressStop, evseId]);

  function getEVSEStatus(evse: EVSEStatusDataForUI): ReactElement {
    if (evse.isError === 1) {
      return <React.Fragment>ERROR</React.Fragment>;
    }
    if (evse.isCharging === 1) {
      return <React.Fragment>CHARGING</React.Fragment>;
    }
    if (evse.isChargingEnabled === 1) {
      return <React.Fragment>WAITING TO START</React.Fragment>;
    }
    if (evse.isEVConnected === 1) {
      return <React.Fragment>CONNECTED</React.Fragment>;
    }
    return <React.Fragment>NOT CONNECTED</React.Fragment>;
  }

  const startCharge = () => {
    setIsRFIDReadInProgressStart(true);
  };
  const stopCharge = () => {
    setIsRFIDReadInProgressStop(true);
  };

  const readRFIDAndStart = (rfid: RFID) => {
    console.info("You clicked the startcharge Chip.");
    fetch("http://127.0.0.1:8090/start/" + evseId, {
      method: "POST",
      mode: "no-cors",
      body: JSON.stringify(rfid),
      headers: { "Content-type": "application/json; charset=UTF-8" },
    })
    setIsRFIDReadInProgressStart(false);
  };
  const readRFIDAndStop = (rfid: RFID) => {
    console.info("You clicked the stopcharge Chip.");
    fetch("http://127.0.0.1:8090/stop/" + evseId, {
      method: "POST",
      mode: "no-cors",
      body: JSON.stringify(rfid),
      headers: { "Content-type": "application/json; charset=UTF-8" },
    })
    setIsRFIDReadInProgressStop(false);
  };

  if (error) {
    return <div>Error: {error.message}</div>;
  } else if (!isLoaded || !evseInfo) {
    return (
      <Stack
        sx={{ height: "50vh" }}
        direction="column"
        justifyContent="center"
        alignItems="center"
      >
        <CircularProgress color="info" />
      </Stack>
    );
  } else if (isRFIDReadInProgressStart) {
    return (
      <Box
        sx={{
          background: appTheme.palette.background.default,
          height: "100vh",
        }}
      >
        <TextField
          sx={{
            background: appTheme.palette.background.default,
            color: appTheme.palette.background.default,
          }}
          id="filled-password-input"
          label=""
          style={{ margin: 8 }}
          placeholder=""
          helperText=""
          fullWidth
          margin="normal"
          InputLabelProps={{
            shrink: true,
          }}
          type="password"
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            if (event.target.value.length >= 10) {
              readRFIDAndStart({ rfid: event.target.value });
            }
          }}
          autoFocus
        />
        <Typography
          variant="h1"
          component="div"
          padding={1}
          align="center"
          sx={{
            fontSize: 50,
            paddingBottom: 1,
            paddingTop: 1,
          }}
        >
          Please touch RFID card to the reader
        </Typography>
      </Box>
    );
  } else if (isRFIDReadInProgressStop) {
    return (
      <Box
        sx={{
          background: appTheme.palette.background.default,
          height: "100vh",
        }}
      >
        <TextField
          sx={{
            background: appTheme.palette.background.default,
            color: appTheme.palette.background.default,
          }}
          id="filled-password-input"
          label=""
          style={{ margin: 8 }}
          placeholder=""
          helperText=""
          fullWidth
          margin="normal"
          InputLabelProps={{
            shrink: true,
          }}
          type="password"
          onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
            if (event.target.value.length >= 10) {
              readRFIDAndStop({ rfid: event.target.value });
            }
          }}
          autoFocus
        />
        <Typography
          variant="h1"
          component="div"
          padding={1}
          align="center"
          sx={{
            fontSize: 50,
            paddingBottom: 1,
            paddingTop: 1,
          }}
        >
          Please touch RFID card to the reader
        </Typography>
      </Box>
    );
  } else {
    return (
      <Box
        sx={{
          m: 0,
          height: "100vh",
          width: "99%",
          background: appTheme.palette.background.default,
        }}
      >
        <Typography
          variant="h1"
          component="div"
          padding={2}
          align="center"
          sx={{
            fontSize: 40,
          }}
        >
          <Grid container spacing={1}>
            <Grid item xs={2}>
              <Button
                sx={{
                  width: "100%",
                  minHeight: "100%",
                  background: appTheme.palette.primary.main,
                }}
                variant="contained"
                onClick={() => {
                  navigate("/");
                }}
              >
                <Typography
                  variant="h1"
                  component="div"
                  padding={5}
                  align="center"
                  sx={{
                    fontSize: 25,
                    paddingBottom: 0,
                    paddingTop: 0,
                  }}
                >
                  {"<<"}
                </Typography>
              </Button>
            </Grid>
            <Grid item xs={10}>
              <Paper
                sx={{
                  background: appTheme.palette.secondary.light,
                  fontWeight: "bold",
                }}
              >
                {getEVSEStatus(evseInfo)}
              </Paper>
            </Grid>
            <Grid item xs={6}>
              <Paper
                sx={{
                  background: appTheme.palette.secondary.light,
                  fontWeight: "bold",
                  fontSize: 50,
                  paddingLeft: 4,
                  paddingRight: 4,
                }}
              >
                {evseInfo.powerActiveImport_kw_float.toFixed(3) + " kW"}
              </Paper>
            </Grid>
            <Grid item xs={6}>
              <Paper
                sx={{
                  background: appTheme.palette.secondary.light,
                  fontWeight: "bold",
                  fontSize: 50,
                  paddingLeft: 4,
                  paddingRight: 4,
                }}
              >
                {evseInfo.energyActiveNet_kwh_float.toFixed(3) + " kWh"}
              </Paper>
            </Grid>
            <Grid item xs={6}>
              <Button
                sx={{
                  width: "100%",
                  background: appTheme.palette.primary.main,
                }}
                variant="contained"
                onClick={() => startCharge()}
              >
                <Typography
                  variant="h1"
                  component="div"
                  padding={2}
                  align="center"
                  sx={{
                    fontSize: 45,
                    paddingBottom: 0,
                    paddingTop: 0,
                  }}
                >
                  Start Charge
                </Typography>
              </Button>
            </Grid>
            <Grid item xs={6}>
              <Button
                variant="contained"
                onClick={() => stopCharge()}
                sx={{
                  width: "100%",
                  background: appTheme.palette.primary.main,
                }}
              >
                <Typography
                  variant="h1"
                  component="div"
                  padding={2}
                  align="center"
                  sx={{
                    fontSize: 45,
                    paddingBottom: 0,
                    paddingTop: 0,
                  }}
                >
                  Stop Charge
                </Typography>
              </Button>
            </Grid>
          </Grid>
        </Typography>
      </Box>
    );
  }
}
