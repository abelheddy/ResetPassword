// App.js
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomeMenu from './components/HomeMenu';
import SMTPConfigView from './components/SMTPConfigView';
import PasswordRecovery from './components/PasswordRecovery';

function App() {
  return (
    <Router>
      <div className="app-container">
        <Routes>
          <Route path="/" element={<HomeMenu />} />
          <Route path="/smtp-config" element={<SMTPConfigView />} />
          <Route path="/recover-password" element={<PasswordRecovery />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;