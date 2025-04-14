// PasswordRecovery.js
import React, { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import axios from "axios";
import './PasswordRecovery.css'

const PasswordRecovery = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [step, setStep] = useState(1); // 1: RequestCode, 2: VerifyCode, 3: ResetPassword
  const [formData, setFormData] = useState({
    email: "",
    code: "",
    newPassword: ""
  });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState({ text: "", type: "" });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const sendCode = async () => {
    if (!formData.email) {
      setMessage({ text: "Por favor ingresa tu correo electrónico", type: "error" });
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post("http://localhost:8080/send-code", { 
        email: formData.email.toLowerCase().trim()
      });
      
      if (response.status === 200) {
        setMessage({ text: "Código enviado a tu correo", type: "success" });
        setStep(2);
      }
    } catch (error) {
      handleError(error, "send-code");
    } finally {
      setLoading(false);
    }
  };

  const verifyCode = async () => {
    if (!formData.code) {
      setMessage({ text: "Por favor ingresa el código de verificación", type: "error" });
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post("http://localhost:8080/verify-code", { 
        email: formData.email,
        code: formData.code
      });
      
      if (response.status === 200) {
        setStep(3);
      }
    } catch (error) {
      handleError(error, "verify-code");
    } finally {
      setLoading(false);
    }
  };

  const resetPassword = async () => {
    if (!formData.newPassword) {
      setMessage({ text: "Por favor ingresa una nueva contraseña", type: "error" });
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post("http://localhost:8080/reset-password", { 
        email: formData.email,
        code: formData.code,
        newPassword: formData.newPassword
      });
      
      if (response.status === 200) {
        setMessage({ text: "Contraseña actualizada correctamente", type: "success" });
        navigate("/");
      }
    } catch (error) {
      handleError(error, "reset-password");
    } finally {
      setLoading(false);
    }
  };

  const handleError = (error, operation) => {
    let errorMessage = "Ocurrió un error. Intenta nuevamente.";
    
    if (error.response) {
      switch (error.response.status) {
        case 404:
          errorMessage = operation === "send-code" 
            ? "El correo no está registrado" 
            : "Código incorrecto o expirado";
          break;
        case 400:
          errorMessage = error.response.data?.error || "Datos inválidos";
          break;
        default:
          errorMessage = error.response.data?.error || errorMessage;
      }
    } else if (error.request) {
      errorMessage = "No se pudo conectar con el servidor";
    }

    setMessage({ text: errorMessage, type: "error" });
    console.error(`Error en ${operation}:`, error);
  };

  return (
    <div className="password-recovery-container">
      {message.text && (
        <div className={`alert ${message.type === "error" ? "alert-danger" : "alert-success"}`}>
          {message.text}
        </div>
      )}

      {step === 1 && (
        <div className="recovery-step">
          <h2>Recuperar Contraseña</h2>
          <div className="form-group">
            <input
              type="email"
              name="email"
              placeholder="Correo electrónico"
              value={formData.email}
              onChange={handleChange}
            />
          </div>
          <button onClick={sendCode} disabled={loading}>
            {loading ? "Enviando..." : "Enviar Código"}
          </button>
        </div>
      )}

      {step === 2 && (
        <div className="recovery-step">
          <h2>Verificar Código</h2>
          <p>Se envió un código a: {formData.email}</p>
          <div className="form-group">
            <input
              type="text"
              name="code"
              placeholder="Código de verificación"
              value={formData.code}
              onChange={handleChange}
            />
          </div>
          <button onClick={verifyCode} disabled={loading}>
            {loading ? "Verificando..." : "Verificar Código"}
          </button>
          <button 
            className="secondary-button" 
            onClick={() => setStep(1)}
          >
            Volver
          </button>
        </div>
      )}

      {step === 3 && (
        <div className="recovery-step">
          <h2>Restablecer Contraseña</h2>
          <div className="form-group">
            <input
              type="password"
              name="newPassword"
              placeholder="Nueva contraseña"
              value={formData.newPassword}
              onChange={handleChange}
            />
          </div>
          <button onClick={resetPassword} disabled={loading}>
            {loading ? "Actualizando..." : "Cambiar Contraseña"}
          </button>
        </div>
      )}
    </div>
  );
};

export default PasswordRecovery;