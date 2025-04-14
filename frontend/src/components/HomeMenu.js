import { useNavigate } from 'react-router-dom';

import './HomeMenu.css';

const HomeMenu = () => {
  const navigate = useNavigate();

  return (
    <div className="home-menu">
      <h2>Sistema de Autenticación</h2>
      
      <div className="menu-options">
        <button
          onClick={() => navigate('/recover-password')}
          className="btn btn-primary"
        >
          Recuperar Contraseña
        </button>
        
        <button
          onClick={() => navigate('/smtp-config')}
          className="btn btn-secondary"
        >
          Configuración SMTP
        </button>
      </div>
    </div>
  );
};

export default HomeMenu;