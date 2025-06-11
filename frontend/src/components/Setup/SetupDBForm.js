import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { saveDBConfig, checkSetupStatus, API_BASE_URL } from "../../services/api";

export async function checkSystemStatus() {
  try {
    const response = await fetch(`${API_BASE_URL}/api/status`);

    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      const text = await response.text();
      throw new Error(`Respuesta no JSON: ${text.substring(0, 100)}...`);
    }

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Error verificando estado del sistema");
    }

    return await response.json();
  } catch (err) {
    console.error("Error en checkSystemStatus:", err);
    throw err;
  }
}

export default function SetupDBForm() {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    host: "localhost",
    port: "5432",
    user: "",
    password: "",
    dbname: ""
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      // 1. Guardar configuración de la DB
      const saveResponse = await saveDBConfig(formData);
      console.log("Respuesta saveDBConfig:", saveResponse);

      // PRUEBA: forzar redirección para ver si navigate funciona
      // navigate("/dashboard");

      // 2. Verificar si el setup está completo
      if (saveResponse.setup === true) {
        console.log("Setup completado según saveResponse. Navegando a /dashboard");
        navigate("/dashboard");
      } else {
        // Si el backend no devolvió explícitamente setup: true, verificamos el estado
        const status = await checkSetupStatus();
        console.log("Respuesta checkSetupStatus:", status);

        if (status.setup) {
          console.log("Setup completado según checkSetupStatus. Navegando a /dashboard");
          navigate("/dashboard");
        } else {
          throw new Error("La configuración no se completó correctamente en el backend");
        }
      }
    } catch (err) {
      setError(err.message || "Error al guardar la configuración");
      console.error("Error en SetupDBForm:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: "600px", margin: "2rem auto", padding: "1rem" }}>
      <h2>Configuración de Base de Datos</h2>
      {error && <div style={{ color: "red", marginBottom: "1rem" }}>{error}</div>}

      <form onSubmit={handleSubmit}>
        <div style={{ display: "flex", gap: "1rem", marginBottom: "1rem" }}>
          <div style={{ flex: 1 }}>
            <label style={{ display: "block", marginBottom: "0.5rem" }}>Host:</label>
            <input
              type="text"
              value={formData.host}
              onChange={(e) => setFormData({ ...formData, host: e.target.value })}
              style={{ width: "100%", padding: "0.5rem" }}
              required
            />
          </div>
          <div style={{ width: "100px" }}>
            <label style={{ display: "block", marginBottom: "0.5rem" }}>Puerto:</label>
            <input
              type="text"
              value={formData.port}
              onChange={(e) => setFormData({ ...formData, port: e.target.value })}
              style={{ width: "100%", padding: "0.5rem" }}
              required
            />
          </div>
        </div>

        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.5rem" }}>Usuario:</label>
          <input
            type="text"
            value={formData.user}
            onChange={(e) => setFormData({ ...formData, user: e.target.value })}
            style={{ width: "100%", padding: "0.5rem" }}
            required
          />
        </div>

        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.5rem" }}>Contraseña:</label>
          <input
            type="password"
            value={formData.password}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
            style={{ width: "100%", padding: "0.5rem" }}
            required
          />
        </div>

        <div style={{ marginBottom: "1rem" }}>
          <label style={{ display: "block", marginBottom: "0.5rem" }}>Nombre de BD:</label>
          <input
            type="text"
            value={formData.dbname}
            onChange={(e) => setFormData({ ...formData, dbname: e.target.value })}
            style={{ width: "100%", padding: "0.5rem" }}
            required
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
            cursor: "pointer",
            width: "100%"
          }}
        >
          {loading ? "Guardando..." : "Guardar Configuración"}
        </button>
      </form>
    </div>
  );
}
