import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import './SMTPConfigView.css';

function SMTPConfigView() {
    const [smtpConfig, setSmtpConfig] = useState({
        host: '',
        port: 587,
        username: '',
        password: '',
        from_email: ''
    });
    const [loading, setLoading] = useState(true);
    const [configExists, setConfigExists] = useState(false);
    const [mode, setMode] = useState('view');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [isTesting, setIsTesting] = useState(false);
    const [testResult, setTestResult] = useState(null);

    const navigate = useNavigate();

    // Configurar axios
    axios.defaults.baseURL = 'http://localhost:8080';
    axios.defaults.headers.post['Content-Type'] = 'application/json';

    const loadSMTPConfig = async () => {
        setLoading(true);
        setError('');
        try {
            const response = await axios.get('/admin/smtp-config');
            if (response.data && response.data.host) {
                setSmtpConfig(response.data);
                setConfigExists(true);
            } else {
                setConfigExists(false);
            }
        } catch (error) {
            if (error.response && error.response.status === 404) {
                setConfigExists(false);
            } else {
                setError('Error loading SMTP configuration');
                console.error('Error:', error);
            }
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadSMTPConfig();
    }, []);

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setSmtpConfig(prev => ({
            ...prev,
            [name]: name === 'port' ? parseInt(value) || 587 : value
        }));
    };

    const saveConfig = async () => {
        setLoading(true);
        setError('');
        setSuccess('');
        try {
            const url = '/admin/smtp-config';
            const method = configExists ? 'put' : 'post';
            const response = await axios[method](url, smtpConfig);

            setSuccess(configExists
                ? 'SMTP configuration updated successfully'
                : 'SMTP configuration created successfully');
            setConfigExists(true);
            setMode('view');
            await loadSMTPConfig();
        } catch (error) {
            const errorMsg = error.response?.data?.error ||
                'Error saving configuration';
            setError(errorMsg);
            console.error('Save error:', error);
        } finally {
            setLoading(false);
        }
    };

    const deleteConfig = async () => {
        if (!window.confirm('Are you sure you want to delete the SMTP configuration?')) {
            return;
        }

        setLoading(true);
        setError('');
        try {
            await axios.delete('/admin/smtp-config');
            setSuccess('SMTP configuration deleted successfully');
            setConfigExists(false);
            setSmtpConfig({
                host: '',
                port: 587,
                username: '',
                password: '',
                from_email: ''
            });
            setMode('create');
        } catch (error) {
            setError(error.response?.data?.error || 'Error deleting configuration');
            console.error('Error:', error);
        } finally {
            setLoading(false);
        }
    };

    const testConfig = async () => {
        setIsTesting(true);
        setTestResult(null);
        setError('');

        try {
            console.log('Testing SMTP with config:', smtpConfig);

            // Validación básica en el cliente
            if (!smtpConfig.host || !smtpConfig.port || !smtpConfig.username || !smtpConfig.password) {
                throw new Error('Please fill all required fields');
            }

            const response = await axios.post('/admin/test-smtp', smtpConfig, {
                timeout: 20000 // 20 segundos timeout
            });

            setTestResult({
                success: true,
                message: response.data.message || 'SMTP connection successful',
                details: response.data.details,
                timestamp: new Date().toLocaleTimeString()
            });

        } catch (error) {
            console.error('Full error details:', error);

            let errorDetails = {
                message: 'SMTP connection test failed',
                details: '',
                suggestion: 'Please check your SMTP settings and try again'
            };

            if (error.response) {
                // Error del servidor (4xx/5xx)
                errorDetails = {
                    ...errorDetails,
                    message: error.response.data?.error || 'Server error',
                    details: error.response.data?.details || error.response.statusText,
                    suggestion: error.response.data?.suggestion || errorDetails.suggestion,
                    field: error.response.data?.field // Campo específico con error
                };
            } else if (error.request) {
                // No se recibió respuesta
                errorDetails.details = 'No response from server';
                errorDetails.suggestion = 'Check if the backend service is running';
            } else {
                // Error al configurar la petición
                errorDetails.details = error.message;
            }

            setTestResult({
                success: false,
                ...errorDetails,
                timestamp: new Date().toLocaleTimeString()
            });

        } finally {
            setIsTesting(false);
        }
    };

    if (loading && mode === 'view') {
        return (
            <div className="loading-container">
                <div className="spinner">Loading...</div>
                <p>Loading SMTP configuration...</p>
            </div>
        );
    }

    return (
        <div className="smtp-config-container">
            <div className="card">
                <div className="card-header">
                    <h2>SMTP Configuration</h2>
                </div>

                <div className="card-body">
                    {error && (
                        <div className="alert alert-danger">
                            <strong>{error}</strong>
                        </div>
                    )}

                    {success && (
                        <div className="alert alert-success">
                            <strong>{success}</strong>
                        </div>
                    )}

                    {testResult && (
                        <div className={`alert ${testResult.success ? 'alert-success' : 'alert-danger'} mt-3`}>
                            <div className="d-flex justify-content-between">
                                <strong>{testResult.message}</strong>
                                <small className="text-muted">{testResult.timestamp}</small>
                            </div>

                            {testResult.details && (
                                <div className="mt-2">
                                    <small className="text-muted">Details:</small>
                                    <div className="font-monospace">{testResult.details}</div>
                                </div>
                            )}

                            {testResult.suggestion && !testResult.success && (
                                <div className="mt-2">
                                    <small className="text-muted">Suggestion:</small>
                                    <div>{testResult.suggestion}</div>
                                </div>
                            )}

                            {testResult.field && (
                                <div className="mt-2">
                                    <small className="text-muted">Field with issue:</small>
                                    <div className="text-warning">{testResult.field}</div>
                                </div>
                            )}
                        </div>
                    )}

                    {mode === 'view' && configExists && (
                        <div className="view-mode">
                            <h3>Current Configuration</h3>

                            <div className="config-display">
                                <div className="form-group">
                                    <label>SMTP Server</label>
                                    <input type="text" value={smtpConfig.host} readOnly />
                                </div>
                                <div className="form-group">
                                    <label>Port</label>
                                    <input type="number" value={smtpConfig.port} readOnly />
                                </div>
                                <div className="form-group">
                                    <label>Username</label>
                                    <input type="text" value={smtpConfig.username} readOnly />
                                </div>
                                <div className="form-group">
                                    <label>Sender Email</label>
                                    <input type="text" value={smtpConfig.from_email} readOnly />
                                </div>
                            </div>

                            <div className="action-buttons">
                                <button
                                    className="btn btn-primary"
                                    onClick={() => setMode('edit')}
                                >
                                    Edit
                                </button>

                                <button
                                    className="btn btn-warning"
                                    onClick={testConfig}
                                    disabled={isTesting}
                                >
                                    {isTesting ? 'Testing...' : 'Test SMTP'}
                                </button>

                                <button
                                    className="btn btn-danger"
                                    onClick={deleteConfig}
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    )}

                    {(mode === 'edit' || mode === 'create') && (
                        <div className="edit-mode">
                            <h3>{mode === 'edit' ? 'Edit Configuration' : 'Create SMTP Configuration'}</h3>

                            <div className="config-form">
                                <div className="form-group">
                                    <label>SMTP Server*</label>
                                    <input
                                        type="text"
                                        name="host"
                                        value={smtpConfig.host}
                                        onChange={handleInputChange}
                                        placeholder="smtp.example.com"
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Port*</label>
                                    <input
                                        type="number"
                                        name="port"
                                        value={smtpConfig.port}
                                        onChange={handleInputChange}
                                        placeholder="587"
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Username*</label>
                                    <input
                                        type="text"
                                        name="username"
                                        value={smtpConfig.username}
                                        onChange={handleInputChange}
                                        placeholder="user@example.com"
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Password*</label>
                                    <input
                                        type="password"
                                        name="password"
                                        value={smtpConfig.password}
                                        onChange={handleInputChange}
                                        placeholder="••••••••"
                                        required
                                    />
                                </div>

                                <div className="form-group">
                                    <label>Sender Email*</label>
                                    <input
                                        type="email"
                                        name="from_email"
                                        value={smtpConfig.from_email}
                                        onChange={handleInputChange}
                                        placeholder="no-reply@example.com"
                                        required
                                    />
                                </div>
                            </div>

                            <div className="form-actions">
                                <button
                                    className="btn btn-success"
                                    onClick={saveConfig}
                                    disabled={loading}
                                >
                                    {loading ? 'Saving...' : 'Save'}
                                </button>

                                <button
                                    className="btn btn-secondary"
                                    onClick={() => {
                                        if (configExists) {
                                            setMode('view');
                                            loadSMTPConfig();
                                        } else {
                                            setSmtpConfig({
                                                host: '',
                                                port: 587,
                                                username: '',
                                                password: '',
                                                from_email: ''
                                            });
                                        }
                                    }}
                                >
                                    Cancel
                                </button>

                                {mode === 'edit' && (
                                    <button
                                        className="btn btn-warning"
                                        onClick={testConfig}
                                        disabled={isTesting}
                                    >
                                        {isTesting ? 'Testing...' : 'Test SMTP'}
                                    </button>
                                )}
                            </div>
                        </div>
                    )}

                    {!configExists && mode === 'view' && (
                        <div className="no-config">
                            <p>No SMTP configuration registered</p>
                            <button
                                className="btn btn-primary"
                                onClick={() => setMode('create')}
                            >
                                Create SMTP Configuration
                            </button>
                        </div>
                    )}
                </div>

                <div className="card-footer">
                    <button
                        className="btn btn-outline-secondary"
                        onClick={() => navigate('/')}
                    >
                        Back to Main Menu
                    </button>
                </div>
            </div>
        </div>
    );
}

export default SMTPConfigView;