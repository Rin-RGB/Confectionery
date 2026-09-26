const API_BASE = (import.meta.env.VITE_API_URL || "/api/v1").replace(/\/$/, "");

const ACCESS_KEY = "caprice_access_token";
const REFRESH_KEY = "caprice_refresh_token";

export const authStorage = {
  getAccess() {
    return localStorage.getItem(ACCESS_KEY);
  },
  getRefresh() {
    return localStorage.getItem(REFRESH_KEY);
  },
  set(tokens) {
    localStorage.setItem(ACCESS_KEY, tokens.access_token);
    localStorage.setItem(REFRESH_KEY, tokens.refresh_token);
  },
  clear() {
    localStorage.removeItem(ACCESS_KEY);
    localStorage.removeItem(REFRESH_KEY);
  }
};

async function parseResponse(response) {
  const text = await response.text();
  let body = null;

  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = { message: text };
    }
  }

  if (!response.ok) {
    const message =
      body?.error?.message ||
      body?.message ||
      `Ошибка API (${response.status})`;

    const error = new Error(message);
    error.status = response.status;
    error.payload = body;
    throw error;
  }

  return body;
}

async function rawRequest(path, options = {}, retry = true) {
  const headers = new Headers(options.headers || {});
  if (options.body !== undefined && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const access = authStorage.getAccess();
  if (access) {
    headers.set("Authorization", `Bearer ${access}`);
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers
  });

  if (response.status === 401 && retry) {
    const refresh = authStorage.getRefresh();

    if (refresh) {
      try {
        const tokens = await rawRequest(
          "/refresh",
          {
            method: "POST",
            body: JSON.stringify({ refresh_token: refresh })
          },
          false
        );

        authStorage.set(tokens);

        return rawRequest(path, options, false);
      } catch {
        authStorage.clear();
      }
    }
  }

  return parseResponse(response);
}

export const api = {
  register(email, password) {
    return rawRequest("/register", {
      method: "POST",
      body: JSON.stringify({ email, password })
    });
  },

  login(email, password) {
    return rawRequest("/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    });
  },

  refresh() {
    const refreshToken = authStorage.getRefresh();
    if (!refreshToken) return Promise.reject(new Error("Нет refresh token"));

    return rawRequest("/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken })
    });
  },

  logout() {
    const refreshToken = authStorage.getRefresh();
    authStorage.clear();

    if (!refreshToken) return Promise.resolve();

    return rawRequest(
      "/logout",
      {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken })
      },
      false
    ).catch(() => undefined);
  },

  getFillings() {
    return rawRequest("/fillings", { method: "GET" });
  },

  getFilling(id) {
    return rawRequest(`/fillings/${encodeURIComponent(id)}`, { method: "GET" });
  },

  createFilling(payload) {
    return rawRequest("/fillings", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },

  updateFilling(id, payload) {
    return rawRequest(`/fillings/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify(payload)
    });
  },

  deleteFilling(id) {
    return rawRequest(`/fillings/${encodeURIComponent(id)}`, {
      method: "DELETE"
    });
  },

  calculatePrice(payload) {
    return rawRequest("/orders/price", {
      method: "POST",
      body: JSON.stringify(payload)
    });
  },

  createOrder(payload) {
    return rawRequest("/orders", {
      method: "POST",
      headers: {
        "Idempotency-Key": crypto.randomUUID()
      },
      body: JSON.stringify(payload)
    });
  },

  getMyOrders() {
    return rawRequest("/orders/my", { method: "GET" });
  },

  getOrder(id) {
    return rawRequest(`/orders/${encodeURIComponent(id)}`, { method: "GET" });
  },

  getAllOrders() {
    return rawRequest("/orders", { method: "GET" });
  },

  updateOrderStatus(id, status) {
    return rawRequest(`/orders/${encodeURIComponent(id)}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status })
    });
  }
};
