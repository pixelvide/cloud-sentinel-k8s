import { API_URL } from "./config";
import { getSubPath, withSubPath } from "./subpath";

interface ApiFetchOptions extends RequestInit {
    skipRedirect401?: boolean;
}

export async function apiFetch(path: string, options: ApiFetchOptions = {}) {
    const subPath = getSubPath(); // e.g. /my-app or empty string
    const apiV1Base = subPath ? `${subPath}${API_URL}` : API_URL; // e.g. /my-app/api/v1 or /api/v1

    // If it's a relative path, prepend API_URL
    let url: string;
    if (path.startsWith("http")) {
        url = path;
    } else if (path.startsWith("/api/auth")) {
        // Special handling for auth routes which are now at /api/auth (not /api/v1/auth)
        const baseUrl = subPath ? subPath : "";
        url = `${baseUrl}${path}`;
    } else if (path.startsWith("/") && path.startsWith(apiV1Base)) {
        url = path;
    } else {
        url = `${apiV1Base}${path.startsWith("/") ? "" : "/"}${path}`;
    }

    const { skipRedirect401, ...fetchOptions } = options;

    const defaultOptions: RequestInit = {
        credentials: "include",
        ...fetchOptions,
        headers: {
            "Content-Type": "application/json",
            ...fetchOptions.headers,
        },
    };

    if (fetchOptions.body instanceof FormData) {
        // Let the browser set the Content-Type with the boundary
        delete (defaultOptions.headers as any)["Content-Type"];
    }

    const response = await fetch(url, defaultOptions);

    if (response.status === 401 && !skipRedirect401) {
        if (typeof window !== "undefined") {
            window.location.href = withSubPath("/login");
        }
    }

    return response;
}

export async function get<T>(path: string, options: ApiFetchOptions = {}): Promise<T> {
    const response = await apiFetch(path, { ...options, method: "GET" });
    if (!response.ok) {
        let errorMessage = `API Error: ${response.statusText}`;
        try {
            const errorData = await response.json();
            if (errorData && errorData.error) {
                errorMessage = errorData.error;
            }
        } catch (e) {
            // Ignore JSON parse error, use default message
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

export async function post<T>(path: string, body?: any, options: ApiFetchOptions = {}): Promise<T> {
    const response = await apiFetch(path, {
        ...options,
        method: "POST",
        body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
        let errorMessage = `API Error: ${response.statusText}`;
        try {
            const errorData = await response.json();
            if (errorData && errorData.error) {
                errorMessage = errorData.error;
            }
        } catch (e) {
            // Ignore JSON parse error, use default message
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

/**
 * Constructs a WebSocket URL for a given path, using the API_URL as a base if it's external.
 * @param path The API path (e.g., "/kube/logs")
 * @returns The full WebSocket URL
 */
export function getWsUrl(path: string): string {
    const subPath = getSubPath();
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    let wsHost = window.location.host;

    // If API_URL is an absolute URL, extract the host
    if (API_URL.startsWith("http")) {
        try {
            const url = new URL(API_URL);
            wsHost = url.host;
        } catch (e) {
            console.error("Invalid API_URL for WebSocket construction:", API_URL);
        }
    }

    const cleanPath = path.startsWith("/") ? path : `/${path}`;
    const fullPath = subPath ? `${subPath}${cleanPath}` : cleanPath;
    return `${protocol}//${wsHost}${fullPath}`;
}

export async function put<T>(path: string, body?: any, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, {
        ...options,
        method: "PUT",
        body: body ? JSON.stringify(body) : undefined,
    });
    if (!response.ok) {
        let errorMessage = `API Error: ${response.statusText}`;
        try {
            const errorData = await response.json();
            if (errorData && errorData.error) {
                errorMessage = errorData.error;
            }
        } catch (e) {
            // Ignore JSON parse error, use default message
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

export async function del<T>(path: string, options: RequestInit = {}): Promise<T> {
    const response = await apiFetch(path, { ...options, method: "DELETE" });
    if (!response.ok) {
        let errorMessage = `API Error: ${response.statusText}`;
        try {
            const errorData = await response.json();
            if (errorData && errorData.error) {
                errorMessage = errorData.error;
            }
        } catch (e) {
            // Ignore JSON parse error, use default message
        }
        throw new Error(errorMessage);
    }
    return response.json();
}

export const api = {
    get,
    post,
    put,
    del,
    fetch: apiFetch,
    getWsUrl,
    getPodMetrics: async (context: string, namespace: string, podName: string) => {
        return get<{
            cpu: Array<{ timestamp: string; value: number }>;
            memory: Array<{ timestamp: string; value: number }>;
            fallback: boolean;
        }>(`/kube/metrics/pods?namespace=${encodeURIComponent(namespace)}&podName=${encodeURIComponent(podName)}`, {
            headers: {
                "x-kube-context": context,
            },
        });
    },
    checkInit: async () => {
        return get<InitCheckResponse>("/init_check");
    },
    createSuperUser: async (data: CreateSuperUserRequest) => {
        return post("/create_superuser", data);
    },
    skipOIDC: async () => {
        return post("/skip_oidc");
    },
    getProviders: async () => {
        return get<string[]>("/api/auth/providers"); // Explicit api/auth path
    },
    loginWithPassword: async (data: any) => {
        return post("/api/auth/login/password", data, { skipRedirect401: true }); // Explicit api/auth path
    },
};

export interface InitCheckResponse {
    initialized: boolean;
    step: number;
}

export interface CreateSuperUserRequest {
    email: string;
    password: string;
    name?: string;
}
