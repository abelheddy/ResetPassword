/*
Aplicación de React para recuperación de contraseña

Esta aplicación consta de tres vistas principales:
1. RequestCode - Solicitud de código de verificación
2. VerifyCode - Verificación del código recibido por correo
3. ResetPassword - Formulario para establecer nueva contraseña

Flujo de trabajo:
1. Usuario ingresa su email y solicita un código
2. Servidor envía código al email (si existe)
3. Usuario ingresa el código recibido
4. Si el código es válido, puede establecer nueva contraseña
*/

import React, { useState } from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  useNavigate,
  useLocation
} from "react-router-dom";
import axios from "axios";

/*
Componente RequestCode
Muestra formulario para solicitar código de recuperación
*/
const RequestCode = () => {
  const [email, setEmail] = useState(""); // Estado para el email ingresado
  const [loading, setLoading] = useState(false); // Estado para controlar carga
  const navigate = useNavigate(); // Hook para navegación programática

  /*
  Función para enviar solicitud de código al servidor
  */
  const sendCode = async () => {
    // Validar que el email no esté vacío
    if (!email) {
      alert("Por favor ingresa tu correo electrónico");
      return;
    }

    setLoading(true); // Activar estado de carga
    
    try {
      // Enviar solicitud POST al servidor
      const response = await axios.post("http://localhost:8080/send-code", { 
        email: email.toLowerCase().trim() // Normalizar email (minúsculas y sin espacios)
      });
      
      // Si la respuesta es exitosa (200 OK)
      if (response.status === 200) {
        alert("Código enviado a tu correo");
        // Navegar a la vista de verificación, pasando el email como estado
        navigate("/verify-code", { state: { email: email.toLowerCase().trim() } });
      }
    } catch (error) {
      // Manejo de errores según el tipo de error
      if (error.response) {
        // El servidor respondió con un código de error
        if (error.response.status === 404) {
          alert("El correo no está registrado en nuestro sistema");
        } else if (error.response.data && error.response.data.error) {
          alert(`Error: ${error.response.data.error}`);
        } else {
          alert("Ocurrió un error inesperado. Intenta nuevamente.");
        }
      } else if (error.request) {
        // La solicitud fue hecha pero no hubo respuesta
        alert("No se pudo conectar con el servidor. Verifica tu conexión a internet.");
      } else {
        // Error al configurar la solicitud
        alert("Error al configurar la solicitud: " + error.message);
      }
      console.error("Error detallado:", error);
    } finally {
      setLoading(false); // Desactivar estado de carga
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "400px", margin: "0 auto" }}>
      <h2 style={{ textAlign: "center" }}>Recuperar Contraseña</h2>
      <div style={{ marginBottom: "15px" }}>
        <input
          type="email"
          placeholder="Correo electrónico"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          style={{ width: "100%", padding: "10px", fontSize: "16px" }}
        />
      </div>
      <button 
        onClick={sendCode}
        disabled={loading}
        style={{
          width: "100%",
          padding: "10px",
          backgroundColor: loading ? "#cccccc" : "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
          fontSize: "16px"
        }}
      >
        {loading ? "Enviando..." : "Enviar Código"}
      </button>
    </div>
  );
};

