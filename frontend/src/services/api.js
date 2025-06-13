// src/services/api.js
export const API_BASE_URL = "http://localhost:8080"; // Asegúrate que coincida con tu puerto backend

export async function checkSetupStatus() {
  const response = await fetch(`${API_BASE_URL}/api/status`);
  if (!response.ok) {
    throw new Error("Error verificando estado del sistema");
  }
  return await response.json();
}


export async function loginSetup(user, pass) {
  const response = await fetch(`${API_BASE_URL}/api/login-setup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ user, pass }),
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || "Credenciales inválidas");
  }

  return await response.json();
}

// Función para guardar configuración
export async function saveDBConfig(config) {
  const response = await fetch(`${API_BASE_URL}/api/setup-db`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(config),
  });

  if (!response.ok) {
    let errorData = {};
    try {
      errorData = await response.json();
    } catch (e) {
      errorData = { error: "Error desconocido" };
    }
    throw new Error(errorData.error || "Error guardando configuración");
  }

  return await response.json();
}
//resetean la configuración del sistema
export async function resetConfiguration() {
  const response = await fetch(`${API_BASE_URL}/api/setup/reset`, {
    method: 'POST'
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || "Error al resetear la configuración");
  }

  // Forzar recarga del estado en el backend
  await fetch(`${API_BASE_URL}/api/status`, {
    method: 'GET',
    cache: 'no-cache'
  });

  return await response.json();
}

// Función para probar conexión con configuración temporal
export async function testDBConnection(config) {
  const response = await fetch(`${API_BASE_URL}/api/db/test-config`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(config),
  });

  const result = await response.json();

  if (!response.ok) {
    throw new Error(result.error || "Error al probar la conexión");
  }

  return result;
}
