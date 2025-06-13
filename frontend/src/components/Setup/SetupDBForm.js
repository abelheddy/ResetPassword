import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { saveDBConfig, testDBConnection } from "../../services/api"; // Nueva función

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
  const [testResult, setTestResult] = useState(null);

  // Función para probar conexión
  const handleTestConnection = async () => {
    setError("");
    setTestResult(null);

    try {
      const result = await testDBConnection(formData);
      setTestResult({
        success: true,
        message: "Conexión exitosa!",
        details: result
      });
    } catch (err) {
      setTestResult({
        success: false,
        message: err.message
      });
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");

    // Requerir prueba de conexión exitosa primero
    if (!testResult || !testResult.success) {
      setError("Debes probar la conexión exitosamente antes de guardar");
      return;
    }

    setLoading(true);

    try {
      await saveDBConfig(formData);
      navigate("/dashboard");
    } catch (err) {
      setError(err.message || "Error al guardar la configuración");
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

        <div style={{ marginTop: "1rem", display: "flex", gap: "1rem" }}>
          <button
            type="button"
            onClick={handleTestConnection}
            style={{
              padding: "0.5rem 1rem",
              background: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              flex: 1
            }}
          >
            Probar Conexión
          </button>

          <button
            type="submit"
            onClick={handleSubmit}
            disabled={loading || !testResult?.success}
            style={{
              padding: "0.5rem 1rem",
              background: loading || !testResult?.success ? "#ccc" : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: testResult?.success ? "pointer" : "not-allowed",
              flex: 1
            }}
          >
            {loading ? "Guardando..." : "Guardar Configuración"}
          </button>
        </div>
      </form>

      {testResult && (
        <div style={{
          marginTop: "1rem",
          padding: "1rem",
          borderLeft: `4px solid ${testResult.success ? "#28a745" : "#dc3545"}`,
          backgroundColor: "#f8f9fa"
        }}>
          <p style={{ fontWeight: "bold", marginBottom: "0.5rem" }}>
            {testResult.success ? "✓ Conexión exitosa" : "✗ Error de conexión"}
          </p>
          <p>{testResult.message}</p>
          {testResult.details && (
            <pre style={{ marginTop: "0.5rem", fontSize: "0.8rem" }}>
              {JSON.stringify(testResult.details, null, 2)}
            </pre>
          )}
        </div>
      )}
    </div>
  );
}