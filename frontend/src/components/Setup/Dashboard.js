import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { API_BASE_URL } from "../../services/api";
import "./Dashboard.css";

export default function Dashboard() {
    // Estado para el estado del sistema
    const [systemStatus, setSystemStatus] = useState({
        dbConnected: false,
        dbType: "Desconocido",
        dbUser: "N/A",
        dbName: "N/A"
    });
    //Nuevo estado para controlar si el usuario ya existe:
    const [adminExists, setAdminExists] = useState(false);
    // Información de la base de datos
    const [dbInfo, setDbInfo] = useState({
        host: "N/A",
        port: "N/A",
        dbname: "N/A",
        user: "N/A"
    });
    // Estado de las etapas de configuración
    const [setupStages, setSetupStages] = useState({
        dbConfigured: false,
        tablesCreated: false,
        adminCreated: false
    });
    const [allowReconfigure, setAllowReconfigure] = useState(false);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [dbTestResult, setDbTestResult] = useState(null);
    const [isTestingDb, setIsTestingDb] = useState(false);
    const [isCreatingTables, setIsCreatingTables] = useState(false);
    const [isCreatingAdmin, setIsCreatingAdmin] = useState(false);
    const [adminForm, setAdminForm] = useState({
        email: "",
        password: "",
        confirmPassword: ""
    });
    const [formErrors, setFormErrors] = useState({});
    const [successMessage, setSuccessMessage] = useState("");
    const navigate = useNavigate();

    useEffect(() => {
        // Cargar estado guardado de localStorage
        const savedSetupStages = localStorage.getItem('setupStages');
        if (savedSetupStages) {
            setSetupStages(JSON.parse(savedSetupStages));
        }
        checkSystemStatus();
    }, []);

    useEffect(() => {
        // Cargar estado guardado de localStorage
        const savedSetupStages = localStorage.getItem('setupStages');
        if (savedSetupStages) {
            setSetupStages(JSON.parse(savedSetupStages));
        }
        checkSystemStatus();
    }, []);

    const checkSystemStatus = async () => {
        try {
            setLoading(true);
            setError(null);

            const response = await fetch(`${API_BASE_URL}/api/status`);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            // Actualizar toda la información de estado
            setSetupStages({
                dbConfigured: data.setup_stages.db_configured,
                tablesCreated: data.setup_stages.tables_created,
                adminCreated: data.setup_stages.admin_created
            });

            setAllowReconfigure(data.setup_stages.allow_reconfigure || false);

            // Verificar si el admin ya existe
            if (data.setup_stages.admin_created) {
                setAdminExists(true);
            }

            // Actualizar información de conexión
            if (data.db_info) {
                setDbInfo(data.db_info);
            }

            // Actualizar estado de conexión
            setSystemStatus(prev => ({
                ...prev,
                dbConnected: data.setup_stages.db_configured,
                dbType: "PostgreSQL",
                dbUser: data.db_info?.user || "N/A",
                dbName: data.db_info?.dbname || "N/A"
            }));
        } catch (err) {
            console.error("System status check failed:", err);
            setError(err.message || "Error al verificar el estado del sistema");
        } finally {
            setLoading(false);
        }
    };


    const testDbConnection = async () => {
        try {
            setIsTestingDb(true);
            setDbTestResult(null);
            setError(null);
            setSuccessMessage("");

            const response = await fetch(`${API_BASE_URL}/api/db/test`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.message || "Error al probar la conexión");
            }

            // Actualizar estado con la información de la conexión
            setSystemStatus(prev => ({
                ...prev,
                dbConnected: true,
                dbType: result.result.db_type || "PostgreSQL",
                dbUser: result.result.user || "N/A",
                dbName: result.result.dbname || "N/A"
            }));

            // Marcar la etapa de configuración de DB como completada
            setSetupStages(prev => ({
                ...prev,
                dbConfigured: true
            }));

            setDbTestResult({
                success: true,
                message: "Conexión exitosa a la base de datos",
                details: result
            });
            setSuccessMessage("Conexión a la base de datos configurada correctamente");
        } catch (err) {
            console.error("DB test failed:", err);
            setSystemStatus(prev => ({
                ...prev,
                dbConnected: false,
                dbType: "Error",
                dbUser: "N/A"
            }));
            setDbTestResult({
                success: false,
                message: err.message || "Error al conectar con la base de datos"
            });
        } finally {
            setIsTestingDb(false);
        }
    };

    const createTables = async () => {
        try {
            setIsCreatingTables(true);
            setError(null);
            setSuccessMessage("");

            const response = await fetch(`${API_BASE_URL}/api/setup/create-tables`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.message || "Error al crear tablas");
            }

            setSetupStages(prev => ({
                ...prev,
                tablesCreated: true
            }));
            setSuccessMessage("Estructura de la base de datos creada correctamente");
        } catch (err) {
            console.error("Create tables failed:", err);
            setError(err.message || "Error al crear las tablas");
        } finally {
            setIsCreatingTables(false);
        }
    };

    const validateAdminForm = () => {
        const errors = {};
        if (!adminForm.email) errors.email = "Email es requerido";
        if (!adminForm.password) errors.password = "Contraseña es requerida";
        if (adminForm.password.length < 6) errors.password = "La contraseña debe tener al menos 6 caracteres";
        if (adminForm.password !== adminForm.confirmPassword) errors.confirmPassword = "Las contraseñas no coinciden";
        return errors;
    };

    const handleAdminSubmit = async (e) => {
        e.preventDefault();

        // Si ya existe admin, no hacer nada
        if (adminExists) {
            setSuccessMessage("✅ Usuario administrador ya existe - configuración completada");
            return;
        }

        const errors = validateAdminForm();
        setFormErrors(errors);

        if (Object.keys(errors).length > 0) {
            return;
        }

        try {
            setIsCreatingAdmin(true);
            setError(null);
            setSuccessMessage("");

            const response = await fetch(`${API_BASE_URL}/api/setup/create-admin`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    email: adminForm.email,
                    password: adminForm.password
                }),
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.message || "Error al crear usuario administrador");
            }

            // Si el backend indica que el usuario ya existía
            if (result.already_exists) {
                setAdminExists(true);
                setSuccessMessage("✅ Usuario administrador ya existe - configuración completada");
            } else {
                setSuccessMessage("✅ Usuario administrador creado correctamente");
            }

            setSetupStages(prev => ({
                ...prev,
                adminCreated: true
            }));
        } catch (err) {
            console.error("Create admin failed:", err);
            setError(err.message || "Error al crear el usuario administrador");
        } finally {
            setIsCreatingAdmin(false);
        }
    };

    const resetStage = async (stage) => {
        try {
            setLoading(true);
            setError(null);

            if (stage === 'db') {
                // Para DB, usamos el reset completo que ya existe
                await resetConfiguration();
            } else {
                // Para tablas/admin, solo actualizamos el estado local
                if (stage === 'tables') {
                    setSetupStages(prev => ({ ...prev, tablesCreated: false }));
                } else if (stage === 'admin') {
                    setSetupStages(prev => ({ ...prev, adminCreated: false }));
                }
                setSuccessMessage(`Etapa "${stage}" reseteada. Puedes reconfigurarla.`);
            }
        } catch (err) {
            setError(`Error al resetear ${stage}: ${err.message}`);
        } finally {
            setLoading(false);
        }
    };
    //reseteo de la configuración completa
    const resetConfiguration = async () => {
        try {
            setLoading(true);
            setError("");

            // 1. Ejecutar reset en el backend
            const response = await fetch(`${API_BASE_URL}/api/setup/reset`, {
                method: 'POST'
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || "Error al resetear");
            }

            const result = await response.json();

            // 2. Verificar que el archivo fue eliminado
            if (result.file_exists) {
                throw new Error("El archivo de configuración no fue eliminado");
            }

            // 3. Actualizar estado local
            setSystemStatus({
                dbConnected: false,
                dbType: "Desconocido",
                dbUser: "N/A",
                dbName: "N/A"
            });

            setSetupStages({
                dbConfigured: false,
                tablesCreated: false,
                adminCreated: false
            });

            setSuccessMessage(result.message);

            // 4. Forzar recarga completa después de 2 segundos
            setTimeout(() => {
                window.location.reload();
            }, 2000);

        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="dashboard-loading">
                <div className="spinner"></div>
                <p>Verificando estado del sistema...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="dashboard-error">
                <div className="error-icon">⚠️</div>
                <h2>Error al cargar el panel</h2>
                <p>{error}</p>
                <button
                    className="btn-retry"
                    onClick={checkSystemStatus}
                >
                    Reintentar
                </button>
            </div>
        );
    }

    return (
        <div className="dashboard-container">
            <header className="dashboard-header">
                <h1>Panel de Configuración</h1>
                <ProgressIndicator stages={setupStages} />
            </header>

            <main className="dashboard-main">
                {successMessage && (
                    <div className="success-message">
                        <span>✓</span> {successMessage}
                    </div>
                )}

                {error && (
                    <div className="error-message">
                        <span>⚠️</span> {error}
                    </div>
                )}

                <section className="status-section">
                    <h2>Estado del Sistema</h2>
                    <div className="status-cards">
                        <StatusCard
                            title="Estado de Conexión"
                            status={systemStatus.dbConnected ? "active" : "error"}
                            value={systemStatus.dbConnected ? "Conectado" : "Desconectado"}
                        />
                        <StatusCard
                            title="Base de Datos"
                            status="neutral"
                            value={dbInfo.dbname}
                        />
                        <StatusCard
                            title="Host"
                            status="neutral"
                            value={dbInfo.host}
                        />
                        <StatusCard
                            title="Usuario DB"
                            status="neutral"
                            value={dbInfo.user}
                        />
                    </div>
                </section>

                <section className="setup-stages">
                    {/* Etapa 1: Configuración DB */}
                    <div className={`setup-stage ${setupStages.dbConfigured ? 'completed' : 'current'}`}>
                        <div className="stage-header">
                            <span className="stage-number">1</span>
                            <h3>Configurar conexión a base de datos</h3>
                            {setupStages.dbConfigured && (
                                <button
                                    className="btn-reset"
                                    onClick={() => resetStage('db')}
                                    title="Resetear esta etapa"
                                >
                                    ↻
                                </button>
                            )}
                        </div>

                        {!setupStages.dbConfigured ? (
                            <div className="stage-content">
                                <p>Configura los parámetros de conexión a la base de datos</p>
                                <div className="action-buttons">
                                    <button
                                        className="btn-primary"
                                        onClick={() => navigate("/db-config")}
                                    >
                                        Configurar DB
                                    </button>
                                    {/*<button
                                        className="btn-secondary"
                                        onClick={testDbConnection}
                                        disabled={isTestingDb}
                                    >
                                        {isTestingDb ? "Probando conexión..." : "Probar Conexión"}
                                    </button>*/}
                                </div>
                                {dbTestResult && (
                                    <div className={`db-test-result ${dbTestResult.success ? 'success' : 'error'}`}>
                                        <p>{dbTestResult.message}</p>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="stage-completed">
                                <span>✓ Configuración completada</span>
                            </div>
                        )}
                    </div>

                    {/* Etapa 2: Crear tablas */}
                    {setupStages.dbConfigured && (
                        <div className={`setup-stage ${setupStages.tablesCreated ? 'completed' : 'current'}`}>
                            <div className="stage-header">
                                <span className="stage-number">2</span>
                                <h3>Crear estructura de base de datos</h3>
                                {setupStages.tablesCreated && (
                                    <button
                                        className="btn-reset"
                                        onClick={() => resetStage('tables')}
                                        title="Resetear esta etapa"
                                    >
                                        ↻
                                    </button>
                                )}
                            </div>

                            {!setupStages.tablesCreated ? (
                                <div className="stage-content">
                                    <p>Se crearán las tablas necesarias para la aplicación</p>
                                    <div className="action-buttons">
                                        <button
                                            className="btn-primary"
                                            onClick={createTables}
                                            disabled={isCreatingTables}
                                        >
                                            {isCreatingTables ? "Creando tablas..." : "Crear Tablas"}
                                        </button>
                                    </div>
                                </div>
                            ) : (
                                <div className="stage-completed">
                                    <span>✓ Estructura creada</span>
                                </div>
                            )}
                        </div>
                    )}

                    {/* Etapa 3: Crear admin */}
                    {setupStages.tablesCreated && (
                        <div className={`setup-stage ${setupStages.adminCreated ? 'completed' : 'current'}`}>
                            <div className="stage-header">
                                <span className="stage-number">3</span>
                                <h3>Crear usuario administrador</h3>
                                {setupStages.adminCreated && (
                                    <button
                                        className="btn-reset"
                                        onClick={() => resetStage('admin')}
                                        title="Resetear esta etapa"
                                    >
                                        ↻
                                    </button>
                                )}
                            </div>

                            {!setupStages.adminCreated ? (
                                <div className="stage-content">
                                    <form onSubmit={handleAdminSubmit} className="admin-form">
                                        <div className="form-group">
                                            <label>Email:</label>
                                            <input
                                                type="email"
                                                value={adminForm.email}
                                                onChange={(e) => !adminExists && setAdminForm({ ...adminForm, email: e.target.value })}
                                                placeholder="admin@example.com"
                                                disabled={adminExists}
                                            />
                                            {formErrors.email && <span className="error">{formErrors.email}</span>}
                                        </div>
                                        <div className="form-group">
                                            <label>Contraseña:</label>
                                            <input
                                                type="password"
                                                value={adminForm.password}
                                                onChange={(e) => !adminExists && setAdminForm({ ...adminForm, password: e.target.value })}
                                                placeholder="Al menos 6 caracteres"
                                                disabled={adminExists}
                                            />
                                            {formErrors.password && <span className="error">{formErrors.password}</span>}
                                        </div>
                                        <div className="form-group">
                                            <label>Confirmar Contraseña:</label>
                                            <input
                                                type="password"
                                                value={adminForm.confirmPassword}
                                                onChange={(e) => !adminExists && setAdminForm({ ...adminForm, confirmPassword: e.target.value })}
                                                placeholder="Repite la contraseña"
                                                disabled={adminExists}
                                            />
                                            {formErrors.confirmPassword && (
                                                <span className="error">{formErrors.confirmPassword}</span>
                                            )}
                                        </div>
                                        <button
                                            type="submit"
                                            className="btn-primary"
                                            disabled={isCreatingAdmin || adminExists}
                                        >
                                            {adminExists ? "Usuario ya existe" : isCreatingAdmin ? "Creando administrador..." : "Crear Administrador"}
                                        </button>

                                        {adminExists && (
                                            <div className="admin-exists-note">
                                                <p>✔️ Usuario administrador ya existe en la base de datos</p>
                                                <p>Puedes continuar con la configuración</p>
                                            </div>
                                        )}
                                    </form>
                                </div>
                            ) : (
                                <div className="stage-completed">
                                    <span>✓ Administrador creado</span>
                                </div>
                            )}
                        </div>
                    )}
                </section>

                {/* Botón de finalización */}
                {setupStages.adminCreated && (
                    <section className="finalization-section">
                        <div className="action-buttons">
                            <button
                                className="btn-success"
                                onClick={() => {
                                    // Aquí podrías limpiar el localStorage si es necesario
                                    localStorage.removeItem('setupStages');
                                    navigate("/");
                                }}
                            >
                                Cerrar Modo Configuración
                            </button>
                            <button
                                className="btn-warning"
                                onClick={resetConfiguration}
                            >
                                Resetear Configuración Completa
                            </button>
                        </div>
                    </section>
                )}
            </main>
        </div>
    );
}

// Componente ProgressIndicator
function ProgressIndicator({ stages }) {
    const totalStages = 3;
    const completedStages = [
        stages.dbConfigured,
        stages.tablesCreated,
        stages.adminCreated
    ].filter(Boolean).length;

    return (
        <div className="progress-indicator">
            <div className="progress-bar">
                <div
                    className="progress-fill"
                    style={{ width: `${(completedStages / totalStages) * 100}%` }}
                ></div>
            </div>
            <div className="progress-text">
                Progreso: {completedStages} de {totalStages} etapas completadas
            </div>
        </div>
    );
}

// Componente StatusCard
function StatusCard({ title, status, value }) {
    return (
        <div className={`status-card ${status}`}>
            <h3>{title}</h3>
            <p>{value}</p>
        </div>
    );
}