/*
Componente VerifyCode
Muestra formulario para ingresar el código de verificación recibido por correo
*/
const VerifyCode = () => {
  const [code, setCode] = useState(""); // Estado para el código ingresado
  const [loading, setLoading] = useState(false); // Estado para controlar carga
  const navigate = useNavigate(); // Hook para navegación programática
  const location = useLocation(); // Hook para acceder al estado de la ruta
  const email = location.state?.email || ""; // Obtener email del estado de navegación

  /*
  Función para verificar el código con el servidor
  */
  const verifyCode = async () => {
    // Validar que el código no esté vacío
    if (!code) {
      alert("Por favor ingresa el código de verificación");
      return;
    }

    setLoading(true); // Activar estado de carga
    
    try {
      // Enviar solicitud POST al servidor
      const response = await axios.post("http://localhost:8080/verify-code", { 
        email, 
        code 
      });
      
      // Si la verificación es exitosa (200 OK)
      if (response.status === 200) {
        // Navegar a la vista de restablecimiento, pasando email y código como estado
        navigate("/reset-password", { state: { email, code } });
      }
    } catch (error) {
      // Manejo de errores
      if (error.response) {
        if (error.response.status === 404) {
          alert("Código incorrecto o expirado");
        } else {
          alert("Error al verificar el código");
        }
      } else {
        alert("Error de conexión. Intenta nuevamente.");
      }
      console.error("Error detallado:", error);
    } finally {
      setLoading(false); // Desactivar estado de carga
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "400px", margin: "0 auto" }}>
      <h2 style={{ textAlign: "center" }}>Verificar Código</h2>
      <p style={{ textAlign: "center" }}>Se envió un código a: {email}</p>
      <div style={{ marginBottom: "15px" }}>
        <input
          type="text"
          placeholder="Código de verificación"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          style={{ width: "100%", padding: "10px", fontSize: "16px" }}
        />
      </div>
      <button 
        onClick={verifyCode}
        disabled={loading}
        style={{
          width: "100%",
          padding: "10px",
          backgroundColor: loading ? "#cccccc" : "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
          fontSize: "16px"
        }}
      >
        {loading ? "Verificando..." : "Verificar Código"}
      </button>
    </div>
  );
};

/*
Componente ResetPassword
Muestra formulario para establecer nueva contraseña
*/
const ResetPassword = () => {
  const [newPassword, setNewPassword] = useState(""); // Estado para la nueva contraseña
  const [loading, setLoading] = useState(false); // Estado para controlar carga
  const navigate = useNavigate(); // Hook para navegación programática
  const location = useLocation(); // Hook para acceder al estado de la ruta
  const { email, code } = location.state || {}; // Obtener email y código del estado

  /*
  Función para enviar la nueva contraseña al servidor
  */
  const resetPassword = async () => {
    // Validar que la contraseña no esté vacía
    if (!newPassword) {
      alert("Por favor ingresa una nueva contraseña");
      return;
    }

    setLoading(true); // Activar estado de carga
    
    try {
      // Enviar solicitud POST al servidor
      const response = await axios.post("http://localhost:8080/reset-password", { 
        email, 
        code, 
        newPassword 
      });
      
      // Si la actualización es exitosa (200 OK)
      if (response.status === 200) {
        alert("Contraseña actualizada correctamente");
        navigate("/"); // Redirigir a la página inicial
      }
    } catch (error) {
      alert("Error al cambiar la contraseña");
      console.error("Error detallado:", error);
    } finally {
      setLoading(false); // Desactivar estado de carga
    }
  };

  return (
    <div style={{ padding: "20px", maxWidth: "400px", margin: "0 auto" }}>
      <h2 style={{ textAlign: "center" }}>Restablecer Contraseña</h2>
      <div style={{ marginBottom: "15px" }}>
        <input
          type="password"
          placeholder="Nueva contraseña"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          style={{ width: "100%", padding: "10px", fontSize: "16px" }}
        />
      </div>
      <button 
        onClick={resetPassword}
        disabled={loading}
        style={{
          width: "100%",
          padding: "10px",
          backgroundColor: loading ? "#cccccc" : "#4CAF50",
          color: "white",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
          fontSize: "16px"
        }}
      >
        {loading ? "Actualizando..." : "Cambiar Contraseña"}
      </button>
    </div>
  );
};

/*
Componente principal que configura el enrutador
*/
const App = () => {
  return (
    <Router>
      <Routes>
        {/* Ruta para solicitar código (página inicial) */}
        <Route path="/" element={<RequestCode />} />
        {/* Ruta para verificar código */}
        <Route path="/verify-code" element={<VerifyCode />} />
        {/* Ruta para restablecer contraseña */}
        <Route path="/reset-password" element={<ResetPassword />} />
      </Routes>
    </Router>
  );
};

export default App;