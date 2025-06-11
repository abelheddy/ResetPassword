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

export async function saveDBConfig(config) {
  const response = await fetch(`${API_BASE_URL}/api/setup-db`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(config),
  });

  if (!response.ok) {
    let errorData = {};
    try {
      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        errorData = await response.json();
      } else {
        const text = await response.text();
        console.error("Respuesta no JSON:", text);
        throw new Error(`Error inesperado del servidor: ${text.substring(0, 100)}...`);
      }
    } catch (parseError) {
      console.error("Error al interpretar respuesta del servidor", parseError);
    }

    throw new Error(errorData.message || "Error guardando configuración");
  }

  return await response.json();
}
