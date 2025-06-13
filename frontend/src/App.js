// src/App.js
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import HomeMenu from "./components/HomeMenu";
import SMTPConfigView from "./components/SMTPConfigView";
import PasswordRecovery from "./components/PasswordRecovery";
import LoginSetup from "./components/Setup/LoginSetup";
import SetupDBForm from "./components/Setup/SetupDBForm";
import Dashboard from "./components/Setup/Dashboard";
import { useEffect, useState } from "react";
import { checkSetupStatus, API_BASE_URL } from "./services/api"; // AÑADE LA IMPORTACIÓN AQUÍ

function SetupFlow() {
  return (
    <Routes>
      <Route path="/login-setup" element={<LoginSetup />} />
      <Route path="/setup-db" element={<SetupDBForm />} />
      <Route path="/dashboard" element={<Dashboard />} />
      <Route path="*" element={<Navigate to="/login-setup" />} />
    </Routes>
  );
}

function MainApp() {
  return (
    <Routes>
      <Route path="/" element={<HomeMenu />} />
      <Route path="/smtp-config" element={<SMTPConfigView />} />
      <Route path="/recover-password" element={<PasswordRecovery />} />
      <Route path="*" element={<Navigate to="/" />} />
    </Routes>
  );
}

function App() {
  const [setupDone, setSetupDone] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchStatus() {
      try {
        // Usa API_BASE_URL importada
        const response = await fetch(`${API_BASE_URL}/api/status?t=${Date.now()}`);
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log("Setup status:", data.setup);
        setSetupDone(data.setup);
        setError(null);
      } catch (err) {
        console.error("Error checking setup status:", err);
        setError(err.message);
        setSetupDone(false);
      }
    }
    fetchStatus();
  }, []);

  if (setupDone === null) {
    return (
      <div className="loading-container">
        <p>Cargando...</p>
        {error && <p className="error-message">{error}</p>}
      </div>
    );
  }

  return (
    <Router>
      <div className="app-container">
        {!setupDone ? <SetupFlow /> : <MainApp />}
      </div>
    </Router>
  );
}

export default App;