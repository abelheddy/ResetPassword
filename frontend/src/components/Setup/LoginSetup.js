import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { loginSetup } from "../../services/api"; // Asegúrate de que la ruta sea correcta

export default function LoginSetup() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    user: "",
    pass: ""
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    
    try {
      await loginSetup(formData.user, formData.pass);
      navigate("/setup-db");
    } catch (err) {
      setError(err.message);
      console.error("Error en login:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: "400px", margin: "2rem auto", padding: "1rem" }}>
      <h2>Login de Configuración</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.5rem" }}>Usuario:</label>
          <input
            type="text"
            value={formData.user}
            onChange={(e) => setFormData({...formData, user: e.target.value})}
            style={{ width: "100%", padding: "0.5rem" }}
          />
        </div>
        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.5rem" }}>Contraseña:</label>
          <input
            type="password"
            value={formData.pass}
            onChange={(e) => setFormData({...formData, pass: e.target.value})}
            style={{ width: "100%", padding: "0.5rem" }}
          />
        </div>
        <button 
          type="submit" 
          disabled={loading}
          style={{
            padding: "0.5rem 1rem",
            background: loading ? "#ccc" : "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer"
          }}
        >
          {loading ? "Verificando..." : "Ingresar"}
        </button>
        {error && <p style={{ color: "red", marginTop: "1rem" }}>{error}</p>}
      </form>
    </div>
  );
}