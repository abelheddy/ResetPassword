import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { API_BASE_URL } from "../../services/api";

export default function Dashboard() {
    const [systemStatus, setSystemStatus] = useState(null);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        const checkSystemStatus = async () => {
            try {
                const response = await fetch(`${API_BASE_URL}/api/status`);
                const text = await response.text();

                // Intentamos parsear JSON
                let data;
                try {
                    data = JSON.parse(text);
                } catch (err) {
                    throw new Error(`Respuesta inválida del servidor: no es JSON:\n${text}`);
                }

                if (!response.ok) throw new Error(data.message || "Error verificando estado");

                if (!data.setup) {
                    navigate("/login-setup");
                } else {
                    setSystemStatus(data);
                }
            } catch (err) {
                console.error("Error en checkSystemStatus:", err);
                navigate("/login-setup");
            } finally {
                setLoading(false);
            }
        };

        checkSystemStatus();
    }, [navigate]);

    if (loading) {
        return <div className="loading">Cargando...</div>;
    }

    return (
        <div className="dashboard">
            <header className="dashboard-header">
                <h1>Panel de Administración</h1>
                <p>Sistema configurado correctamente</p>
            </header>

            <div className="dashboard-content">
                <div className="status-card">
                    <h3>Estado del Sistema</h3>
                    <p>Base de datos: <span className="status-active">Conectada</span></p>
                    <p>Modo: <strong>Producción</strong></p>
                </div>

                <div className="actions">
                    <button className="btn-primary">Gestionar Usuarios</button>
                    <button className="btn-secondary">Ver Estadísticas</button>
                </div>
            </div>
        </div>
    );
}
