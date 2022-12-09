import { BrowserRouter as Router, Route, Link, Routes } from "react-router-dom";
import EVSEList from "../evselist/EVSEList";
import EVSE from "../evsedetails/EVSEDetails";

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<EVSEList />} />
        <Route path="/evses/:evseId" element={<EVSE />} />
      </Routes>
    </Router>
  );
}

export default App;
