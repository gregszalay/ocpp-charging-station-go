import React, { useRef, useState } from "react";
import Box from "@mui/material/Box";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import Divider from "@mui/material/Divider";
import Stack from "@mui/material/Stack";
import { CircularProgress } from "@mui/material";
import EVSEItem from "./util-components/EVSEItem";
import Typography from "@mui/material/Typography";
import appTheme from "../app/theme/AppTheme";

export default function EVSEList() {
  const [errorLoading, setErrorLoading] = useState<any>(null);
  const [isLoaded, setIsLoaded] = useState<boolean>(false);
  const [evses, setEvses] = useState<any[]>(["1", "2", "3"]);

  React.useEffect(() => {
    setInterval(() => {
      fetch("http://127.0.0.1:8090/evses/active/ids")
        .then((res) => {
          console.log(res);
          return res.json();
        })
        .then(
          (result) => {
            setIsLoaded(true);
            console.log(result);
            setEvses(result);
          },
          (errorLoading) => {
            console.log(errorLoading);
            setIsLoaded(true);
            setErrorLoading(errorLoading);
          }
        );
    }, 1000);
  }, []);

  return (
    <Box
      sx={{
        mb: 0,
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        overflow: "hidden",
        overflowY: "scroll",
        padding: 0,
        background: appTheme.palette.background.default,
      }}
    >
      <Typography
        variant="h1"
        component="div"
        padding={0}
        align="center"
        sx={{
          fontSize: 25,
          paddingBottom: 1,
          paddingTop: 1,
        }}
      >
        {errorLoading ? (
          <div>Failed to load page: {errorLoading.message}</div>
        ) : !isLoaded || !evses || evses.length === 0 ? (
          <Stack direction="column" justifyContent="center" alignItems="center">
            <CircularProgress color="info" />
          </Stack>
        ) : (
          <List disablePadding>
            {evses.map((evse: any) => {
              return (
                <ListItem key={evse} sx={{ margin: "0", padding: 0.5 }}>
                  <EVSEItem evseId={evse} />
                  <Divider />
                </ListItem>
              );
            })}
          </List>
        )}
      </Typography>
    </Box>
  );
}
