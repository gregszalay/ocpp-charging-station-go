import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import { useNavigate } from "react-router-dom";
import appTheme from "../../app/theme/AppTheme";

interface Props {
  evseId: number;
}

export default function EVSEItem(props: Props) {
  let navigate = useNavigate();

  return (
    <Paper
      sx={{
        background: appTheme.palette.secondary.light,
        width: "100%",
        maxHeight: "35vh",
        m: "0",
      }}
    >
      <Stack
        direction="row"
        justifyContent="space-between"
        alignContent="center"
        padding={0}
      >
        <Typography
          variant="h1"
          padding={3}
          align="center"
          sx={{
            fontSize: 35,
          }}
        >
          {"EVSE " + props.evseId}
        </Typography>
        <Button
          variant="contained"
          sx={{
            width: "50vw",
            background: appTheme.palette.primary.main,
            m: 1,
            p: 0,
          }}
          onClick={() => {
            navigate("/evses/" + props.evseId);
          }}
        >
          <Typography
            variant="h1"
            component="div"
            padding={0}
            align="center"
            sx={{
              fontSize: 35,
              paddingBottom: 0,
              paddingTop: 0,
              m: 0,
            }}
          >
            {"Details >>"}
          </Typography>
        </Button>
      </Stack>
    </Paper>
  );
}
